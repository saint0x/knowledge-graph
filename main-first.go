package main

import (
  "bufio"
	"fmt"
	"log"
	"os"

)

func main() {
	// Ask for user inputs
	fmt.Println("Welcome to the Knowledge Graph Builder!")
	fmt.Println("Please provide some information to build your knowledge graph.")

	// Get text input for building nodes
	fmt.Print("Enter a note (node) for the knowledge graph: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	note := scanner.Text()

	// Add note to the database
	noteID, err := knowledgegraph.AddNoteToDB(1, note) // Replace 1 with actual user ID
	if err != nil {
		log.Fatalf("Failed to add note to database: %v", err)
	}

	// Extract concepts from the note
	fmt.Println("Extracting concepts from the note...")
	concepts, err := knowledgegraph.ExtractConcepts(note)
	if err != nil {
		log.Fatalf("Failed to extract concepts: %v", err)
	}

	// Update note with extracted concepts
	fmt.Println("Updating note with extracted concepts...")
	err = knowledgegraph.UpdateNoteConcepts(noteID, concepts)
	if err != nil {
		log.Fatalf("Failed to update note with concepts: %v", err)
	}

	// Update notes with concepts
	fmt.Println("Updating all notes with extracted concepts...")
	err = knowledgegraph.UpdateNotesWithConcepts()
	if err != nil {
		log.Fatalf("Failed to update notes with concepts: %v", err)
	}

	// Generate summary of the note
	fmt.Println("Generating summary of the note...")
	summary, err := knowledgegraph.GenerateSummary(note)
	if err != nil {
		log.Fatalf("Failed to generate summary: %v", err)
	}

	// Print the summary
	fmt.Println("Summary of the note:")
	fmt.Println(summary)
}
