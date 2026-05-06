package main

import (
	"bytes"
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/ollama"
	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"

	"github.com/siber/go-genkit-rag-chatbot/internal/vectorstore"
)

func main() {
	ctx := context.Background()

	filePath := flag.String("file", "/Users/mac/projects/siber/go-projects/go-genkit-RAG-chatbot/assets/SDMH24.pdf", "file path")
	flag.Parse()

	if *filePath == "" {
		log.Fatal("Please provide a file path using -file")
	}

	var text string
	if strings.HasSuffix(strings.ToLower(*filePath), ".pdf") {
		f, r, err := pdf.Open(*filePath)
		if err != nil {
			log.Fatalf("Failed to open pdf: %v", err)
		}
		defer f.Close()

		var buf bytes.Buffer
		b, err := r.GetPlainText()
		if err != nil {
			log.Fatalf("Failed to extract pdf text: %v", err)
		}
		buf.ReadFrom(b)
		text = buf.String()
	} else {
		content, err := os.ReadFile(*filePath)
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}
		text = string(content)
	}

	chunks := vectorstore.ChunkText(text, 200, 50)
	log.Printf("Created %d chunks from the document", len(chunks))

	plugin := &ollama.Ollama{
		ServerAddress: "http://localhost:11434",
	}
	g := genkit.Init(ctx, genkit.WithPlugins(plugin))

	embedder := plugin.DefineEmbedder(g, "http://localhost:11434", "nomic-embed-text", nil)

	qClient := vectorstore.NewSimpleQdrantClient("localhost", 6333)

	collectionName := "manuals"
	err := qClient.InitCollection(ctx, collectionName, 768)
	if err != nil {
		log.Printf("Warning or error during collection init: %v", err)
	}

	var points []vectorstore.Point

	for i, chunk := range chunks {
		res, err := genkit.Embed(ctx, g,
			ai.WithEmbedder(embedder),
			ai.WithTextDocs(chunk),
		)
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
		err := qClient.Upsert(ctx, collectionName, points)
		if err != nil {
			log.Fatalf("Failed to upsert points: %v", err)
		}
	}

	log.Println("Ingestion completed successfully!")
}
