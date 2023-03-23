build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w"  -o wormhole_client_linux_x86  cmd/client/main.go
	GOARCH=amd64 GOOS=windows CGO_ENABLED=0 go build -ldflags "-s -w" -o wormhole_client.exe cmd/client/main.go
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w"  -o wormhole_server_linux_x86 cmd/server/main.go

docker_build:
	docker build -f cmd/server/Dockerfile -t dollarkiller/wormhole:latest  .
	docker build -f cmd/client/Dockerfile -t dollarkiller/wormhole_client:latest  .


build_proto_protocol:
	protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        internal/proto/*.proto