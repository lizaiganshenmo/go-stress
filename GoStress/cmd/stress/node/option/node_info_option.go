package option

import (
	"github.com/lizaiganshenmo/GoStress/library/utils"
)

var (
	defaultIP = utils.GetOutboundIP()
)

type NodeInfoOption struct {
	F func(o *NodeInfoOptions)
}

type NodeInfoOptions struct {
	Host   string // 节点主机ip
	NodeId int64  // 节点id
}

func (o *NodeInfoOptions) Apply(opts []NodeInfoOption) {
	for _, op := range opts {
		op.F(o)
	}
}

func NewNodeInfoOptions(opts []NodeInfoOption) *NodeInfoOptions {
	o := &NodeInfoOptions{
		Host: defaultIP,
	}

	o.Apply(opts)
	return o
}

func WithNodeID(id int64) NodeInfoOption {
	return NodeInfoOption{F: func(o *NodeInfoOptions) {
		o.NodeId = id
	}}
}

func WithHost(host string) NodeInfoOption {
	return NodeInfoOption{F: func(o *NodeInfoOptions) {
		o.Host = host
	}}
}
