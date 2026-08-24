FROM alpine:3.21
ENV TZ=Asia/Shanghai
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache ca-certificates tzdata \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
WORKDIR /app
COPY bin/ /prebuilt/
RUN set -e; \
    arch=$(uname -m); \
    case "$arch" in \
      aarch64) bin=server-linux-arm64 ;; \
      x86_64) bin=server-linux-amd64 ;; \
      *) echo "unsupported $arch"; exit 1 ;; \
    esac; \
    cp "/prebuilt/$bin" /server; \
    chmod +x /server
EXPOSE 8080
ENTRYPOINT ["/server"]
