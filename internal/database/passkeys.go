package database

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// Passkey representa uma credencial WebAuthn / FIDO2 registada para um utilizador.
type Passkey struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"user_id"`
	CredentialID string     `json:"credential_id"`
	PublicKey    string     `json:"public_key"`
	Name         string     `json:"name"`
	SignCount    int        `json:"sign_count"`
	CreatedAt    time.Time  `json:"created_at"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
}

// OIDCConfig contém os parâmetros de configuração do fornecedor OpenID Connect / SSO.
type OIDCConfig struct {
	IssuerURL    string `json:"issuer_url"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
	DisplayName  string `json:"display_name"`
	Enabled      bool   `json:"enabled"`
}

// EnsurePasskeyTable cria a tabela de credenciais WebAuthn caso não exista.
func (d *DB) EnsurePasskeyTable(ctx context.Context) error {
	if d.appPool == nil {
		return nil
	}
	_, err := d.appPool.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS user_passkey (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			credential_id TEXT UNIQUE NOT NULL,
			public_key TEXT NOT NULL,
			name TEXT NOT NULL,
			sign_count INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_used_at DATETIME,
			FOREIGN KEY(user_id) REFERENCES user(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_user_passkey_cred ON user_passkey(credential_id);
	`)
	return err
}

// EnsureOIDCTable cria a tabela de configuração OIDC caso não exista.
func (d *DB) EnsureOIDCTable(ctx context.Context) error {
	if d.appPool == nil {
		return nil
	}
	_, err := d.appPool.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS oidc_config (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			issuer_url TEXT NOT NULL DEFAULT '',
			client_id TEXT NOT NULL DEFAULT '',
			client_secret TEXT NOT NULL DEFAULT '',
			redirect_url TEXT NOT NULL DEFAULT '',
			display_name TEXT NOT NULL DEFAULT 'Single Sign-On',
			enabled BOOLEAN NOT NULL DEFAULT 0
		);
		INSERT OR IGNORE INTO oidc_config (id, enabled) VALUES (1, 0);
	`)
	return err
}

// AddPasskey regista uma nova chave WebAuthn / Passkey para o utilizador.
func (d *DB) AddPasskey(ctx context.Context, userID int64, credID, pubKey, name string) error {
	if d.appPool == nil {
		return errors.New("app.db não inicializado")
	}
	_ = d.EnsurePasskeyTable(ctx)

	_, err := d.appPool.ExecContext(ctx, `
		INSERT INTO user_passkey (user_id, credential_id, public_key, name)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(credential_id) DO UPDATE SET
			public_key = excluded.public_key,
			name = excluded.name,
			last_used_at = CURRENT_TIMESTAMP
	`, userID, credID, pubKey, name)
	return err
}

// ListPasskeys devolve todas as passkeys ativas de um utilizador.
func (d *DB) ListPasskeys(ctx context.Context, userID int64) ([]Passkey, error) {
	if d.appPool == nil {
		return []Passkey{}, nil
	}
	_ = d.EnsurePasskeyTable(ctx)

	rows, err := d.appPool.QueryContext(ctx, `
		SELECT id, user_id, credential_id, public_key, name, sign_count, created_at, last_used_at
		FROM user_passkey
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return []Passkey{}, err
	}
	defer rows.Close()

	var list []Passkey
	for rows.Next() {
		var pk Passkey
		var lastUsed sql.NullTime
		if err := rows.Scan(&pk.ID, &pk.UserID, &pk.CredentialID, &pk.PublicKey, &pk.Name, &pk.SignCount, &pk.CreatedAt, &lastUsed); err != nil {
			continue
		}
		if lastUsed.Valid {
			pk.LastUsedAt = &lastUsed.Time
		}
		list = append(list, pk)
	}
	if err := rows.Err(); err != nil {
		return []Passkey{}, err
	}
	return list, nil
}

// GetPasskeyByCredentialID localiza a passkey e o respetivo utilizador pelo identificador da credencial.
func (d *DB) GetPasskeyByCredentialID(ctx context.Context, credID string) (*Passkey, *User, error) {
	if d.appPool == nil {
		return nil, nil, errors.New("app.db não inicializado")
	}
	_ = d.EnsurePasskeyTable(ctx)

	var pk Passkey
	var lastUsed sql.NullTime
	var u User

	err := d.appPool.QueryRowContext(ctx, `
		SELECT p.id, p.user_id, p.credential_id, p.public_key, p.name, p.sign_count, p.created_at, p.last_used_at,
		       u.id, u.name, COALESCE(u.email, ''), COALESCE(u.kindle_mail, ''), u.role, COALESCE(u.locale, 'pt')
		FROM user_passkey p
		JOIN user u ON p.user_id = u.id
		WHERE p.credential_id = ?
		LIMIT 1
	`, credID).Scan(
		&pk.ID, &pk.UserID, &pk.CredentialID, &pk.PublicKey, &pk.Name, &pk.SignCount, &pk.CreatedAt, &lastUsed,
		&u.ID, &u.Name, &u.Email, &u.KindleMail, &u.Role, &u.Locale,
	)
	if err != nil {
		return nil, nil, err
	}
	if lastUsed.Valid {
		pk.LastUsedAt = &lastUsed.Time
	}
	u.IsAdmin = (u.Role & 1) != 0

	// Atualiza data de última utilização
	_, _ = d.appPool.ExecContext(ctx, `UPDATE user_passkey SET last_used_at = CURRENT_TIMESTAMP WHERE id = ?`, pk.ID)

	return &pk, &u, nil
}

// DeletePasskey elimina uma passkey pertencente ao utilizador.
func (d *DB) DeletePasskey(ctx context.Context, userID, passkeyID int64) error {
	if d.appPool == nil {
		return errors.New("app.db não inicializado")
	}
	_ = d.EnsurePasskeyTable(ctx)

	_, err := d.appPool.ExecContext(ctx, `
		DELETE FROM user_passkey WHERE id = ? AND user_id = ?
	`, passkeyID, userID)
	return err
}

// GetOIDCConfig carrega a configuração OIDC do sistema.
func (d *DB) GetOIDCConfig(ctx context.Context) (*OIDCConfig, error) {
	if d.appPool == nil {
		return &OIDCConfig{}, nil
	}
	_ = d.EnsureOIDCTable(ctx)

	var cfg OIDCConfig
	err := d.appPool.QueryRowContext(ctx, `
		SELECT issuer_url, client_id, client_secret, redirect_url, display_name, enabled
		FROM oidc_config
		WHERE id = 1
	`).Scan(&cfg.IssuerURL, &cfg.ClientID, &cfg.ClientSecret, &cfg.RedirectURL, &cfg.DisplayName, &cfg.Enabled)
	if err != nil {
		return &OIDCConfig{}, err
	}
	return &cfg, nil
}

// SaveOIDCConfig guarda a configuração do fornecedor OIDC.
func (d *DB) SaveOIDCConfig(ctx context.Context, cfg *OIDCConfig) error {
	if d.appPool == nil {
		return errors.New("app.db não inicializado")
	}
	_ = d.EnsureOIDCTable(ctx)

	_, err := d.appPool.ExecContext(ctx, `
		UPDATE oidc_config SET
			issuer_url = ?,
			client_id = ?,
			client_secret = ?,
			redirect_url = ?,
			display_name = ?,
			enabled = ?
		WHERE id = 1
	`, cfg.IssuerURL, cfg.ClientID, cfg.ClientSecret, cfg.RedirectURL, cfg.DisplayName, cfg.Enabled)
	return err
}

