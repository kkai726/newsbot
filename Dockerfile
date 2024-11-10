# 使用一个小的基础镜像来运行你的 Go 程序
FROM alpine:latest

# 安装必要的依赖，通常是 CA 证书
RUN apk --no-cache add ca-certificates

# 设置工作目录
WORKDIR /root/

# 将编译好的二进制文件从本机复制到容器中
COPY newsbot .

# 暴露容器的端口（例如 8080）
EXPOSE 8080

# 设置容器启动时运行的命令
CMD ["./newsbot"]
