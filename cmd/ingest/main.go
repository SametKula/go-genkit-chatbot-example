package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/plugins/ollama"
	"github.com/google/uuid"

	"github.com/siber/go-genkit-rag-chatbot/internal/vectorstore"
)

func main() {
	ctx := context.Background()

	filePath := flag.String("file", "", "Path to the text file to ingest")
	flag.Parse()

	if *filePath == "" {
		log.Fatal("Please provide a file path using -file")
	}

	content, err := os.ReadFile(*filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	text := string(content)

	chunks := vectorstore.ChunkText(text, 200, 50)
	log.Printf("Created %d chunks from the document", len(chunks))

	err = ollama.Init(ctx, &ollama.Config{
		ServerAddress: "http://localhost:11434",
	})
	if err != nil {
		log.Fatalf("Failed to init ollama: %v", err)
	}

	embedder := ollama.Embedder("nomic-embed-text")

	qClient := vectorstore.NewSimpleQdrantClient("localhost", 6333)
	
	collectionName := "manuals"
	err = qClient.InitCollection(ctx, collectionName, 768)
	if err != nil {
		log.Printf("Warning or error during collection init: %v", err)
	}

	var points []vectorstore.Point

	for i, chunk := range chunks {
		req := &ai.EmbedRequest{
			Documents: []*ai.Document{
				ai.DocumentFromText(chunk, nil),
			},
		}
		
		res, err := embedder.Embed(ctx, req)
		if err != nil {
			log.Fatalf("Failed to embed chunk %d: %v", i, err)
		}
		
		if len(res.Embeddings) == 0 {
			continue
		}

		vec := res.Embeddings[0].Embedding
		points = append(points, vectorstore.Point{
			ID:     uuid.New().String(),
			Vector: vec,
			Payload: map[string]interface{}{
				"text":   chunk,
				"source": *filePath,
			},
		})
	}

	if len(points) > 0 {
		err = qClient.Upsert(ctx, collectionName, points)
		if err != nil {
			log.Fatalf("Failed to upsert points: %v", err)
		}
	}

	log.Println("Ingestion completed successfully!")
}
