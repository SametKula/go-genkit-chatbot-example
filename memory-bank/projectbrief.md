# Project Brief

## Project Overview
Agentic RAG tabanlı, sorun giderme odaklı bir Chatbot projesi geliştirilecek. Proje Go, Genkit, Qdrant (Vektör DB) ve Qwen3.5:14b LLM kullanılarak inşa edilecek.
İki ana bileşenden oluşacak: Ingestion aracı (pdf/metin okuma, chunking, embedding, qdrant'a yazma) ve Websocket Chatbot Sunucusu (kullanıcı etkileşimi, Genkit RAG agent akışı).

## Requirements
- Modüler mimari (Clean Architecture). İş mantığı, veritabanı, websocket (ağ) ayrılacak.
- Genkit akışlarını ve Qdrant entegrasyonunu kullanacak.
- Hafızayı (Memory) in-memory olarak yönetecek ama interface odaklı yapıyla (Dependency Injection) Redis'e vs. genişletilebilecek.
- LLM'e araç (tool) sağlanacak: `search_db`.
- Git commitleri anlamlı noktalarda atılacak.

## Goals
Sorun giderme ve rehberlik yapabilen, kılavuzları referans alarak Agent gibi düşünebilen bir chatbot oluşturmak.
