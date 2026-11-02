package cleaner

import (
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// Remove tags de trackers e releases entre colchetes: [MAM], [Retail], [v1.0], [ePub], etc.
	bracketTagRegex = regexp.MustCompile(`(?i)\[[^\]]*\]`)

	// Remove domínios web promocionais: www.livros.me, torrents.to, etc.
	webDomainRegex = regexp.MustCompile(`(?i)\b(?:www\.)?[a-zA-Z0-9-]+\.(?:com|org|net|me|io|info|biz|club|xyz|to|cc|pt|br|es|fr|de)\b`)

	// Remove anos entre parênteses: (2021), (1998)
	yearRegex = regexp.MustCompile(`\((?:19|20)\d{2}\)`)

	// Remove tags de formato, idioma ou release entre parênteses: (EPUB), (Retail), (PT-PT), (Scan), etc.
	formatParenRegex = regexp.MustCompile(`(?i)\(\s*(?:epub|pdf|mobi|azw3?|cbz|cbr|retail|repack|pt-pt|pt-br|eng?|scan|v\d+(\.\d+)*)\s*\)`)

	// Remove palavras de formato soltas no final ou entre delimitadores
	formatLooseRegex = regexp.MustCompile(`(?i)\b(?:epub|pdf|mobi|azw3?|cbr|cbz|repack|retail|hq)\b`)

	// Colapsa múltiplos espaços
	multiSpaceRegex = regexp.MustCompile(`\s{2,}`)
)

var knownBookExtensions = map[string]bool{
	".epub": true,
	".pdf":  true,
	".mobi": true,
	".azw":  true,
	".azw3": true,
	".cbz":  true,
	".cbr":  true,
	".zip":  true,
	".rar":  true,
	".fb2":  true,
}

// CleanString aplica a cadeia rigorosa de higienização de nomes e títulos.
func CleanString(raw string) string {
	// 1. Remove extensão de livro se presente
	ext := strings.ToLower(filepath.Ext(raw))
	if knownBookExtensions[ext] {
		raw = raw[:len(raw)-len(ext)]
	}

	// 2. Substitui underscores por espaços
	s := strings.ReplaceAll(raw, "_", " ")

	// 3. Remove domínios web
	s = webDomainRegex.ReplaceAllString(s, " ")

	// 4. Remove tags entre colchetes [Tracker]
	s = bracketTagRegex.ReplaceAllString(s, " ")

	// 5. Remove tags de formato/release entre parênteses (EPUB) e anos (2024)
	s = formatParenRegex.ReplaceAllString(s, " ")
	s = yearRegex.ReplaceAllString(s, " ")

	// 6. Remove palavras de formato soltas
	s = formatLooseRegex.ReplaceAllString(s, " ")

	// 7. Limpa traços duplicados ou órfãos
	s = strings.ReplaceAll(s, " - - ", " - ")
	s = strings.ReplaceAll(s, "--", "-")

	// 8. Colapsa espaços e limpa pontuação nas extremidades
	s = multiSpaceRegex.ReplaceAllString(s, " ")
	s = strings.Trim(s, " -_.,;:[]()")

	return strings.TrimSpace(s)
}

// CleanTitleAndAuthor decompõe e higieniza uma string no formato "Autor - Título" ou "Título".
func CleanTitleAndAuthor(raw string) (title string, author string) {
	cleaned := CleanString(raw)

	parts := strings.Split(cleaned, " - ")
	if len(parts) >= 2 {
		// Padrão comum em releases de livros: Autor - Título (ou Autor - Série - Título)
		first := strings.TrimSpace(parts[0])
		rest := strings.TrimSpace(strings.Join(parts[1:], " - "))

		// Se a primeira parte não contiver dígitos e não for muito longa, assume-se como autor
		if len(first) > 0 && len(first) < 50 && !strings.ContainsAny(first, "0123456789#") {
			return rest, first
		}

		// Caso contrário, usa toda a string tratada como título
		return cleaned, ""
	}

	return cleaned, ""
}
