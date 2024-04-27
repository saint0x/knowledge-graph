package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// KnowledgeGraph represents the knowledge graph
type KnowledgeGraph struct {
	Nodes    map[int64]*Node
	Edges    map[int64]*Edge
	Vertices map[int64]*Vertex
}

var (
	nodeIDCounter   int64 = 1
	edgeIDCounter   int64 = 1
	vertexIDCounter int64 = 1
)

// Node represents a node in the knowledge graph
type Node struct {
	ID       int64
	Text     string
	Concepts []string
}

// Edge represents an edge in the knowledge graph
type Edge struct {
	ID       int64
	SourceID int64
	TargetID int64
	Weight   float64
}

// Vertex represents a vertex in the knowledge graph
type Vertex struct {
	ID       int64
	NodeID   int64
	TargetID int64
	Concept  string
}

// NewKnowledgeGraph creates a new instance of KnowledgeGraph
func NewKnowledgeGraph() *KnowledgeGraph {
	return &KnowledgeGraph{
		Nodes:    make(map[int64]*Node),
		Edges:    make(map[int64]*Edge),
		Vertices: make(map[int64]*Vertex),
	}
}

func main() {
	// Retrieve OpenAI API key from environment variables
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// Try retrieving from secrets if environment variable is not set
		apiKey = os.Getenv("MY_SECRET")
		if apiKey == "" {
			log.Fatal("OpenAI API key not found. Please set the OPENAI_API_KEY or MY_SECRET environment variable with your API key.")
		}
	}
	log.Println("AI Client Initialized!")

	// Initialize the graph
	graph := NewKnowledgeGraph()

	// Check if a graph file exists
	graphFilePath := "knowledge_graph.txt"
	_, err := os.Stat(graphFilePath)
	if os.IsNotExist(err) {
		// Create a new knowledge graph if the file doesn't exist
		log.Println("Knowledge Graph Created")
		if err := SaveGraph(graphFilePath, graph); err != nil {
			log.Fatalf("Failed to save knowledge graph: %v", err)
		}
	} else if err == nil {
		// Load the existing knowledge graph
		log.Println("Knowledge Graph Accessed!")
		graph, err = LoadGraph(graphFilePath)
		if err != nil {
			log.Fatalf("Failed to load knowledge graph: %v", err)
		}
	} else {
		log.Fatalf("Error checking graph file: %v", err)
	}

	// Get user input for note text
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("Please enter the text for your note (or type 'exit' to finish):")
		scanner.Scan()
		noteText := scanner.Text()
		if noteText == "exit" {
			break
		}

		// Extract concepts from the note text
		concepts, err := ExtractConcepts(noteText, apiKey)
		if err != nil {
			log.Fatalf("Failed to extract concepts from note: %v", err)
		}

		// Build or update knowledge graph with the provided note text and concepts
		if err := BuildOrUpdateKnowledgeGraph(graph, noteText, concepts); err != nil {
			log.Fatalf("Failed to build or update knowledge graph: %v", err)
		}

		// Save the updated graph
		if err := SaveGraph(graphFilePath, graph); err != nil {
			log.Fatalf("Failed to save knowledge graph: %v", err)
		}

		fmt.Println("Note Added!")
		fmt.Println("Graph Updated!")
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error while reading user input: %v", err)
	}
}

// ExtractConcepts extracts concepts from a text using OpenAI's GPT
func ExtractConcepts(text string, apiKey string) ([]string, error) {
	// Initialize OpenAI client with API key
	client := openai.NewClient(apiKey)

	// Prepare system prompt
	prompt := "You are an AI assistant that is an expert at understanding context. You will take the provided text:\n" + text + "\nConcepts and extract the main concepts for our knowledge graph based on sentiment. Our aim is to allow the nodes, edges, and vertices to be parsed for added context, so keep that in mind. Do not respond with the label, just the words."

	// Extract concepts using OpenAI GPT-3
	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to extract concepts: %v", err)
	}

	conceptsText := resp.Choices[0].Message.Content
	conceptsText = strings.TrimPrefix(conceptsText, "Concepts:")
	conceptsText = strings.TrimSpace(conceptsText)

	concepts := strings.Split(conceptsText, ", ")
	return concepts, nil
}

// BuildOrUpdateKnowledgeGraph builds the knowledge graph with the provided note text and concepts, or updates an existing graph
func BuildOrUpdateKnowledgeGraph(graph *KnowledgeGraph, noteText string, concepts []string) error {
	// Create nodes for the note
	node := Node{
		ID:       generateNodeID(),
		Text:     noteText,
		Concepts: concepts,
	}
	graph.Nodes[node.ID] = &node

	// Create edges and vertices based on the relationships between nodes
	for _, existingNode := range graph.Nodes {
		if existingNode.ID != node.ID {
			// Calculate edge weight based on concept similarity
			weight := CalculateWeight(node.Concepts, existingNode.Concepts)
			if weight > 0 {
				// Create an edge between the nodes
				edge := Edge{
					ID:       generateEdgeID(),
					SourceID: node.ID,
					TargetID: existingNode.ID,
					Weight:   weight,
				}
				graph.Edges[edge.ID] = &edge

				// Create vertices for the concepts shared by the nodes
				for _, concept := range node.Concepts {
					if Contains(existingNode.Concepts, concept) {
						vertex := Vertex{
							ID:       generateVertexID(),
							NodeID:   node.ID,
							TargetID: existingNode.ID,
							Concept:  concept,
						}
						graph.Vertices[vertex.ID] = &vertex
					}
				}
			}
		}
	}

	return nil
}

