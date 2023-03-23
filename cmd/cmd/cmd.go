package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"strings"

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
		fmt.Print("(token, nodeId, nodeName, nodeIp) $: ")
		fmt.Scanln(&input)
		input = strings.TrimSpace(input)
		split := strings.Split(input, ",")
		_, e := client.RegisterNode(context.TODO(), &proto.RegisterNodeRequest{
			Token:    split[0],
			NodeId:   split[1],
			NodeName: split[2],
			NodeIp:   split[3],
		})
		if e != nil {
			panic(e)
		}
		fmt.Println("RegisterNode: ")
		node, e := client.ListNode(context.TODO(), &proto.ListNodeRequest{})
		if e != nil {
			panic(e)
		}
		fmt.Println("------------------------------------")
		for _, v := range node.Nodes {
			fmt.Printf("NodeID: %s %s %s\n", v.NodeId, v.NodeName, v.NodeIp)
		}
		fmt.Println("------------------------------------")
		fmt.Println()
	case "2":
		node, e := client.ListNode(context.TODO(), &proto.ListNodeRequest{})
		if e != nil {
			panic(e)
		}
		fmt.Println("------------------------------------")
		for _, v := range node.Nodes {
			fmt.Printf("NodeID: %s %s %s\n", v.NodeId, v.NodeName, v.NodeIp)
		}
		fmt.Println("------------------------------------")
		fmt.Println()
	case "3":
		fmt.Print("(nodeId, remotePort, localPort) $: ")
		fmt.Scanln(&input)
		input = strings.TrimSpace(input)
		split := strings.Split(input, ",")
		remotePort, e := strconv.Atoi(split[1])
		if e != nil {
			panic(e)
		}
		localPort, e := strconv.Atoi(split[2])
		if e != nil {
			panic(e)
		}
		_, e = client.AddTask(context.TODO(), &proto.AddTaskRequest{
			NodeId:     split[0],
			RemotePort: int64(remotePort),
			LocalPort:  int64(localPort),
		})
		if e != nil {
			panic(e)
		}
		fmt.Println("Tasks: ")
		tasks, e := client.ListTask(context.TODO(), &proto.ListTaskRequest{})
		if e != nil {
			panic(e)
		}
		fmt.Println("------------------------------------")
		for _, v := range tasks.Tasks {
			fmt.Printf("TaskID: %s NodeID: %s %s %s %d %d\n", v.TaskId, v.Node.NodeId, v.Node.NodeName, v.Node.NodeIp, v.RemotePort, v.LocalPort)
		}
		fmt.Println("------------------------------------")
		fmt.Println()
	case "4":
		tasks, e := client.ListTask(context.TODO(), &proto.ListTaskRequest{})
		if e != nil {
			panic(e)
		}
		fmt.Println("------------------------------------")
		for _, v := range tasks.Tasks {
			fmt.Printf("TaskID: %s NodeID: %s %s %s %d %d\n", v.TaskId, v.Node.NodeId, v.Node.NodeName, v.Node.NodeIp, v.RemotePort, v.LocalPort)
		}
		fmt.Println("------------------------------------")
		fmt.Println()
	case "5":
		fmt.Print("($: ")
		fmt.Scanln(&input)
		input = strings.TrimSpace(input)
		_, e := client.DelTask(context.TODO(), &proto.DelTaskRequest{
			TaskId: input,
		})
		if e != nil {
			panic(e)
		}
		fmt.Println("Tasks: ")
		tasks, e := client.ListTask(context.TODO(), &proto.ListTaskRequest{})
		if e != nil {
			panic(e)
		}
		fmt.Println("------------------------------------")
		for _, v := range tasks.Tasks {
			fmt.Printf("TaskID: %s NodeID: %s %s %s %d %d\n", v.TaskId, v.Node.NodeId, v.Node.NodeName, v.Node.NodeIp, v.RemotePort, v.LocalPort)
		}
		fmt.Println("------------------------------------")
		fmt.Println()
	}
}
