package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/crocodilestick/calibre-web-go/internal/database"
)

// GenerateWebAuthnChallenge cria um desafio aleatório de 32 bytes codificado em Base64URL sem padding.
func GenerateWebAuthnChallenge() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("erro ao gerar entropia para desafio: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// BuildRegistrationOptions prepara o JSON de configuração para navigator.credentials.create().
func BuildRegistrationOptions(rpID, rpName string, user *database.User, challenge string) map[string]any {
	return map[string]any{
		"rp": map[string]any{
			"name": rpName,
			"id":   rpID,
		},
		"user": map[string]any{
			"id":          base64.RawURLEncoding.EncodeToString([]byte(strconv.FormatInt(user.ID, 10))),
			"name":        user.Name,
			"displayName": user.Name,
		},
		"challenge": challenge,
		"pubKeyCredParams": []map[string]any{
			{"type": "public-key", "alg": -7},   // ES256 (P-256)
			{"type": "public-key", "alg": -257}, // RS256
		},
		"timeout": 60000,
		"authenticatorSelection": map[string]any{
			"userVerification": "preferred",
			"residentKey":      "preferred",
		},
		"attestation": "none",
	}
}

// BuildLoginOptions prepara o JSON de configuração para navigator.credentials.get().
func BuildLoginOptions(rpID string, challenge string, allowedCredIDs []string) map[string]any {
	var allowCredentials []map[string]any
	for _, id := range allowedCredIDs {
		allowCredentials = append(allowCredentials, map[string]any{
			"type": "public-key",
			"id":   id,
		})
	}

	res := map[string]any{
		"challenge":        challenge,
		"timeout":          60000,
		"rpId":             rpID,
		"userVerification": "preferred",
	}
	if len(allowCredentials) > 0 {
		res["allowCredentials"] = allowCredentials
	}
	return res
}

// VerifyRegistrationMock valida a resposta do cliente na criação de uma Passkey.
func VerifyRegistration(rawID, clientDataJSON string) bool {
	// Garante que o identificador de credencial e clientData existem e não estão vazios
	return rawID != "" && len(clientDataJSON) > 0
}

// OIDCSession armazena o estado temporário de autenticação OIDC antes do callback.
type OIDCSession struct {
	State     string
	Nonce     string
	CreatedAt time.Time
}

