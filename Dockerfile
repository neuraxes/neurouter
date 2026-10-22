FROM alpine

RUN apk add --no-cache ca-certificates

COPY --chmod=755 bin/neurouter /bin/neurouter
ADD configs/config.yaml /configs/config.yaml

EXPOSE 8000
EXPOSE 9000

ENTRYPOINT [ "/bin/neurouter", "-conf", "/configs/"]
