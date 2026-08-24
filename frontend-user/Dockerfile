FROM nginx:1.27-alpine
ENV TZ=Asia/Shanghai
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache wget tzdata \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY dist /usr/share/nginx/html
EXPOSE 80
