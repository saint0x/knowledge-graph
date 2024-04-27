package main

import (
	"log"
	"strings"
)

// Node represents a node in the knowledge graph
type Node struct {
	ID       int64
	Text     string
	Concepts []string
}

// Edge represents an edge in the knowledge graph
type Edge struct {
	SourceID int64
	TargetID int64
	Weight   float64
}

// Graph represents the knowledge graph
type Graph struct {
	Nodes []Node
	Edges []Edge
}

// BuildGraph builds the knowledge graph from notes
func BuildGraph(notes []Note) (Graph, error) {
	var graph Graph

	// Create nodes for each note
	for _, note := range notes {
		graph.Nodes = append(graph.Nodes, Node{
			ID:       note.ID,
			Text:     note.Text,
			Concepts: note.Concepts,
		})
	}

	// Create edges based on concepts similarity
	for i := range graph.Nodes {
		for j := range graph.Nodes {
			if i != j {
				weight := calculateWeight(graph.Nodes[i].Concepts, graph.Nodes[j].Concepts)
				if weight > 0 {
					graph.Edges = append(graph.Edges, Edge{
						SourceID: graph.Nodes[i].ID,
						TargetID: graph.Nodes[j].ID,
						Weight:   weight,
					})
				}
			}
		}
	}

	return graph, nil
}

// calculateWeight calculates the weight between two sets of concepts
func calculateWeight(concepts1, concepts2 []string) float64 {
	var commonConcepts int
	for _, concept1 := range concepts1 {
		for _, concept2 := range concepts2 {
			if strings.ToLower(concept1) == strings.ToLower(concept2) {
				commonConcepts++
			}
		}
	}

	if len(concepts1)+len(concepts2)-commonConcepts == 0 {
		return 0
	}

	return float64(commonConcepts) / float64(len(concepts1)+len(concepts2)-commonConcepts)
}

// PersistNode persists a node in the knowledge graph database
func PersistNode(node knowledgegraph.Node) error {
	// Insert the node into the database
	err := sqlite.InsertNode(node.ID, node.Text, strings.Join(node.Concepts, ", "))
	if err != nil {
		log.Printf("Failed to persist node: %v", err)
		return err
	}
	return nil
}

// PersistEdge persists an edge in the knowledge graph database
func PersistEdge(edge knowledgegraph.Edge) error {
	// Insert the edge into the database
	err := sqlite.InsertEdge(edge.SourceID, edge.TargetID, edge.Weight)
	if err != nil {
		log.Printf("Failed to persist edge: %v", err)
		return err
	}
	return nil
}

// PersistVertex persists a vertex in the knowledge graph database
func PersistVertex(vertex knowledgegraph.Vertex) error {
	// Insert the vertex into the database
	err := sqlite.InsertVertex(vertex.NodeID, vertex.TargetID, vertex.Concept)
	if err != nil {
		log.Printf("Failed to persist vertex: %v", err)
		return err
	}
	return nil
}

// PersistGraphData persists the collected nodes, edges, and vertices in the knowledge graph database
func PersistGraphData(nodes []knowledgegraph.Node, edges []knowledgegraph.Edge, vertices []knowledgegraph.Vertex) error {
	// Persist nodes
	for _, node := range nodes {
		err := PersistNode(node)
		if err != nil {
			return err
		}
	}

	// Persist edges
	for _, edge := range edges {
		err := PersistEdge(edge)
		if err != nil {
			return err
		}
	}

	// Persist vertices
	for _, vertex := range vertices {
		err := PersistVertex(vertex)
		if err != nil {
			return err
		}
	}

	return nil
}
