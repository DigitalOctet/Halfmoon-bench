#!/bin/bash

set -u

BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/../../..`

BENCH_IMAGE=emptyredbox/halfmoon-bench:test-v0

# Do I need them?
AWS_REGION=ap-southeast-1
NUM_KEYS=100
EXP_DIR=$BASE_DIR/results/QPS15  # $1=QPS15
QPS=15                           # $2=15

WRK_DIR=/usr/local/bin

TABLE_PREFIX=$(head -c 64 /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 8 | head -n 1)
TABLE_PREFIX="${TABLE_PREFIX}-"

minikube start --force

kubectl label nodes minikube node-restriction.kubernetes.io/placement_label=engine_node

minikube ssh -- docker pull $BENCH_IMAGE

minikube ssh -- docker run -v /home/docker:/tmp \
    $BENCH_IMAGE \
    cp -r /bokiflow-bin/singleop /tmp/

minikube ssh -- TABLE_PREFIX=$TABLE_PREFIX AWS_REGION=$AWS_REGION NUM_KEYS=$NUM_KEYS \
    /home/docker/singleop/init create
minikube ssh -- TABLE_PREFIX=$TABLE_PREFIX AWS_REGION=$AWS_REGION NUM_KEYS=$NUM_KEYS \
    /home/docker/singleop/init populate

# move local zk_setup.sh to minikube node
minikube cp $ROOT_DIR/scripts/zk_setup.sh minikube:/tmp/zk_setup.sh
# change execution permission bit
minikube ssh -- sudo chmod +x /tmp/zk_setup.sh
minikube ssh -- sudo mkdir -p /mnt/inmem/store

# to all hosts
minikube cp $BASE_DIR/nightcore_config.json minikube:/tmp/nightcore_config.json
# to engine hosts
minikube cp $BASE_DIR/run_launcher minikube:/tmp/run_launcher
minikube ssh -- sudo chmod +x /tmp/run_launcher
minikube ssh -- sudo rm -rf /mnt/inmem/boki
minikube ssh -- sudo mkdir -p /mnt/inmem/boki
minikube ssh -- sudo mkdir -p /mnt/inmem/boki/output /mnt/inmem/boki/ipc
minikube ssh -- sudo cp /tmp/run_launcher /mnt/inmem/boki/run_launcher
minikube ssh -- sudo cp /tmp/nightcore_config.json /mnt/inmem/boki/func_config.json
# to storage hosts
minikube ssh -- sudo rm -rf   /mnt/storage/logdata
minikube ssh -- sudo mkdir -p /mnt/storage/logdata

sleep 10
# start zookeeper
kubectl apply -f "$BASE_DIR/k8s_files/zookeeper.yaml"
sleep 10
# set up zookeeper
kubectl apply -f "$BASE_DIR/k8s_files/zookeeper-setup.yaml"
sleep 10
kubectl exec -it zookeeper-setup -- bash /tmp/boki/zk_setup.sh &
kubectl apply -f "$BASE_DIR/k8s_files/boki-engine.yaml"
kubectl apply -f "$BASE_DIR/k8s_files/boki-gateway.yaml"
kubectl apply -f "$BASE_DIR/k8s_files/boki-storage.yaml"
kubectl apply -f "$BASE_DIR/k8s_files/boki-sequencer.yaml"
sleep 60
kubectl apply -f "$BASE_DIR/k8s_files/boki-controller.yaml"
kubectl apply -f "$BASE_DIR/k8s_files/app.yaml"
sleep 120

rm -rf $EXP_DIR
mkdir -p $EXP_DIR

minikube ssh -- cat /proc/cmdline >>$EXP_DIR/kernel_cmdline
minikube ssh -- uname -a >>$EXP_DIR/kernel_version
minikube cp ~/wrk2/wrk minikube:/usr/local/bin/
minikube ssh -- sudo chmod +x /usr/local/bin/wrk

minikube cp $ROOT_DIR/workloads/workflow/boki/benchmark/singleop/workload.lua minikube:/tmp/
minikube ssh -- $WRK_DIR/wrk -t 2 -c 2 -d 40 -L -U \
    -s /tmp/workload.lua \
    http://10.244.0.6:8080 -R $QPS >$EXP_DIR/wrk_warmup.log
sleep 10
minikube ssh -- $WRK_DIR/wrk -t 2 -c 2 -d 200 -L -U \
    -s /tmp/workload.lua \
    http://10.244.0.6:8080 -R $QPS 2>/dev/null >$EXP_DIR/wrk.log
sleep 10

minikube ssh -- TABLE_PREFIX=$TABLE_PREFIX AWS_REGION=$AWS_REGION NUM_KEYS=$NUM_KEYS \
    /home/docker/singleop/init clean