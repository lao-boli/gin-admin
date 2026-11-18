# syntax=docker/dockerfile:1
# 国内镜像源：默认 DaoCloud 加速；其它可选 docker.1ms.run / dockerpull.org / dockerproxy.net
# 可通过 docker compose 的 REGISTRY 变量或 --build-arg 覆盖（见 docker-compose.yml）
ARG REGISTRY=docker.m.daocloud.io

# ---------- 构建 ----------
FROM ${REGISTRY}/library/golang:1.22-alpine AS builder

# Go 模块代理走国内
ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY} \
    CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags "-s -w" -o /src/ginadmin .

# ---------- 运行 ----------
FROM ${REGISTRY}/library/alpine:3.20

# alpine 包源换中科大，并装时区/证书
RUN sed -i 's#dl-cdn.alpinelinux.org#mirrors.ustc.edu.cn#g' /etc/apk/repositories \
    && apk add --no-cache ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app
COPY --from=builder /src/ginadmin /app/ginadmin
# 配置目录：start 用 -d 指定 WorkDir=configs，-c prod 读 configs/prod/*.toml
# menu_cn.json / rbac_model.conf 在 configs 根目录，所以整个 configs 一起拷
COPY configs /app/configs
RUN mkdir -p /app/data

EXPOSE 8040
ENTRYPOINT ["/app/ginadmin", "start", "-d", "/app/configs", "-c", "prod"]
