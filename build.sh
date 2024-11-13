#!/bin/bash

# 定义镜像名称和标签
IMAGE_NAME="newsbot"
TAG="latest"
NETWORK_NAME="my_network"  # 已经存在的 Docker 网络名称
REDIS_CONTAINER_NAME="my-redis"  # 已经存在的 Redis 容器名称
REDIS_PORT=6379             # Redis 容器的端口

# 检查自定义 Docker 网络是否存在
echo "Checking if Docker network $NETWORK_NAME exists..."
docker network inspect $NETWORK_NAME &>/dev/null
if [ $? -ne 0 ]; then
    echo "Docker network $NETWORK_NAME does not exist. Please create it first."
    exit 1
fi

# 检查 Redis 容器是否存在并在运行
echo "Checking if Redis container $REDIS_CONTAINER_NAME is running..."
docker ps --filter "name=$REDIS_CONTAINER_NAME" --format "{{.ID}}" | grep -q . 
if [ $? -ne 0 ]; then
    echo "Redis container $REDIS_CONTAINER_NAME is not running! Please start it first."
    exit 1
fi

# 构建 Docker 镜像
echo "Building Docker image..."
build_output=$(docker build -t $IMAGE_NAME:$TAG . 2>&1)
build_status=$?

# 检查构建是否成功
if [ $build_status -ne 0 ]; then
  echo "Docker image build failed!"
  echo "Error output:"
  echo "$build_output"
  exit 1
fi

# 启动应用容器，连接到已存在的 Docker 网络
echo "Running Docker container $IMAGE_NAME..."
docker run -d -p 10086:8080 --name $IMAGE_NAME --network $NETWORK_NAME $IMAGE_NAME:$TAG

# 检查容器是否成功启动
if [ $? -ne 0 ]; then
  echo "Failed to start the Docker container!"
  exit 1
fi

echo "Docker container is running and exposed on port 10086."
