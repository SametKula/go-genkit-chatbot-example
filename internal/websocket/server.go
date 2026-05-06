package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/siber/go-genkit-rag-chatbot/internal/agent"
	"github.com/siber/go-genkit-rag-chatbot/internal/memory"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all for simplicity
	},
}

type Server struct {
	agent *agent.Agent
	store memory.SessionStore
}

func NewServer(agent *agent.Agent, store memory.SessionStore) *Server {
	return &Server{
		agent: agent,
		store: store,
	}
}

type IncomingMessage struct {
	SessionID string `json:"session_id"`
	Content   string `json:"content"`
}

type OutgoingMessage struct {
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Websocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		if messageType != websocket.TextMessage {
			continue
		}

		var incoming IncomingMessage
		if err := json.Unmarshal(p, &incoming); err != nil {
			log.Printf("JSON unmarshal error: %v", err)
			s.sendError(conn, "Invalid message format. Expected JSON.")
			continue
		}

		if incoming.SessionID == "" || incoming.Content == "" {
			s.sendError(conn, "session_id and content are required")
			continue
		}

		ctx := context.Background()
		
		history, err := s.store.GetHistory(ctx, incoming.SessionID)
		if err != nil {
			log.Printf("History retrieval error: %v", err)
		}

		reply, err := s.agent.Chat(ctx, incoming.Content, history)
		if err != nil {
			log.Printf("Agent chat error: %v", err)
			s.sendError(conn, "Failed to generate response")
			continue
		}

		err = s.store.SaveMessage(ctx, incoming.SessionID, memory.Message{Role: "user", Content: incoming.Content})
		if err != nil {
			log.Printf("Failed to save user message: %v", err)
		}

		err = s.store.SaveMessage(ctx, incoming.SessionID, memory.Message{Role: "model", Content: reply})
		if err != nil {
			log.Printf("Failed to save model message: %v", err)
		}

		out := OutgoingMessage{Content: reply}
		outBytes, _ := json.Marshal(out)
		conn.WriteMessage(websocket.TextMessage, outBytes)
	}
}

func (s *Server) sendError(conn *websocket.Conn, errMsg string) {
	out := OutgoingMessage{Error: errMsg}
	outBytes, _ := json.Marshal(out)
	conn.WriteMessage(websocket.TextMessage, outBytes)
}
