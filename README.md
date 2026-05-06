# Agentic RAG Chatbot

Bu proje, Go dili ve Genkit framework'ü kullanılarak geliştirilmiş, teknik dökümantasyonlar üzerinden akıllı destek sağlayan bir Agentic RAG (Retrieval-Augmented Generation) chatbot uygulamasıdır.

## Mimari Özellikler

- Agentic Akış: Model, kullanıcı sorularını yanıtlamak için dökümantasyon araması yapıp yapmayacağına ve yanıt stratejisine (bilgi verme veya adım adım rehberlik) kendi karar verir.
- Hybrid Search ve Re-ranking: Vektör tabanlı benzerlik aramasını, anahtar kelime eşleşmesi ile birleştiren bir re-ranking algoritması kullanır. Bu sayede teknik terimlerin ve hata kodlarının isabet oranı artırılmıştır.
- create_ticket Aracı: Sistemde çözüm bulunamadığında otomatik olarak destek talebi oluşturma yeteneğine sahiptir.
- Turuncu Temalı Chat Widget: Web arayüzünde sağ alt köşede konumlanan, kullanıcı dostu ve modern bir sohbet arayüzü sunar.

## Kurulum ve Çalıştırma Adımları

// Vektör veritabanını Docker üzerinde ayağa kaldırmak için aşağıdaki komutu kullanın.
```bash
docker run -p 6333:6333 qdrant/qdrant
```

// Gerekli LLM ve Embedding modellerini Ollama üzerinden indirin.
```bash
ollama pull llama3.1:8b
ollama pull nomic-embed-text
```

// Teknik dökümanları analiz edip veritabanına yüklemek için ingestion aracını çalıştırın.
```bash
go run cmd/ingest/main.go -file assets/kullanma-kilavuzu.txt
```

// Chatbot sunucusunu başlatmak için server uygulamasını çalıştırın.
```bash
go run cmd/server/main.go
```

## Proje Yapısı

- cmd/ingest: Veri yükleme ve vektörize etme işlemleri.
- cmd/server: Websocket tabanlı ana sunucu.
- internal/agent: Agent tanımları, prompt yönetimi ve araçlar.
- internal/vectorstore: Qdrant entegrasyonu ve hibrit arama mantığı.
- internal/websocket: Ağ iletişimi ve oturum yönetimi.
- client: Web tabanlı Teknik Asistan arayüzü.
