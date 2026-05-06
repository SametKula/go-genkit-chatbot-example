package agent

import (
	"context"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/plugins/ollama"
	"github.com/siber/go-genkit-rag-chatbot/internal/memory"
	"github.com/siber/go-genkit-rag-chatbot/internal/vectorstore"
)

type Agent struct {
	qClient      *vectorstore.SimpleQdrantClient
	embedder     ai.Embedder
	model        ai.Model
	searchDbTool ai.Tool
}

type SearchDbInput struct {
	Query string `json:"query"`
}

func NewAgent(ctx context.Context, qClient *vectorstore.SimpleQdrantClient) (*Agent, error) {
	err := ollama.Init(ctx, &ollama.Config{
		ServerAddress: "http://localhost:11434",
	})
	if err != nil {
		return nil, err
	}

	embedder := ollama.Embedder("nomic-embed-text")
	model := ollama.Model("qwen3.5:14b")

	searchDbTool := ai.DefineTool(
		"search_db",
		"Vektör veritabanında (Qdrant) arama yapıp bağlamı ve kılavuzları getiren araç. Kullanıcının sorununu çözmek için ilgili dökümantasyonları bu araçla ara.",
		func(ctx context.Context, input *SearchDbInput) (string, error) {
			req := &ai.EmbedRequest{
				Documents: []*ai.Document{
					ai.DocumentFromText(input.Query, nil),
				},
			}
			res, err := embedder.Embed(ctx, req)
			if err != nil {
				return "", err
			}

			if len(res.Embeddings) == 0 {
				return "No embedding generated for query", nil
			}

			vec := res.Embeddings[0].Embedding

			texts, err := qClient.Search(ctx, "manuals", vec, 3)
			if err != nil {
				return "", err
			}

			if len(texts) == 0 {
				return "No relevant information found in manuals.", nil
			}

			return "Bulunan kılavuz içerikleri: \n\n" + strings.Join(texts, "\n\n---\n\n"), nil
		},
	)

	return &Agent{
		qClient:      qClient,
		embedder:     embedder,
		model:        model,
		searchDbTool: searchDbTool,
	}, nil
}

func (a *Agent) Chat(ctx context.Context, userInput string, history []memory.Message) (string, error) {
	systemPrompt := `Sen uzman bir sorun giderme asistanısın. Agentic RAG tabanlı bir Chatbotsun.
Adım adım ilerlemen ve sorunları çözmen gerekiyor.
ÖNEMLİ KURALLAR:
1. Kullanıcının sorununun çözümü için her zaman 'search_db' aracını kullanarak bağlam araştır.
2. Bir sorunu çözerken, kullanıcının o adımı uygulayıp uygulamadığını bekle.
3. Bir adımı uygulamadan diğerine GEÇME. İlk adım başarılıysa sonrasında ikinci adıma geçeceğini belirt.
4. Çözümleri her zaman veritabanından aldığın bilgilere (search_db sonuçlarına) dayandır. Uydurma bilgi verme.`

	messages := []*ai.Message{
		{
			Role: ai.RoleSystem,
			Content: []*ai.Part{
				ai.NewTextPart(systemPrompt),
			},
		},
	}

	for _, msg := range history {
		role := ai.RoleUser
		if msg.Role == "model" {
			role = ai.RoleModel
		}
		messages = append(messages, &ai.Message{
			Role: role,
			Content: []*ai.Part{
				ai.NewTextPart(msg.Content),
			},
		})
	}

	messages = append(messages, &ai.Message{
		Role: ai.RoleUser,
		Content: []*ai.Part{
			ai.NewTextPart(userInput),
		},
	})

	req := &ai.GenerateRequest{
		Messages: messages,
		Tools:    []ai.Tool{a.searchDbTool},
	}

	res, err := a.model.Generate(ctx, req, nil)
	if err != nil {
		return "", err
	}

	return res.Text(), nil
}
