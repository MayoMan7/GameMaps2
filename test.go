package main

// import (
// 	"encoding/xml"
// 	"os"
// )

// type graphML struct {
// 	Graph graph `xml:"graph"`
// }
// type graph struct {
// 	Nodes []node `xml:"node"`
// 	Edges []edge `xml:"edge"`
// }
// type node struct {
// 	ID string `xml:"id,attr"`
// }
// type edge struct {
// 	Source string `xml:"source,attr"`
// 	Target string `xml:"target,attr"`
// }

// type Adj map[string][]string

// func LoadGraphML(path string) (Adj, error) {
// 	b, err := os.ReadFile(path)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var g graphML
// 	if err := xml.Unmarshal(b, &g); err != nil {
// 		return nil, err
// 	}

// 	adj := make(Adj, len(g.Graph.Nodes))
// 	// ensure all nodes exist in map
// 	for _, n := range g.Graph.Nodes {
// 		adj[n.ID] = adj[n.ID]
// 	}
// 	// add edges
// 	for _, e := range g.Graph.Edges {
// 		adj[e.Source] = append(adj[e.Source], e.Target)
// 	}
// 	return adj, nil
// }
