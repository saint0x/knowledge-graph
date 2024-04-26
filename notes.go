// pkg/knowledgegraph/notes.go

package notes

import (
  "context"
  "log"
  "strings"

  "github.com/sashabaranov/go-openai"
  "sqlite"
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
    notes = append(notes, Note{
      ID:   vn.ID,
      Text: vn.Transcription,
      // Concepts will be populated later
    })
  }
  return notes, nil
}

// GenerateSummary generates a summary of the note using OpenAI GPT-3
func GenerateSummary(text string) (string, error) {
  // Prepare system prompt
  prompt := "Generate a summary of the following note:\n\"" + text + "\""

  // Generate summary using OpenAI GPT-3
  resp, err := openai.NewClient("your-token-here").CreateCompletion(context.Background(), openai.CompletionRequest{
    Model:     openai.GPT3Dot5Turbo,
    Prompt:    prompt,
    MaxTokens: 100,
  })
  if err != nil {
    log.Printf("Failed to generate summary: %v", err)
    return "", err
  }

  summary := resp.Choices[0].Text
  return summary, nil
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
    concepts, err := ExtractConcepts(note.Text)
    if err != nil {
      log.Printf("Failed to extract concepts for note %d: %v", note.ID, err)
      continue
    }

    err = UpdateNoteConcepts(note.ID, concepts)
    if err != nil {
      log.Printf("Failed to update concepts for note %d: %v", note.ID, err)
      continue
    }
  }

  return nil
}
