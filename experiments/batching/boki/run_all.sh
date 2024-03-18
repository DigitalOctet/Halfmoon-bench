#!/bin/bash
BASE_DIR=`realpath $(dirname $0)`
ROOT_DIR=`realpath $BASE_DIR/../../..`

BOKI_MACHINE_IAM=boki-ae-experiments
HELPER_SCRIPT=$ROOT_DIR/scripts/exp_helper

RUN=$1

$HELPER_SCRIPT start-machines --base-dir=$BASE_DIR --instance-iam-role=$BOKI_MACHINE_IAM

cd $BASE_DIR
if ! [ -f "$BASE_DIR/machines.json" ]; then
    echo "[ERROR] machines not started, skipping $BASE_DIR"
    rm ":~"
    exit 1
fi

NUM_KEYS=10000
QPS=(100) # QPS=(100 200 300 400)
NUM_OPS=(10)
READ_RATIO=(0.1 0.3 0.5 0.7 0.9)
VALUE_SIZE=(256)

for qps in ${QPS[@]}; do
    for ops in ${NUM_OPS[@]}; do
        for rr in ${READ_RATIO[@]}; do
            for v in ${VALUE_SIZE[@]}; do
                EXP_DIR=ReadRatio${rr}_QPS${qps}_v${v}
                if [ -d "$BASE_DIR/results/${EXP_DIR}_$RUN" ]; then
                    echo "finished $BASE_DIR/$EXP_DIR"
                    EXP_DIR=$BASE_DIR/results/${EXP_DIR}_$RUN
                    $ROOT_DIR/scripts/compute_latency.py --async-result-file $EXP_DIR/async_results >$EXP_DIR/latency.txt
                    continue
                fi
                sleep 60
                $BASE_DIR/run_once.sh $EXP_DIR $qps $ops $rr $v $NUM_KEYS # 2>&1 | tee $BASE_DIR/run.log 
                mv $BASE_DIR/results/$EXP_DIR $BASE_DIR/results/${EXP_DIR}_$RUN
                echo "finished $BASE_DIR/$EXP_DIR"
                EXP_DIR=$BASE_DIR/results/${EXP_DIR}_$RUN
                $ROOT_DIR/scripts/compute_latency.py --async-result-file $EXP_DIR/async_results >$EXP_DIR/latency.txt
            done
        done
    done
done

$HELPER_SCRIPT stop-machines --base-dir=$BASE_DIR
