# 构建阶段：编译出静态二进制（CGO_ENABLED=0，运行镜像不需要 gcc 和 libc 兼容层）
FROM golang:1.25-alpine AS build

WORKDIR /src

# 先只拷依赖清单：代码改动不会让这两层的缓存失效
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# -trimpath 去掉本机路径，-s -w 去掉符号表和调试信息
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/task-system .

# 运行阶段：只带二进制和一个最小系统
FROM alpine:3.20

RUN adduser -D -u 10001 app
COPY --from=build /out/task-system /usr/local/bin/task-system

# 不用 root 跑服务
USER app

EXPOSE 8080

ENTRYPOINT ["task-system"]
