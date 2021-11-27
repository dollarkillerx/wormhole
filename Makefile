build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w"  -o wormhole_client  cmd/client/main.go
	GOARCH=amd64 GOOS=windows CGO_ENABLED=0 go build -ldflags "-s -w" -o wormhole_client.exe cmd/client/main.go
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w"  -o wormhole cmd/server/main.go

docker_build:
	docker build -f cmd/server/Dockerfile -t dollarkiller/wormhole:latest  .
	docker build -f cmd/server/Dockerfile -t dollarkiller/wormhole_client:latest  .
