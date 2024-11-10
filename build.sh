#!/bin/bash

# 退出脚本时如果有任何命令失败
set -e

# 编译 Go 程序
echo "Building Go application..."
go build -o newsbot main.go
if [ $? -ne 0 ]; then
  echo "Go build failed!"
  exit 1
fi

# 构建 Docker 镜像
echo "Building Docker image..."
docker build -t newsbot .
if [ $? -ne 0 ]; then
  echo "Docker build failed!"
  exit 1
fi

# 运行 Docker 容器
echo "Running Docker container..."
docker run -p 10086:8080 newsbot
if [ $? -ne 0 ]; then
  echo "Docker run failed!"
  exit 1
fi

echo "Build and run completed successfully."
