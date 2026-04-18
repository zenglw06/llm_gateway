#!/bin/bash

# 启动脚本
set -e

CONFIG_FILE="configs/config.yaml"

if [ ! -f "$CONFIG_FILE" ]; then
    echo "Config file $CONFIG_FILE not found, copying from example..."
    cp configs/config.example.yaml $CONFIG_FILE
fi

echo "Starting LLM Gateway..."
exec ./bin/llm-gateway -c $CONFIG_FILE "$@"
