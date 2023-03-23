package core

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/dollarkillerx/wormhole/internal/proto"
)

type CoreServer struct {
	flashStorage flashStorage
	proto.UnimplementedWormholeServer

	mu            sync.Mutex
	taskConn      map[string]*TaskCore
	penetrate     map[string]chan proto.Wormhole_PenetrateServer
	penetrateTask map[string]proto.Wormhole_PenetrateTaskServer // 管理器
}

func NewCoreServer() *CoreServer {
	return &CoreServer{
		taskConn:      map[string]*TaskCore{},
		penetrate:     map[string]chan proto.Wormhole_PenetrateServer{},
		penetrateTask: map[string]proto.Wormhole_PenetrateTaskServer{},
	}
}

type TaskCore struct {
	listener *net.TCPListener
	close    chan struct{}
	task     *proto.Task
}

func (c *CoreServer) init() {
	tasks := c.flashStorage.listTask()
	for i, v := range tasks {
		addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%d", v.RemotePort))
		if err != nil {
			log.Println(err)
			continue
		}
		listener, err := net.ListenTCP("tcp", addr)
		if err != nil {
			log.Println(err)
			continue
		}

		var fn = func() {
			c.mu.Lock()
			defer c.mu.Unlock()

			core := &TaskCore{
				listener: listener,
				close:    make(chan struct{}),
				task:     tasks[i],
			}

			go c.processor(core)
			c.taskConn[v.TaskId] = core
		}

		fn()
	}
}

func (c *CoreServer) processor(core *TaskCore) {
loop:
	for {
		select {
		case <-core.close:
			break loop
		default:
			var fn = func() {
				accept, err := core.listener.Accept()
				if err != nil {
					return
				}

				defer accept.Close()

				server, ex := c.penetrateTask[core.task.Node.NodeId]
				if !ex {
					return
				}

				e := server.Send(&proto.PenetrateTaskResponse{
					TaskId:    core.task.TaskId,
					LocalPort: core.task.RemotePort,
				})
				if e != nil {
					log.Println(err)
					return
				}

				ser := <-c.penetrate[core.task.TaskId]

				wm := proto.WormholeReadWrite{ser}

				go io.Copy(&wm, accept)
				io.Copy(accept, &wm)
			}

			fn()
		}
	}
}

func (c *CoreServer) RegisterNode(ctx context.Context, request *proto.RegisterNodeRequest) (*proto.RegisterNodeResponse, error) {
	c.flashStorage.registerNode(request.NodeIp, request.NodeName, request.NodeIp)
	return &proto.RegisterNodeResponse{}, nil
}

func (c *CoreServer) ListNode(ctx context.Context, request *proto.ListNodeRequest) (*proto.ListNodeResponse, error) {
	return &proto.ListNodeResponse{Nodes: c.flashStorage.listNode()}, nil
}

func (c *CoreServer) AddTask(ctx context.Context, request *proto.AddTaskRequest) (*proto.AddTaskResponse, error) {
	taskID, err := c.flashStorage.addTask(request.NodeId, request.RemotePort, request.LocalPort)
	return &proto.AddTaskResponse{TaskId: taskID}, err
}

func (c *CoreServer) ListTask(ctx context.Context, request *proto.ListTaskRequest) (*proto.ListTaskResponse, error) {
	return &proto.ListTaskResponse{Tasks: c.flashStorage.listTask()}, nil
}

func (c *CoreServer) DelTask(ctx context.Context, request *proto.DelTaskRequest) (*proto.DelTaskResponse, error) {
	c.flashStorage.delTask(request.TaskId)
	return &proto.DelTaskResponse{}, nil
}

func (c *CoreServer) Penetrate(server proto.Wormhole_PenetrateServer) error {
	recv, err := server.Recv()
	if err != nil {
		log.Println(err)
		return err
	}

	ser := c.penetrate[recv.TaskId]
	ser <- server

	return nil
}

func (c *CoreServer) PenetrateTask(server proto.Wormhole_PenetrateTaskServer) error {
	recv, err := server.Recv()
	if err != nil {
		log.Println(err)
		return err
	}

	// 心跳
	go func() {
		for {
			_, err := server.Recv()
			if err != nil {
				log.Println(err)
				return
			}
		}
	}()

	var fn = func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.penetrateTask[recv.NodeId] = server
	}

	fn()

	return nil
}
