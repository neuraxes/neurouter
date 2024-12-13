FROM alpine

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories \
    && apk add --no-cache ca-certificates

ADD bin/router /bin/router
ADD configs/config.yaml /configs/config.yaml

EXPOSE 8000
EXPOSE 9000

ENTRYPOINT [ "/bin/router", "-conf", "/configs/"]
