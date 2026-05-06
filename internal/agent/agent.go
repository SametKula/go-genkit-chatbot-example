package agent

import (
	"context"
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
	searchDbTool ai.Tool
	g            *genkit.Genkit
}

type SearchDbInput struct {
	Query string `json:"query" jsonschema:"description=The search query string to look for in the manuals"`
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
		g:            g,
	}, nil
}

func (a *Agent) Chat(ctx context.Context, userInput string, history []memory.Message) (string, error) {
	systemPrompt := `Sen uzman bir sorun giderme asistanısın. Agentic RAG tabanlı bir Chatbotsun.
ÖNEMLİ KURALLAR:
1. Bir soru aldığında, ASLA KULLANICIYA İZİN SORMADAN doğrudan 'search_db' aracını (tool) ÇAĞIR. Aramayı otomatik yap.
2. Çözümleri her zaman 'search_db' aracından gelen sonuçlara dayandır. Veritabanında olmayan uydurma bilgiler verme.
3. Sorun çözümü birden fazla adımdan oluşuyorsa, adımları tek tek ver. Kullanıcı ilk adımı uygulayıp onaylamadan ikinci adıma geçme.
4. 'search_db' dışındaki araçlar (ileride eklenecek) için izin isteyebilirsin, ancak arama işlemi için beklemeden aracı kullan.`

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
		ai.WithTools(a.searchDbTool),
	)
	if err != nil {
		return "", err
	}

	return res, nil
}
