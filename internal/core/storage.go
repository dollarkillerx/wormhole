package core

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/dollarkillerx/wormhole/internal/proto"
	"github.com/google/uuid"
)

type flashStorage struct {
	mu sync.Mutex

	Nodes []*proto.Node `json:"nodes"`
	Tasks []*proto.Task `json:"tasks"`
}

func newFlashStorage() *flashStorage {
	return &flashStorage{
		Nodes: []*proto.Node{},
		Tasks: []*proto.Task{},
	}
}

func (s *flashStorage) init() {
	file, err := os.ReadFile("wormhole_storage.json")
	if err == nil {
		err := json.Unmarshal(file, s)
		if err != nil {
			panic(err)
		}
	}
}

func (s *flashStorage) flash() {
	//s.mu.Lock()
	//defer s.mu.Unlock()

	marshal, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("wormhole_storage.json", marshal, 00666)
	if err != nil {
		panic(err)
	}
}

func (s *flashStorage) registerNode(nodeId string, nodeName string, nodeIp string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.flash()

	var ex bool

	// 存在更新
	for i := range s.Nodes {
		if s.Nodes[i].NodeId == nodeId {
			ex = true
			s.Nodes[i].NodeName = nodeName
			s.Nodes[i].NodeIp = nodeIp
		}
	}

	// 不存在就插入
	if !ex {
		s.Nodes = append(s.Nodes, &proto.Node{
			NodeId:   nodeId,
			NodeName: nodeName,
			NodeIp:   nodeIp,
		})
	}
}

func (s *flashStorage) listNode() []*proto.Node {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result []*proto.Node

	for i := range s.Nodes {
		result = append(result, s.Nodes[i])
	}

	return result
}

func (s *flashStorage) getNodeById(nodeId string) (*proto.Node, bool) {
	for i := range s.Nodes {
		if s.Nodes[i].NodeId == nodeId {
			return s.Nodes[i], true
		}
	}

	return nil, false
}

func (s *flashStorage) addTask(nodeId string, remotePort int64, localPort int64) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.flash()

	// 获取node
	node, ex := s.getNodeById(nodeId)
	if !ex {
		return "", errors.New("not found node")
	}

	taskID := uuid.New().String()
	s.Tasks = append(s.Tasks, &proto.Task{
		TaskId:     taskID,
		RemotePort: remotePort,
		LocalPort:  localPort,
		Node:       node,
	})

	return taskID, nil
}

func (s *flashStorage) listTask() (tasks []*proto.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.Tasks {
		tasks = append(tasks, s.Tasks[i])
	}

	return tasks
}

func (s *flashStorage) delTask(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.flash()

	var tasks []*proto.Task

	for i := range s.Tasks {
		if s.Tasks[i].TaskId != taskID {
			tasks = append(tasks, s.Tasks[i])
		}
	}

	s.Tasks = tasks
}
