#!/bin/bash

NODE_ID=$(echo "$NODE_NAME" | grep -oE '[0-9]+$')
echo "export FAAS_NODE_ID=$NODE_ID" >> ~/.bashrc
source ~/.bashrc

FAAS_NODE_ID=$NODE_ID /boki/sequencer \
    --zookeeper_host=10.10.1.8:2181 \
    --listen_iface=eth0 \
    --num_io_workers=2 \
    --io_uring_entries=2048 \
    --io_uring_fd_slots=4096 \
    --slog_global_cut_interval_us=300 \
    # --v=1