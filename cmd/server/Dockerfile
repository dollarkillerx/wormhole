FROM alpine:latest
WORKDIR /app
COPY ./cmd/server/wormhole_server /app
COPY ./configs/ /app/configs
CMD ["/app/wormhole_server"]