package auth

import "testing"

func TestWerkzeugCompatibility(t *testing.T) {
	storedHash := "pbkdf2:sha256:600000$3fr5sRXuYJHgub84$d6ca5283c141e14d6c0d6fbc35e4451199ac7a65a392217cac48f7f3455ea9ea"

	if !CheckPassword(storedHash, "admin123") {
		t.Fatal("Esperado sucesso ao validar hash do Werkzeug com senha admin123")
	}

	if CheckPassword(storedHash, "wrongpass") {
		t.Fatal("Esperada falha com senha errada")
	}
}

func TestHashPasswordRoundTrip(t *testing.T) {
	plain := "MinhaSenhaSuperSecreta@2026"
	h, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("Erro ao gerar hash: %v", err)
	}

	if !CheckPassword(h, plain) {
		t.Fatal("Falha no roundtrip de HashPassword e CheckPassword")
	}

	if CheckPassword(h, "outrasenha") {
		t.Fatal("Senha diferente não deve passar")
	}
}

