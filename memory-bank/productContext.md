# Product Context

## Problem
Kullanıcıların kılavuzları ve dokümantasyonları manuel taraması yerine, Agentic RAG mimarisiyle etkileşime girip doğrudan sorunlarını çözebilecekleri akıllı bir sisteme ihtiyaç duyulmaktadır.

## Solution
Go, Genkit ve Qdrant kullanılarak geliştirilen Agentic Chatbot.
- Ingestion servisi kılavuzları vektörize eder.
- Server, Websocket üzerinden iletişim kurarak kullanıcının sorunlarını `search_db` aracı ile Qdrant üzerinden araştırır ve LLM ile cevaplar.
- LLM model olarak Qwen3.5:14b kullanır. Model, bir adımda sorunu çözemezse sonraki adımlara geçmesi için prompt ile yönlendirilir.

## Experience
Kullanıcı Websocket üzerinden bir session başlatır. Sorunu iletir. Bot, ilgili kılavuzları araştırıp çözüm önerilerini adım adım sunar.
