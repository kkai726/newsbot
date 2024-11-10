#!/bin/bash

# 定义镜像名称和标签
IMAGE_NAME="newsbot"
TAG="latest"

# 构建 Docker 镜像
echo "Building Docker image..."
docker build -t $IMAGE_NAME:$TAG .

# 检查构建是否成功
if [ $? -ne 0 ]; then
  echo "Docker image build failed!"
  exit 1
fi

# 运行 Docker 容器，映射容器端口 8080 到主机的端口 10086
echo "Running Docker container..."
docker run -d -p 10086:8080 --name newsbot $IMAGE_NAME:$TAG

# 检查容器是否成功启动
if [ $? -ne 0 ]; then
  echo "Failed to start the Docker container!"
  exit 1
fi

echo "Docker container is running and exposed on port 10086."
