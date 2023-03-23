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
	flashStorage *flashStorage
	proto.UnimplementedWormholeServer

	mu                sync.Mutex
	taskConn          map[string]*TaskCore
	penetrateTaskConn *PenetrateTaskConnMap
	penetrateTask     map[string]proto.Wormhole_PenetrateTaskServer // 管理器
}

type PenetrateTaskConnMap struct {
	rmap map[string]chan net.Conn
	mu   sync.Mutex
}

func NewPenetrateTaskConnMap() *PenetrateTaskConnMap {
	return &PenetrateTaskConnMap{
		rmap: map[string]chan net.Conn{},
	}
}

func (p *PenetrateTaskConnMap) Storage(key string, conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	conns, ex := p.rmap[key]
	if !ex {
		p.rmap[key] = make(chan net.Conn, 9999)
		conns = p.rmap[key]
	}

	conns <- conn
}

func (p *PenetrateTaskConnMap) Get(key string) chan net.Conn {
	p.mu.Lock()
	defer p.mu.Unlock()
	conns, ex := p.rmap[key]
	if !ex {
		p.rmap[key] = make(chan net.Conn, 9999)
		conns = p.rmap[key]
	}

	return conns
}

func NewCoreServer() *CoreServer {
	fs := newFlashStorage()
	fs.init()

	core := &CoreServer{
		flashStorage:      fs,
		taskConn:          map[string]*TaskCore{},
		penetrateTaskConn: NewPenetrateTaskConnMap(),
		penetrateTask:     map[string]proto.Wormhole_PenetrateTaskServer{},
	}

	core.init()

	return core
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

		fmt.Println("ListenTCP: ", addr)
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
				c.penetrateTaskConn.Storage(core.task.TaskId, accept)
				server, ex := c.penetrateTask[core.task.Node.NodeId]
				if !ex {
					return
				}

				e := server.Send(&proto.PenetrateTaskResponse{
					TaskId:    core.task.TaskId,
					LocalPort: core.task.LocalPort,
				})
				if e != nil {
					log.Println(e)
					return
				}
			}

			fn()
		}
	}
}

func (c *CoreServer) RegisterNode(ctx context.Context, request *proto.RegisterNodeRequest) (*proto.RegisterNodeResponse, error) {
	c.flashStorage.registerNode(request.NodeId, request.NodeName, request.NodeIp)
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

	conn := <-c.penetrateTaskConn.Get(recv.TaskId)
	defer conn.Close()

	w := proto.WormholeReadWrite{server}
	go io.Copy(&w, conn)
	io.Copy(conn, &w)

	//go func() {
	//	for {
	//		request, e := server.Recv()
	//		if e != nil {
	//			log.Println(e)
	//			return
	//		}
	//		_, e = conn.Write(request.Data)
	//		if e != nil {
	//			log.Println(e)
	//			return
	//		}
	//	}
	//}()
	//
	//for {
	//	buffer := make([]byte, 1024)
	//	n, e := conn.Read(buffer)
	//	if e != nil {
	//		log.Println(e)
	//		break
	//	}
	//	server.Send(&proto.PenetrateResponse{
	//		Data: buffer[:n],
	//	})
	//}

	return nil
}

func (c *CoreServer) PenetrateTask(server proto.Wormhole_PenetrateTaskServer) error {
	recv, err := server.Recv()
	if err != nil {
		log.Println(err)
		return err
	}

	var fn = func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.penetrateTask[recv.NodeId] = server
	}

	fn()

	// 心跳
	for {
		_, err := server.Recv()
		if err != nil {
			log.Println(err)
			return nil
		}
	}
}
