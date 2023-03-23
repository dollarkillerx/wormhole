package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/dollarkillerx/wormhole/internal/ca"
	"github.com/dollarkillerx/wormhole/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var rpcAddr string
var nodeId string

func init() {
	flag.StringVar(&rpcAddr, "r", "127.0.0.1:8454", "rpc addr")
	flag.StringVar(&nodeId, "n", "node1", "node id")
	flag.Parse()
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	// 设置心跳间隔和超时时间
	keepAliveParams := keepalive.ClientParameters{
		Time:                10 * time.Second, // 心跳间隔
		Timeout:             5 * time.Second,  // 超时时间
		PermitWithoutStream: true,             // 允许无流量的连接
	}

	credentials, err := ca.LoadTLSCredentials([]byte(ca.ClientPem), "www.p-pp.cn")
	if err != nil {
		panic(err)
	}

lc:
	for {
		conn, err := grpc.Dial(rpcAddr, grpc.WithTransportCredentials(credentials), grpc.WithKeepaliveParams(keepAliveParams), grpc.WithBlock())
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second * 3)
			continue
		}

		client := proto.NewWormholeClient(conn)

		// 主线程
		task, err := client.PenetrateTask(context.TODO())
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second * 3)
			continue
		}

		// 心跳
		go func() {
			for {
				fmt.Println("node id: ", nodeId)
				e := task.Send(&proto.PenetrateTaskRequest{
					NodeId: nodeId,
				})
				if e != nil {
					log.Println(e)
					return
				}

				time.Sleep(time.Second * 3)
			}
		}()

		for {
			recv, e := task.Recv()
			if e != nil {
				log.Println(e)
				time.Sleep(time.Second * 3)
				continue lc
			}

			if recv.Heartbeat {
				continue
			}

			penetrate, e := client.Penetrate(context.TODO())
			if e != nil {
				log.Println(err)
				continue
			}

			e = penetrate.Send(&proto.PenetrateRequest{
				TaskId: recv.TaskId,
				Data:   nil,
			})
			if e != nil {
				log.Println(e)
				continue
			}

			go func() {
				//fmt.Println("dial: ", fmt.Sprintf("127.0.0.1:%d", recv.LocalPort))
				dial, e := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", recv.LocalPort))
				if e != nil {
					log.Println(e)
					return
				}
				defer dial.Close()

				w := proto.WormholePenetrateClientReadWrite{penetrate}
				go io.Copy(&w, dial)
				io.Copy(dial, &w)
			}()
		}
	}
}
