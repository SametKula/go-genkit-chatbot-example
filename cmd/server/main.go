package main

import (
	"context"
	"log"
	"net/http"

	"github.com/siber/go-genkit-rag-chatbot/internal/agent"
	"github.com/siber/go-genkit-rag-chatbot/internal/memory"
	"github.com/siber/go-genkit-rag-chatbot/internal/vectorstore"
	"github.com/siber/go-genkit-rag-chatbot/internal/websocket"
)

func main() {
	ctx := context.Background()

	store := memory.NewInMemorySessionStore()

	qClient := vectorstore.NewSimpleQdrantClient("localhost", 6333)

	botAgent, err := agent.NewAgent(ctx, qClient)
	if err != nil {
		log.Fatalf("Failed to initialize agent: %v", err)
	}

	wsServer := websocket.NewServer(botAgent, store)

	http.HandleFunc("/ws", wsServer.HandleWebSocket)

	log.Println("Server is listening on port :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
