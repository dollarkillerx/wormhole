package main

import (
	"github.com/dollarkillerx/wormhole/internal/ca"
	"github.com/dollarkillerx/wormhole/internal/core"
	"github.com/dollarkillerx/wormhole/internal/proto"
	"google.golang.org/grpc"

	"flag"
	"fmt"
	"log"
	"net"
)

var rpcAddr string

func init() {
	flag.StringVar(&rpcAddr, "r", "0.0.0.0:8454", "rpc addr")
	flag.Parse()
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	lis, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	creds, err := ca.LoadTLSServerCredentials([]byte(ca.ServerPem), []byte(ca.ServerKey))
	if err != nil {
		log.Fatalln(err)
	}

	srv := grpc.NewServer(
		grpc.Creds(creds),
	)

	server := core.NewCoreServer()

	proto.RegisterWormholeServer(srv, server)

	fmt.Println("GRPC: ", rpcAddr)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
