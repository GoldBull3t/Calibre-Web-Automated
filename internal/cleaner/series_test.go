package cleaner

import (
	"testing"
)

func TestParseSeries(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedTitle string
		expectedSerie string
		expectedIndex float64
		expectedFound bool
	}{
		{
			name:          "Pattern 1 with Hash #01",
			input:         "Harry Potter #01 - A Pedra Filosofal",
			expectedTitle: "A Pedra Filosofal",
			expectedSerie: "Harry Potter",
			expectedIndex: 1.0,
			expectedFound: true,
		},
		{
			name:          "Pattern 1 with float index #2.5",
			input:         "The Expanse #2.5 - O Açougueiro de Anderson Station",
			expectedTitle: "O Açougueiro de Anderson Station",
			expectedSerie: "The Expanse",
			expectedIndex: 2.5,
			expectedFound: true,
		},
		{
			name:          "Pattern 2 with Vol 2",
			input:         "[Retail] Fundação - Vol. 2 - Segunda Fundação [EPUB]",
			expectedTitle: "Segunda Fundação",
			expectedSerie: "Fundação",
			expectedIndex: 2.0,
			expectedFound: true,
		},
		{
			name:          "Pattern 2 with Volume written out",
			input:         "Duna - Volume 03 - Os Filhos de Duna (2020)",
			expectedTitle: "Os Filhos de Duna",
			expectedSerie: "Duna",
			expectedIndex: 3.0,
			expectedFound: true,
		},
		{
			name:          "Pattern 3 with Livro 4",
			input:         "A Torre Negra, Livro 4 - Mago e Vidro",
			expectedTitle: "Mago e Vidro",
			expectedSerie: "A Torre Negra",
			expectedIndex: 4.0,
			expectedFound: true,
		},
		{
			name:          "Pattern 3 with Book 1",
			input:         "Mistborn - Book 1 - O Império Final",
			expectedTitle: "O Império Final",
			expectedSerie: "Mistborn",
			expectedIndex: 1.0,
			expectedFound: true,
		},
		{
			name:          "Pattern 4 with number and dash",
			input:         "Sandman - 01 - Prelúdios e Noturnos",
			expectedTitle: "Prelúdios e Noturnos",
			expectedSerie: "Sandman",
			expectedIndex: 1.0,
			expectedFound: true,
		},
		{
			name:          "No series present",
			input:         "Dom Casmurro - Machado de Assis",
			expectedTitle: "Dom Casmurro - Machado de Assis",
			expectedSerie: "",
			expectedIndex: 0,
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanTitle, seriesName, seriesIndex, found := ParseSeries(tt.input)
			if found != tt.expectedFound {
				t.Fatalf("ParseSeries(%q) found = %v, want %v", tt.input, found, tt.expectedFound)
			}
			if found {
				if cleanTitle != tt.expectedTitle {
					t.Errorf("cleanTitle = %q, want %q", cleanTitle, tt.expectedTitle)
				}
				if seriesName != tt.expectedSerie {
					t.Errorf("seriesName = %q, want %q", seriesName, tt.expectedSerie)
				}
				if seriesIndex != tt.expectedIndex {
					t.Errorf("seriesIndex = %v, want %v", seriesIndex, tt.expectedIndex)
				}
			}
		})
	}
}
