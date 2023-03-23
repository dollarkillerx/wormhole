#!/bin/sh

source_path=./cmd/server
go_file=main.go
image_name=wormhole_server
build_output=wormhole_server
version=0.0.1

echo CGO_ENABLED=0 GOOS="linux" GOARCH="amd64" go build -o $source_path/$build_output $source_path/$go_file
CGO_ENABLED=0 GOOS="linux" GOARCH="amd64" go build -o $source_path/$build_output $source_path/$go_file

upx $source_path/$build_output

docker rmi -f $image_name:$version
docker build -f $source_path/Dockerfile -t $image_name:$version  .
rm $source_path/$build_output

docker save -o $image_name-$version.tar $image_name:$version