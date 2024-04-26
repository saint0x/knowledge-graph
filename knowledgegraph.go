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
func BuildGraph() (Graph, error) {
	var graph Graph

	// Update notes in the database with extracted concepts
	err := UpdateNotesWithConcepts()
	if err != nil {
		log.Printf("Failed to update notes with concepts: %v", err)
		return graph, err
	}

	// Retrieve all notes from the database
	notes, err := GetNotesFromDB()
	if err != nil {
		log.Printf("Failed to retrieve notes: %v", err)
		return graph, err
	}

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

// GenerateInsight generates insights from the knowledge graph
func GenerateInsight(graph Graph) (string, error) {
	// TODO: Implement insight generation based on the knowledge graph
	// For now, return a placeholder insight
	return "Placeholder insight from knowledge graph", nil
}
