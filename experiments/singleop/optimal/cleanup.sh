#!/bin/bash

kubectl delete pods zookeeper zookeeper-setup boki-gateway boki-controller
kubectl delete daemonsets boki-engine boki-sequencer boki-storage nop singleop prewarm
kubectl delete service zookeeper-service
kubectl delete configmap env-config
