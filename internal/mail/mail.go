package mail

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SMTPSettings struct {
	Server   string
	Port     int
	UseSSL   int // 0: None/STARTTLS, 1: Direct SSL/TLS
	Login    string
	Password string
	From     string
}

// SendBookEmail monta a mensagem MIME com anexo em base64 e envia por SMTP.
func SendBookEmail(cfg SMTPSettings, recipient, subject, bookFilePath, attachmentName string) error {
	if cfg.Server == "" || cfg.Port == 0 {
		return errors.New("servidor SMTP não configurado nas definições")
	}
	if recipient == "" {
		return errors.New("endereço de email do destinatário não configurado no perfil")
	}

	from := cfg.From
	if from == "" {
		from = cfg.Login
	}
	if from == "" {
		from = "calibre-web@localhost"
	}

	content, err := os.ReadFile(bookFilePath)
	if err != nil {
		return fmt.Errorf("ficheiro do livro não encontrado: %w", err)
	}

	if attachmentName == "" {
		attachmentName = filepath.Base(bookFilePath)
	}

	boundary := fmt.Sprintf("=_cwa_%d", time.Now().UnixNano())

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", recipient))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", boundary))

	// Corpo do email
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	msg.WriteString(fmt.Sprintf("O livro '%s' foi enviado pelo Calibre-Web Automated.\r\n\r\n", subject))

	// Anexo
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString(fmt.Sprintf("Content-Type: application/octet-stream; name=\"%s\"\r\n", attachmentName))
	msg.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", attachmentName))
	msg.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")

	b64Data := base64.StdEncoding.EncodeToString(content)
	for i := 0; i < len(b64Data); i += 76 {
		end := i + 76
		if end > len(b64Data) {
			end = len(b64Data)
		}
		msg.WriteString(b64Data[i:end])
		msg.WriteString("\r\n")
	}

	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	addr := fmt.Sprintf("%s:%d", cfg.Server, cfg.Port)
	rawMsg := []byte(msg.String())

	// Conexão direta SSL/TLS (porta 465 ou UseSSL == 1)
	if cfg.UseSSL == 1 || cfg.Port == 465 {
		tlsConfig := &tls.Config{
			ServerName: cfg.Server,
			InsecureSkipVerify: false,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("falha ao estabelecer conexão SSL/TLS: %w", err)
		}
		defer conn.Close()

		c, err := smtp.NewClient(conn, cfg.Server)
		if err != nil {
			return err
		}
		defer c.Close()

		if cfg.Login != "" && cfg.Password != "" {
			auth := smtp.PlainAuth("", cfg.Login, cfg.Password, cfg.Server)
			if err := c.Auth(auth); err != nil {
				return fmt.Errorf("falha de autenticação SMTP: %w", err)
			}
		}

		if err := c.Mail(from); err != nil {
			return err
		}
		if err := c.Rcpt(recipient); err != nil {
			return err
		}
		w, err := c.Data()
		if err != nil {
			return err
		}
		if _, err := w.Write(rawMsg); err != nil {
			_ = w.Close()
			return err
		}
		return w.Close()
	}

	// Conexão padrão STARTTLS ou texto simples
	var auth smtp.Auth
	if cfg.Login != "" && cfg.Password != "" {
		auth = smtp.PlainAuth("", cfg.Login, cfg.Password, cfg.Server)
	}

	c, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("falha ao conectar ao servidor SMTP: %w", err)
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName: cfg.Server,
		}
		if err := c.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("falha no STARTTLS: %w", err)
		}
	}

	if auth != nil {
		if err := c.Auth(auth); err != nil {
			return fmt.Errorf("falha na autenticação SMTP: %w", err)
		}
	}

	if err := c.Mail(from); err != nil {
		return err
	}
	if err := c.Rcpt(recipient); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(rawMsg); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}
