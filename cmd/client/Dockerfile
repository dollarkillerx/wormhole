FROM alpine:latest
WORKDIR /app
COPY ./cmd/client/wormhole_client /app
COPY ./configs/ /app/configs
CMD ["/app/wormhole_client"]