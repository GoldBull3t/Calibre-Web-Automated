package cleaner

import (
	"testing"
)

func TestCleanString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Tracker tags and years",
			input:    "[MAM] J.K. Rowling - Harry Potter (2001) [Retail] [EPUB]",
			expected: "J.K. Rowling - Harry Potter",
		},
		{
			name:     "Web domains and brackets",
			input:    "Stephen King - O Iluminado (www.livrosgratis.com) [v1.0].epub",
			expected: "Stephen King - O Iluminado",
		},
		{
			name:     "Underscores and format tags",
			input:    "Isaac_Asimov_-_Fundacao_e_Imperio_[PT-PT]_(2023)",
			expected: "Isaac Asimov - Fundacao e Imperio",
		},
		{
			name:     "Loose formats and multiple spaces",
			input:    "Brandon Sanderson    -   Mistborn   REPACK  EPUB",
			expected: "Brandon Sanderson - Mistborn",
		},
		{
			name:     "HQ and scan tag",
			input:    "[HQ] Batman - Ano Um (Scan) (1987) [CBR]",
			expected: "Batman - Ano Um",
		},
		{
			name:     "Clean filename without artifacts",
			input:    "George Orwell - 1984",
			expected: "George Orwell - 1984",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanString(tt.input)
			if got != tt.expected {
				t.Errorf("CleanString(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCleanTitleAndAuthor(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedTitle  string
		expectedAuthor string
	}{
		{
			name:           "Standard Author - Title",
			input:          "[Release] Arthur Conan Doyle - Um Estudo em Vermelho [EPUB]",
			expectedTitle:  "Um Estudo em Vermelho",
			expectedAuthor: "Arthur Conan Doyle",
		},
		{
			name:           "Title only without author separator",
			input:          "A Metamorfose (1915).epub",
			expectedTitle:  "A Metamorfose",
			expectedAuthor: "",
		},
		{
			name:           "Author with multiple parts",
			input:          "Gabriel Garcia Marquez - Cem Anos de Solidao (www.livros.me)",
			expectedTitle:  "Cem Anos de Solidao",
			expectedAuthor: "Gabriel Garcia Marquez",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, author := CleanTitleAndAuthor(tt.input)
			if title != tt.expectedTitle || author != tt.expectedAuthor {
				t.Errorf("CleanTitleAndAuthor(%q) = (title: %q, author: %q), want (title: %q, author: %q)",
					tt.input, title, author, tt.expectedTitle, tt.expectedAuthor)
			}
		})
	}
}

