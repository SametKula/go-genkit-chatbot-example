# System Architecture

## Core Technologies
- **Language**: Go
- **Framework**: Genkit (RAG ve Agent Akışları)
- **Vector DB**: Qdrant
- **LLM**: Qwen3.5:14b (Ollama veya benzeri bir yerel LLM sunucusu üzerinden)

## Project Structure
- `cmd/ingest/main.go`: Veri yükleme aracı. (PDF/Text okuma, chunking, embedding, DB kaydı).
- `cmd/server/main.go`: Websocket ve Chatbot sunucusu. Genkit akışlarını çalıştırır.
- `internal/agent`: Genkit prompt, tool (`search_db`) ve flow tanımlamaları.
- `internal/vectorstore`: Qdrant bağlantısı, arama ve kayıt işlemleri.
- `internal/memory`: `SessionStore` interface'i ve in-memory (hash map) implementasyonu.
- `internal/websocket`: Websocket sunucu yönetimi, odalar/session yapıları.

## Architecture Decisions
- Clean Architecture (İş, DB, Ağ katmanlarının ayrışması)
- Dependency Injection (SessionStore gibi interface kullanımları)
- Chunking yapısında overlap kullanılması
