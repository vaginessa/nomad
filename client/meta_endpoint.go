package client

import (
	"fmt"
	"time"

	metrics "github.com/armon/go-metrics"
	"github.com/hashicorp/nomad/nomad/structs"
	nstructs "github.com/hashicorp/nomad/nomad/structs"
)

type NodeMeta struct {
	c *Client
}

func newNodeMetaEndpoint(c *Client) *NodeMeta {
	n := &NodeMeta{c: c}
	return n
}

func (n *NodeMeta) Apply(args *structs.NodeMetaApplyRequest, reply *structs.NodeMetaResponse) error {
	defer metrics.MeasureSince([]string{"client", "node_meta", "apply"}, time.Now())

	// Check node write permissions
	if aclObj, err := n.c.ResolveToken(args.AuthToken); err != nil {
		return err
	} else if aclObj != nil && !aclObj.AllowNodeWrite() {
		return nstructs.ErrPermissionDenied
	}

	var err error

	newNode := n.c.UpdateNode(func(node *structs.Node) {
		// First update the Client's state store. This must be done
		// atomically with updating the metadata inmemory to avoid
		// interleaving updates causing incoherency between the state
		// store and inmemory.
		if err := n.c.stateDB.MergeNodeMeta(args.Meta); err != nil {
			err = fmt.Errorf("failed to apply dynamic node metadata: %w", err)
			return
		}

		for k, v := range args.Meta {
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

	// Trigger an async node update
	n.c.updateNode()

	reply.Meta = newNode.Meta
	return nil
}

func (n *NodeMeta) Read(args *structs.NodeSpecificRequest, resp *structs.NodeMetaResponse) error {
	defer metrics.MeasureSince([]string{"client", "node_meta", "reaj"}, time.Now())

	// Check node read permissions
	if aclObj, err := n.c.ResolveToken(args.AuthToken); err != nil {
		return err
	} else if aclObj != nil && !aclObj.AllowNodeRead() {
		return nstructs.ErrPermissionDenied
	}

	resp.Meta = n.c.Node().Meta

	return nil
}
