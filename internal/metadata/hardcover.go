package metadata

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const HardcoverGraphQLEndpoint = "https://api.hardcover.app/v1/graphql"

type HardcoverClient struct {
	httpClient *http.Client
	apiKey     string
}

func NewHardcoverClient(apiKey string) *HardcoverClient {
	return &HardcoverClient{
		httpClient: &http.Client{Timeout: 8 * time.Second},
		apiKey:     apiKey,
	}
}

type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type hardcoverResponse struct {
	Data struct {
		Books []struct {
			ID    int    `json:"id"`
			Title string `json:"title"`
			Slug  string `json:"slug"`
		} `json:"books"`
		Editions []struct {
			BookID int `json:"book_id"`
		} `json:"editions"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// FetchHardcoverID pesquisa no Hardcover.app GraphQL por ISBN ou Título e devolve o Hardcover ID.
func (c *HardcoverClient) FetchHardcoverID(ctx context.Context, title, author, isbn string) (string, error) {
	isbn = strings.ReplaceAll(strings.ReplaceAll(isbn, "-", ""), " ", "")

	// 1. Tenta pesquisar por ISBN se disponível
	if len(isbn) >= 10 {
		query := `query SearchEditions($isbn: String!) {
			editions(where: {isbn_13: {_eq: $isbn}}, limit: 1) {
				book_id
			}
		}`
		reqBody, _ := json.Marshal(graphQLRequest{
			Query:     query,
			Variables: map[string]any{"isbn": isbn},
		})

		req, err := http.NewRequestWithContext(ctx, "POST", HardcoverGraphQLEndpoint, bytes.NewReader(reqBody))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
			if c.apiKey != "" {
				req.Header.Set("Authorization", "Bearer "+c.apiKey)
			}
			resp, err := c.httpClient.Do(req)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					var hResp hardcoverResponse
					if err := json.NewDecoder(resp.Body).Decode(&hResp); err == nil {
						if len(hResp.Data.Editions) > 0 && hResp.Data.Editions[0].BookID > 0 {
							return strconv.Itoa(hResp.Data.Editions[0].BookID), nil
						}
					}
				}
			}
		}
	}

	// 2. Pesquisa por Título
	if strings.TrimSpace(title) != "" {
		query := `query SearchBooks($title: String!) {
			books(where: {title: {_ilike: $title}}, limit: 5) {
				id
				title
				slug
			}
		}`
		reqBody, _ := json.Marshal(graphQLRequest{
			Query:     query,
			Variables: map[string]any{"title": "%" + strings.TrimSpace(title) + "%"},
		})

		req, err := http.NewRequestWithContext(ctx, "POST", HardcoverGraphQLEndpoint, bytes.NewReader(reqBody))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")
		if c.apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.apiKey)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("erro na conexão ao Hardcover: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("hardcover retornou status HTTP %d", resp.StatusCode)
		}

		var hResp hardcoverResponse
		if err := json.NewDecoder(resp.Body).Decode(&hResp); err != nil {
			return "", fmt.Errorf("falha ao descodificar resposta Hardcover: %w", err)
		}

		if len(hResp.Errors) > 0 {
			return "", fmt.Errorf("hardcover GraphQL erro: %s", hResp.Errors[0].Message)
		}

		if len(hResp.Data.Books) > 0 {
			return strconv.Itoa(hResp.Data.Books[0].ID), nil
		}
	}

	return "", fmt.Errorf("livro não encontrado no Hardcover")
}

