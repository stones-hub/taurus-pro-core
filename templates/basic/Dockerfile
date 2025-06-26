# 使用官方的 Go 语言镜像作为基础镜像
FROM golang:1.24-alpine AS builder

# 设置Goproxy防止有些机器没办法访问外网
ENV GOPROXY=https://goproxy.cn,direct

# 设置工作目录, 容器启动后会进入该目录
ARG WORKDIR
WORKDIR ${WORKDIR}

# 将 go.mod 和 go.sum 复制到工作目录
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 将项目的所有文件复制到工作目录
COPY . .

# 编译 Go 应用程序
RUN go build -o main cmd/main.go

# 使用一个更小的基础镜像来运行应用程序
FROM alpine:latest

# 安装必要的依赖, tzdata和ENV TZ=Asia/Shanghai 配合使用， 保证镜像时区的正确性
RUN apk --no-cache add ca-certificates curl tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 设置工作目录
ARG WORKDIR
# 设置应用配置文件
ARG APP_CONFIG

RUN echo "APP_CONFIG is set to: ${APP_CONFIG}"
RUN echo "WORKDIR is set to: ${WORKDIR}"

# 设置工作目录
WORKDIR ${WORKDIR}

# 从构建阶段复制编译好的二进制文件
# 在 Dockerfile 中，COPY 指令用于将文件或目录从构建阶段复制到新镜像中。
# 这里，--from=builder 指定从构建阶段（builder）中复制文件，/app/main 是构建阶段中编译好的二进制文件路径。
COPY --from=builder ${WORKDIR}/main .
# 复制应用配置文件
COPY ${APP_CONFIG} ${WORKDIR}/config
# 复制静态文件
COPY ./static ${WORKDIR}/static
# 复制模板文件
COPY ./templates ${WORKDIR}/templates
# 运行应用程序, 为什么这里的配置文件路径是${WORKDIR}/config, 是因为我在Makefile中 docker run的时候bind的目录就是这个
CMD ["sh", "-c", "./main -config=${WORKDIR}/config"]