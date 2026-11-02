package cleaner

import (
	"regexp"
	"strconv"
)

var (
	// Padrão 1: Série #01 - Título ou Série #1.5 - Título
	seriesHashRegex = regexp.MustCompile(`(?i)^(.*?)\s*#(\d+(?:\.\d+)?)\s*[-:]\s*(.*)$`)

	// Padrão 2: Série - Vol. 2 - Título ou Série - Volume 02 - Título
	seriesVolRegex = regexp.MustCompile(`(?i)^(.*?)\s*[-:]\s*(?:vol(?:ume|\.)?)\s*(\d+(?:\.\d+)?)\s*[-:]\s*(.*)$`)

	// Padrão 3: Série, Livro 3 - Título ou Série - Book 3 - Título
	seriesBookRegex = regexp.MustCompile(`(?i)^(.*?)\s*[,-:]\s*(?:livro|book)\s*(\d+(?:\.\d+)?)\s*[-:]\s*(.*)$`)

	// Padrão 4: Série - 01 - Título
	seriesNumDashRegex = regexp.MustCompile(`(?i)^(.*?)\s*[-:]\s*(\d{1,3})\s*[-:]\s*(.*)$`)
)

// ParseSeries extrai dinamicamente a Série, o Índice (float64) e o Título limpo.
func ParseSeries(raw string) (cleanTitle string, seriesName string, seriesIndex float64, found bool) {
	cleaned := CleanString(raw)

	// Tenta Padrão 1: Série #01 - Título
	if matches := seriesHashRegex.FindStringSubmatch(cleaned); len(matches) == 4 {
		idx, err := strconv.ParseFloat(matches[2], 64)
		if err == nil {
			return CleanString(matches[3]), CleanString(matches[1]), idx, true
		}
	}

	// Tenta Padrão 2: Série - Vol. 2 - Título
	if matches := seriesVolRegex.FindStringSubmatch(cleaned); len(matches) == 4 {
		idx, err := strconv.ParseFloat(matches[2], 64)
		if err == nil {
			return CleanString(matches[3]), CleanString(matches[1]), idx, true
		}
	}

	// Tenta Padrão 3: Série, Livro 3 - Título
	if matches := seriesBookRegex.FindStringSubmatch(cleaned); len(matches) == 4 {
		idx, err := strconv.ParseFloat(matches[2], 64)
		if err == nil {
			return CleanString(matches[3]), CleanString(matches[1]), idx, true
		}
	}

	// Tenta Padrão 4: Série - 01 - Título
	if matches := seriesNumDashRegex.FindStringSubmatch(cleaned); len(matches) == 4 {
		idx, err := strconv.ParseFloat(matches[2], 64)
		if err == nil {
			return CleanString(matches[3]), CleanString(matches[1]), idx, true
		}
	}

	return cleaned, "", 0, false
}
