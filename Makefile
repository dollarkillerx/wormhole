build:
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w"  -o wormhole_client  cmd/client/main.go
	GOARCH=amd64 GOOS=windows CGO_ENABLED=0 go build -ldflags "-s -w" -o wormhole_client.exe cmd/client/main.go
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w"  -o wormhole cmd/server/main.go