// SaveGraph saves the knowledge graph to storage
func SaveGraph(filePath string, graph *KnowledgeGraph) error {
	// Create or open the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Write concepts to the file
	_, err = fmt.Fprintf(file, "Concepts: %s\n", strings.Join(getAllConcepts(graph), ", "))
	if err != nil {
		return fmt.Errorf("failed to write concepts: %v", err)
	}

	// Write nodes to the file
	for id, node := range graph.Nodes {
		_, err := fmt.Fprintf(file, "Node %d: %s\n", id, node.Text)
		if err != nil {
			return fmt.Errorf("failed to write node: %v", err)
		}
	}

	// Write edges to the file
	for id, edge := range graph.Edges {
		_, err := fmt.Fprintf(file, "Edge %d: SourceID=%d, TargetID=%d, Weight=%f\n", id, edge.SourceID, edge.TargetID, edge.Weight)
		if err != nil {
			return fmt.Errorf("failed to write edge: %v", err)
		}
	}

	// Write vertices to the file
	for id, vertex := range graph.Vertices {
		_, err := fmt.Fprintf(file, "Vertex %d: NodeID=%d, TargetID=%d, Concept=%s\n", id, vertex.NodeID, vertex.TargetID, vertex.Concept)
		if err != nil {
			return fmt.Errorf("failed to write vertex: %v", err)
		}
	}

	return nil
}

// LoadGraph loads the knowledge graph from storage
func LoadGraph(filePath string) (*KnowledgeGraph, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Initialize maps for nodes, edges, and vertices
	graph := &KnowledgeGraph{
		Nodes:    make(map[int64]*Node),
		Edges:    make(map[int64]*Edge),
		Vertices: make(map[int64]*Vertex),
	}

	// Create a scanner to read from the file
	scanner := bufio.NewScanner(file)

	// Read each line and parse graph elements
	for scanner.Scan() {
		line := scanner.Text()

		// Parse node
		if strings.HasPrefix(line, "Node") {
			var node Node
			if _, err := fmt.Sscanf(line, "Node %d:", &node.ID); err != nil {
				return nil, fmt.Errorf("failed to parse node: %v", err)
			}
			node.Text = strings.TrimSpace(strings.TrimPrefix(line, fmt.Sprintf("Node %d:", node.ID)))
			graph.Nodes[node.ID] = &node
		}

		// Parse edge
		if strings.HasPrefix(line, "Edge") {
			var edge Edge
			if _, err := fmt.Sscanf(line, "Edge %d: SourceID=%d, TargetID=%d, Weight=%f",
				&edge.ID, &edge.SourceID, &edge.TargetID, &edge.Weight); err != nil {
				return nil, fmt.Errorf("failed to parse edge: %v", err)
			}
			graph.Edges[edge.ID] = &edge
		}

		// Parse vertex
		if strings.HasPrefix(line, "Vertex") {
			var vertex Vertex
			if _, err := fmt.Sscanf(line, "Vertex %d: NodeID=%d, TargetID=%d, Concept=%s",
				&vertex.ID, &vertex.NodeID, &vertex.TargetID, &vertex.Concept); err != nil {
				return nil, fmt.Errorf("failed to parse vertex: %v", err)
			}
			graph.Vertices[vertex.ID] = &vertex
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error while scanning file: %v", err)
	}

	return graph, nil
}

// Helper functions for generating unique IDs
func generateNodeID() int64 {
	nodeIDCounter++
	return nodeIDCounter
}

func generateEdgeID() int64 {
	edgeIDCounter++
	return edgeIDCounter
}

func generateVertexID() int64 {
	vertexIDCounter++
	return vertexIDCounter
}

// CalculateWeight calculates the weight between two sets of concepts based on Jaccard similarity
func CalculateWeight(concepts1, concepts2 []string) float64 {
	// Convert concept slices to sets for easier comparison
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, concept := range concepts1 {
		set1[concept] = true
	}

	for _, concept := range concepts2 {
		set2[concept] = true
	}

	// Calculate Jaccard similarity
	intersection := 0
	for concept := range set1 {
		if set2[concept] {
			intersection++
		}
	}

	union := len(set1) + len(set2) - intersection

	// Prevent division by zero
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// Contains checks if a string exists in a slice
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ExtractNodesEdgesVertices uses AI to extract nodes, edges, and vertices based on the provided concepts
func ExtractNodesEdgesVertices(graph *KnowledgeGraph, client *openai.Client) error {
	// Prepare system prompt
	prompt := "You are an AI assistant that is an expert at understanding context. You will take the provided concepts and extract various nodes, edges, and vertices for our knowledge graph based on sentiment. Our aim is to allow the nodes, edges, and vertices to be parsed for added context, so keep that in mind. Respond with the nodes and any potential edges or vertices you want to add to the knowledge graph. Do not respond with the words edge, nodes, vertices, just the words themselves."

	// Get all concepts as a single string
	conceptsString := strings.Join(getAllConcepts(graph), ", ")

	// Extract nodes, edges, and vertices
	_, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt + "\nConcepts: " + conceptsString,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to extract nodes, edges, and vertices: %v", err)
	}

	// Since we're not using the response here, we can discard it

	// Assuming the extraction logic returns structured data, we'll create a dummy graph for now
	return nil
}

// getAllConcepts retrieves all unique concepts from the knowledge graph
func getAllConcepts(graph *KnowledgeGraph) []string {
	conceptsMap := make(map[string]bool)
	for _, node := range graph.Nodes {
		for _, concept := range node.Concepts {
			conceptsMap[concept] = true
		}
	}
	concepts := make([]string, 0, len(conceptsMap))
	for concept := range conceptsMap {
		concepts = append(concepts, concept)
	}
	return concepts
}
