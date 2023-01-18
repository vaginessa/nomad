package client

import (
	"fmt"
	"time"

	metrics "github.com/armon/go-metrics"
	"github.com/hashicorp/nomad/nomad/structs"
)

type NodeMeta struct {
	c *Client
}

func newNodeMetaEndpoint(c *Client) *NodeMeta {
	n := &NodeMeta{c: c}
	return n
}

func (n *NodeMeta) Apply(req *structs.NodeMetaApplyRequest, resp *structs.NodeMetaResponse) error {
	defer metrics.MeasureSince([]string{"client", "node_meta", "apply"}, time.Now())

	//TODO permissions check

	var err error

	newNode := n.c.UpdateNode(func(node *structs.Node) {
		// First update the Client's state store. This must be done
		// atomically with updating the metadata inmemory to avoid
		// interleaving updates causing incoherency between the state
		// store and inmemory.
		if err := n.c.stateDB.MergeNodeMeta(req.Meta); err != nil {
			err = fmt.Errorf("failed to apply dynamic node metadata: %w", err)
			return
		}

		for k, v := range req.Meta {
			if v == nil {
				delete(node.Meta, k)
				continue
			}

			node.Meta[k] = *v
		}
	})

	if err != nil {
		return err
	}

	resp.Meta = newNode.Meta
	return nil
}

func (n *NodeMeta) Read(req *structs.NodeSpecificRequest, resp *structs.NodeMetaResponse) error {
	defer metrics.MeasureSince([]string{"client", "node_meta", "reaj"}, time.Now())

	//TODO permissions check

	resp.Meta = n.c.Node().Meta

	return nil
}
