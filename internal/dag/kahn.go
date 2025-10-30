package dag

import (
	"errors"
	"fmt"
)

type Node struct {
	IncomingEdges []string
	Name          string
}

func SortDAG(nodes []Node) ([]string, error) {
	// https://en.wikipedia.org/wiki/Topological_sorting#Kahn's_algorithm
	// Kahn's algorithm
	// L ← Empty list that will contain the sorted elements
	// S ← Set of all nodes with no incoming edge
	// while S is non-empty do
	//     remove a node n from S
	//     add n to tail of L
	//     for each node m with an edge e from n to m do
	//         remove edge e from the graph
	//         if m has no other incoming edges then
	//             insert m into S
	// if graph has edges then
	//     return error (graph has at least one cycle)
	// else
	//     return L (a topologically sorted order)

	// In our DAG
	// Edges defines as A -> B
	// B(Blocked) 'Depends On' A(Blocker)
	// A(Blocker) 'Blocks'     B(Blocked)
	// So 'no incoming edge' = nodes which are not blocked / only act as A

	// S is a set of unblocked nodes.
	S := make([]string, 0)
	// Edges is a list of B Blocked -> A Blocker
	type edge struct {
		From string
		To   string
		Used bool
	}
	edges := make([]*edge, 0)

	for _, node := range nodes {
		incoming := node.IncomingEdges
		if len(incoming) == 0 {
			S = append(S, node.Name)
		} else {
			for _, blockedNode := range incoming {
				edges = append(edges, &edge{From: blockedNode, To: node.Name})
			}
		}
	}

	L := make([]string, 0)
	var n string
	for len(S) > 0 {
		// remove a node n from S
		n, S = S[0], S[1:]
		// add n to tail of L
		L = append(L, n)

		// for each node m with an edge e from n to m do
		for _, edge := range edges {
			if edge.Used {
				continue
			}
			if edge.From != n {
				continue
			}
			m := edge.To

			// remove edge e from the graph
			edge.Used = true

			// if m has no other incoming edges then
			found := false
			for _, edge := range edges {
				if edge.Used {
					continue
				}
				if edge.To == m {
					found = true
					break
				}
			}
			if !found {
				// insert m into S
				S = append(S, m)
			}
		}
	}

	errs := make([]error, 0)
	cyclic := false
	for _, edge := range edges {
		if edge.Used {
			continue
		}
		cyclic = true
		errs = append(errs, fmt.Errorf("edge %s -> %s is cyclic", edge.From, edge.To))
	}
	if cyclic {
		joined := errors.Join(errs...)
		return nil, fmt.Errorf("graph has at least one cycle: %w", joined)
	}

	return L, nil
}
