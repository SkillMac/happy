package hCluster

import (
	"custom/happy/hRpc"
	"errors"
	"math/rand"
)

type NodeID struct {
	Addr          string
	nodeComponent *NodeComponent
}

var NotExistIndex error = errors.New("[ERROR] node info detial index is not exist")

func (this *NodeID) GetClient() (*rpc.TcpClient, error) {
	if this.Addr == "" {
		return nil, errors.New("this node id is empty")
	}
	return this.nodeComponent.GetNodeClient(this.Addr)
}

//无需加锁，只读
type NodeIDGroup struct {
	nodeComponent *NodeComponent
	nodes         []*InquiryReply
}

func NewNodeIDGrop() *NodeIDGroup {
	return &NodeIDGroup{nodes: []*InquiryReply{}}
}

//所有节点，仅地址
func (this *NodeIDGroup) Nodes() []string {
	nodes := make([]string, len(this.nodes))
	for _, v := range this.nodes {
		nodes = append(nodes, v.Node)
	}
	return nodes
}

//所有节点，详细信息
func (this *NodeIDGroup) NodesDetail() []*InquiryReply {
	return this.nodes
}

//随机选择一个
func (this *NodeIDGroup) RandOne() (string, error) {
	if this.nodes == nil {
		return "", errors.New("this node id group is empty")
	}
	length := len(this.nodes)
	if length == 0 {
		return "", errors.New("this node id group is empty")
	}
	index := rand.Intn(length)
	return this.nodes[index].Node, nil
}

//随机选择一个
func (this *NodeIDGroup) RandClient() (*rpc.TcpClient, error) {
	length := len(this.nodes)
	if length == 0 {
		return nil, errors.New("this node id group is empty")
	}
	index := rand.Intn(length)

	return this.nodeComponent.GetNodeClient(this.nodes[index].Node)
}

//所有客户端
func (this *NodeIDGroup) Clients() ([]*rpc.TcpClient, error) {
	length := len(this.nodes)
	if length == 0 {
		return nil, errors.New("this node id group is empty")
	}
	clients := []*rpc.TcpClient{}
	for _, nodeID := range this.nodes {
		client, err := this.nodeComponent.GetNodeClient(nodeID.Node)
		if err != nil {
			continue
		}
		clients = append(clients, client)

	}
	if len(clients) <= 0 {
		return nil, errors.New("this node id group is empty")
	}
	return clients, nil
}

func (this *NodeIDGroup) SelectOneNodeInfo(selectType SelectorType) (*NodeInfoDetail, error) {
	index := SourceGroup(this.nodes).DoQuery(selectType)
	if index == -1 {
		return nil, NotExistIndex
	}

	return &NodeInfoDetail{
		Node:       this.nodes[index].Node,
		Info:       this.nodes[index].Info,
		CustomData: this.nodes[index].CustomData,
	}, nil
}

//选择一个负载最低的节点
func (this *NodeIDGroup) MinLoadClient() (*rpc.TcpClient, error) {
	if len(this.nodes) == 0 {
		return nil, errors.New("this node id group is empty")
	}
	index := SourceGroup(this.nodes).SelectMinLoad()
	return this.nodeComponent.GetNodeClient(this.nodes[index].Node)
}
