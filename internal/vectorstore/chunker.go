package vectorstore

import (
	"strings"
)

// ChunkText splits text into chunks by word count with an overlap.
func ChunkText(text string, chunkSize, overlap int) []string {
	words := strings.Fields(text)
	var chunks []string

	if len(words) == 0 {
		return chunks
	}

	for i := 0; i < len(words); {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}

		chunk := strings.Join(words[i:end], " ")
		chunks = append(chunks, chunk)

		if end == len(words) {
			break
		}

		step := chunkSize - overlap
		if step <= 0 {
			step = 1
		}
		i += step
	}

	return chunks
}
