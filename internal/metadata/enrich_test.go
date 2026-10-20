package metadata

import (
	"context"
	"testing"
	"time"
)

func TestCleanAuthorName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Shelley, Mary Wollstonecraft", "Mary Wollstonecraft Shelley"},
		{"active 6th century B.C. Sunzi", "Sun Tzu"},
		{"Wells, H. G. (Herbert George)", "H. G. Wells"},
		{"Stevenson, Robert Louis (1850-1894)", "Robert Louis Stevenson"},
		{"Niccolò Machiavelli", "Niccolò Machiavelli"},
	}

	for _, tt := range tests {
		got := CleanAuthorName(tt.input)
		if got != tt.expected {
			t.Errorf("CleanAuthorName(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestCleanBookTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Frankenstein (2)", "Frankenstein"},
		{"The Great Gatsby [Project Gutenberg]", "The Great Gatsby"},
		{"Dracula (eBook)", "Dracula"},
	}

	for _, tt := range tests {
		got := CleanBookTitle(tt.input)
		if got != tt.expected {
			t.Errorf("CleanBookTitle(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFetchEnrichmentCandidatesOnline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	candidates, err := FetchEnrichmentCandidates(ctx, "Frankenstein Mary Shelley", "Frankenstein", "Mary Shelley")
	if err != nil {
		t.Fatalf("FetchEnrichmentCandidates returned error: %v", err)
	}

	if len(candidates) == 0 {
		t.Log("Nenhum candidato retornado (possível problema de rede em teste)")
		return
	}

	first := candidates[0]
	if first.Title == "" {
		t.Errorf("Esperado título preenchido, recebido vazio")
	}
	t.Logf("Candidato encontrado: %s por %v (Capas HD: %d, Tags: %v)", first.Title, first.Authors, len(first.Covers), first.Tags)
}

