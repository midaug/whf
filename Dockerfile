FROM golang:1.18.0-alpine3.15
LABEL maintainer="midaug <days0814@gmail.com>"
# go mod download && \
COPY ./* /data/
WORKDIR /data
# 更新安装源未国内
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk --no-cache add git
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://proxy.golang.com.cn,direct
RUN go mod tidy
RUN go build .



FROM alpine:3.15
LABEL maintainer="midaug <days0814@gmail.com>"

RUN mkdir -p /data/js
COPY --from=0 /data/whf /data/whf
COPY ./js/* /data/js/

VOLUME ["/data/js"]
WORKDIR  /data
EXPOSE 9090

ENTRYPOINT ["/data/whf"]