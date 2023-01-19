package command

import (
	"fmt"
	"strings"

	"github.com/hashicorp/nomad/api"
	"github.com/mitchellh/cli"
)

type NodeCommand struct {
	Meta
}

func (f *NodeCommand) Help() string {
	helpText := `
Usage: nomad node <subcommand> [options] [args]

  This command groups subcommands for interacting with nodes. Nodes in Nomad are
  agent's that can run submitted workloads. This command can be used to examine
  nodes and operate on nodes, such as draining workloads off of them.

  Examine the status of a node:

      $ nomad node status <node-id>

  Mark a node as ineligible for running workloads. This is useful when the node
  is expected to be removed or upgraded so new allocations aren't placed on it:

      $ nomad node eligibility -disable <node-id>

  Mark a node to be drained, allowing batch jobs four hours to finish before
  forcing them off the node:

      $ nomad node drain -enable -deadline 4h <node-id>

  Please see the individual subcommand help for detailed usage information.
`

	return strings.TrimSpace(helpText)
}

func (f *NodeCommand) Synopsis() string {
	return "Interact with nodes"
}

func (f *NodeCommand) Name() string { return "node" }

func (f *NodeCommand) Run(args []string) int {
	return cli.RunResultHelp
}

// lookupNodeID looks up a nodeID prefix and returns the full ID or an error.
// The error will always be suitable for displaying to users.
func lookupNodeID(client *api.Nodes, nodeID string) (string, error) {
	if len(nodeID) == 1 {
		return "", fmt.Errorf("Node ID must contain at least two characters.")
	}

	nodeID = sanitizeUUIDPrefix(nodeID)
	nodes, _, err := client.PrefixList(nodeID)
	if err != nil {
		return "", fmt.Errorf("Error querying node: %w", err)
	}

	if len(nodes) == 0 {
		return "", fmt.Errorf("No node(s) with prefix or id %q found", nodeID)
	}

	if len(nodes) > 1 {
		return "", fmt.Errorf("Prefix matched multiple nodes\n\n%s",
			formatNodeStubList(nodes, true))
	}

	return nodes[0].ID, nil
}
