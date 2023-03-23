package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/dollarkillerx/wormhole/internal/ca"
	"github.com/dollarkillerx/wormhole/internal/proto"
	"google.golang.org/grpc"
)

var rpcAddr string

func init() {
	flag.StringVar(&rpcAddr, "r", "127.0.0.1:8454", "rpc addr")
}

var menu1 = `
#################################
Wormhole 
RPCAddr: %s
#################################
1. RegisterNode
2. ListNode
3. AddTask
4. ListTask
5. DelTask
`

func main() {
	credentials, err := ca.LoadTLSCredentials([]byte(ca.ClientPem), "www.p-pp.cn")
	if err != nil {
		panic(err)
	}

	conn, err := grpc.Dial(rpcAddr, grpc.WithTransportCredentials(credentials))
	if err != nil {
		panic(err)
	}

	client := proto.NewWormholeClient(conn)

	var input string
	fmt.Printf(menu1, rpcAddr)
	fmt.Print("$: ")
	fmt.Scanln(&input)
	switch input {
	case "1":
		client.RegisterNode(context.TODO(), &proto.RegisterNodeRequest{})
	case "2":

	case "3":

	case "4":

	case "5":

	}
}
