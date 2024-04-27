package main

import (
  "context"
  "log"
  "strings"

  "github.com/sashabaranov/go-openai"

)

// Note represents a note in the knowledge graph
type Note struct {
  ID        int64
  Text      string
  Concepts  []string
}

// AddNoteToDB adds a new note to the SQLite database
func AddNoteToDB(userID int, text string) (int64, error) {
  // Insert the note into the database
  noteID, err := sqlite.InsertVoiceNote(userID, "", text) // Replace with actual user ID and file path
  if err != nil {
    log.Printf("Failed to insert voice note: %v", err)
    return 0, err
  }
  return noteID, nil
}

// GetNotesFromDB retrieves all notes from the SQLite database
func GetNotesFromDB() ([]Note, error) {
  // Retrieve all voice notes from the database
  voiceNotes, err := sqlite.GetAllVoiceNotes()
  if err != nil {
    log.Printf("Failed to retrieve voice notes: %v", err)
    return nil, err
  }

  // Convert voice notes to knowledge graph notes
  var notes []Note
  for _, vn := range voiceNotes {
    // Extract concepts from the summary
    concepts, err := ExtractConcepts(vn.Summary)
    if err != nil {
      log.Printf("Failed to extract concepts from summary for note %d: %v", vn.ID, err)
      continue
    }

    notes = append(notes, Note{
      ID:       vn.ID,
      Text:     vn.Summary, // Storing summary as text for now
      Concepts: concepts,
    })
  }
  return notes, nil
}

// ExtractConcepts extracts concepts from a text using OpenAI's GPT
func ExtractConcepts(text string) ([]string, error) {
  // Prepare system prompt
  prompt := "Extract the main concepts from the following text:\n" + text + "\nConcepts:"

  // Extract concepts using OpenAI GPT-3
  resp, err := openai.NewClient("your-token-here").CreateCompletion(context.Background(), openai.CompletionRequest{
    Model:     openai.GPT3Dot5Turbo,
    Prompt:    prompt,
    MaxTokens: 50,
  })
  if err != nil {
    log.Printf("Failed to extract concepts: %v", err)
    return nil, err
  }

  conceptsText := resp.Choices[0].Text
  conceptsText = strings.TrimPrefix(conceptsText, "Concepts:")
  conceptsText = strings.TrimSpace(conceptsText)

  concepts := strings.Split(conceptsText, ", ")
  return concepts, nil
}

// UpdateNoteConcepts updates the concepts of a note in the SQLite database
func UpdateNoteConcepts(noteID int64, concepts []string) error {
  // Update the concepts of the note in the database
  for _, concept := range concepts {
    topicID, err := sqlite.InsertTopic(concept)
    if err != nil {
      log.Printf("Failed to insert topic: %v", err)
      continue
    }

    // Insert connection between note and topic into database
    err = sqlite.InsertVoiceNoteTopic(noteID, topicID)
    if err != nil {
      log.Printf("Failed to insert voice note topic: %v", err)
      continue
    }
  }
  return nil
}

// UpdateNotesWithConcepts updates notes in the database with extracted concepts
func UpdateNotesWithConcepts() error {
  notes, err := GetNotesFromDB()
  if err != nil {
    log.Printf("Failed to retrieve notes: %v", err)
    return err
  }

  for _, note := range notes {
    err := UpdateNoteConcepts(note.ID, note.Concepts)
    if err != nil {
      log.Printf("Failed to update concepts for note %d: %v", note.ID, err)
      continue
    }
  }

  return nil

}

// ConvertVoiceNotesToKnowledgeGraph converts voice note summaries to knowledge graph nodes, edges, and vertices
func ConvertVoiceNotesToKnowledgeGraph() error {
  // Retrieve all voice note summaries from the database
  voiceNoteSummaries, err := sqlite.GetAllVoiceNoteSummaries()
  if err != nil {
    log.Printf("Failed to retrieve voice note summaries: %v", err)
    return err
  }

  // Initialize slices to store nodes, edges, and vertices
  var nodes []knowledgegraph.Node
  var edges []knowledgegraph.Edge
  var vertices []knowledgegraph.Vertex

  // Convert voice note summaries to knowledge graph
  for _, summary := range voiceNoteSummaries {
    // Extract concepts from the summary
    concepts, err := ExtractConcepts(summary)
    if err != nil {
      log.Printf("Failed to extract concepts for summary: %v", err)
      continue
    }

    // Create a node in the knowledge graph for the summary text
    node := knowledgegraph.Node{
      Text:     summary,
      Concepts: concepts,
    }

    // Store the node for appending to the graph
    nodes = append(nodes, node)

    // Create edges and vertices based on the relationships between nodes
    for _, n := range nodes {
      // Skip creating edges and vertices for the same node
      if n.ID == node.ID {
        continue
      }

      // Calculate edge weight based on concept similarity
      weight := calculateWeight(node.Concepts, n.Concepts)
      if weight > 0 {
        // Create an edge between the nodes
        edge := knowledgegraph.Edge{
          SourceID: node.ID,
          TargetID: n.ID,
          Weight:   weight,
        }
        edges = append(edges, edge)

        // Create vertices for the concepts shared by the nodes
        for _, concept := range node.Concepts {
          if contains(n.Concepts, concept) {
            vertex := knowledgegraph.Vertex{
              NodeID:   node.ID,
              TargetID: n.ID,
              Concept:  concept,
            }
            vertices = append(vertices, vertex)
          }
        }
      }
    }

    // Persist the node in the knowledge graph database
    err = PersistNode(node)
    if err != nil {
      log.Printf("Failed to persist node for summary: %v", err)
      continue
    }
  }

  // Persist the collected nodes, edges, and vertices in the knowledge graph database
  err = PersistGraphData(nodes, edges, vertices)
  if err != nil {
    log.Printf("Failed to persist graph data: %v", err)
    return err
  }

  return nil
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

// contains checks if a string slice contains a specific string
func contains(slice []string, str string) bool {
  for _, s := range slice {
    if s == str {
      return true
    }
  }
  return false
}
