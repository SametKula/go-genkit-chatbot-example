package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/ollama"
	"github.com/siber/go-genkit-rag-chatbot/internal/memory"
	"github.com/siber/go-genkit-rag-chatbot/internal/vectorstore"
)

type Agent struct {
	qClient      *vectorstore.SimpleQdrantClient
	embedder     ai.Embedder
	model        ai.Model
	searchDbTool     ai.Tool
	createTicketTool ai.Tool
	g                *genkit.Genkit
}

type SearchDbInput struct {
	Query string `json:"query" jsonschema:"description=The search query string to look for in the manuals"`
}

type CreateTicketInput struct {
	Issue string `json:"issue" jsonschema:"description=The detailed description of the user's issue to be included in the support ticket"`
}

func NewAgent(ctx context.Context, qClient *vectorstore.SimpleQdrantClient) (*Agent, error) {
	plugin := &ollama.Ollama{
		ServerAddress: "http://localhost:11434",
	}
	g := genkit.Init(ctx, genkit.WithPlugins(plugin))

	embedder := plugin.DefineEmbedder(g, "http://localhost:11434", "nomic-embed-text", nil)
	model := plugin.DefineModel(g, ollama.ModelDefinition{
		Name: "llama3.1:8b",
		Type: "chat",
	}, &ai.ModelOptions{
		Supports: &ai.ModelSupports{
			Tools:      true,
			SystemRole: true,
			Multiturn:  true,
		},
	})

	searchDbTool := genkit.DefineTool[*SearchDbInput, string](
		g,
		"search_db",
		"Searches the knowledge base for manuals to solve the user's problem. Input MUST be a JSON object with a single 'query' string field.",
		func(ctx *ai.ToolContext, input *SearchDbInput) (string, error) {
			res, err := genkit.Embed(ctx, g,
				ai.WithEmbedder(embedder),
				ai.WithTextDocs(input.Query),
			)
			if err != nil {
				return "", err
			}

			if len(res.Embeddings) == 0 {
				return "No embedding generated for query", nil
			}

			vec := res.Embeddings[0].Embedding

			texts, err := qClient.Search(ctx, "manuals", vec, input.Query, 3)
			if err != nil {
				return "", err
			}

			if len(texts) == 0 {
				return "No relevant information found in manuals.", nil
			}

			return "Bulunan kılavuz içerikleri: \n\n" + strings.Join(texts, "\n\n---\n\n"), nil
		},
	)

	createTicketTool := genkit.DefineTool[*CreateTicketInput, string](
		g,
		"create_ticket",
		"Creates a support ticket when the user's issue cannot be resolved using the manuals or if the user explicitly requests a ticket.",
		func(ctx *ai.ToolContext, input *CreateTicketInput) (string, error) {
			// In a real app, this would call a Jira/Zendesk API.
			// For now, we just log it to the console as requested.
			fmt.Printf("[TICKET] New ticket created! Issue: %s\n", input.Issue)
			return "Destek talebiniz başarıyla oluşturuldu. Teknik ekibimiz en kısa sürede size dönüş yapacaktır.", nil
		},
	)

	return &Agent{
		qClient:          qClient,
		embedder:         embedder,
		model:            model,
		searchDbTool:     searchDbTool,
		createTicketTool: createTicketTool,
		g:                g,
	}, nil
}

func (a *Agent) Chat(ctx context.Context, userInput string, history []memory.Message) (string, error) {
	systemPrompt := `Sen uzman bir ürün destek asistanısın. Agentic RAG tabanlı bir Chatbotsun.
ÖNEMLİ KURALLAR:
1. GÖREVİN: Kullanıcının EN SON gönderdiği mesajdaki soruya odaklanmak ve ona cevap vermektir.
2. BAĞLAM: Önceki mesajları sadece konuşmanın akışını anlamak için bağlam (context) olarak kullan. Eğer kullanıcı önceki konudan bağımsız yeni bir soru sorduysa, sadece yeni soruya odaklan ve önceki cevapları tekrar etme.
3. ARAÇ KULLANIMI: Bir soru aldığında, ASLA KULLANICIYA İZİN SORMADAN doğrudan 'search_db' aracını (tool) ÇAĞIR. Aramayı otomatik yap.
4. BİLGİ KAYNAĞI: Kullanıcının sorusu sadece bilgi almak amaçlıysa, 'search_db' sonuçlarına dayanarak doğrudan tam ve açıklayıcı bilgiyi ver.
5. ADIM YÖNETİMİ: Kullanıcı bir sorun çözümü veya kurulum adımı istiyorsa, adımları nasıl vereceğine (tek seferde mi yoksa adım adım bekleyerek mi) DURUMA GÖRE KENDİN KARAR VER.
6. DOĞRULUK: Tüm cevaplarını kesinlikle 'search_db' sonuçlarına dayandır. Veritabanında olmayan uydurma bilgiler verme.`

	var messages []*ai.Message
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

	res, err := genkit.GenerateText(ctx, a.g,
		ai.WithModel(a.model),
		ai.WithSystem(systemPrompt),
		ai.WithMessages(messages...),
		ai.WithPrompt(userInput),
		ai.WithTools(a.searchDbTool, a.createTicketTool),
	)
	if err != nil {
		return "", err
	}

	return res, nil
}
