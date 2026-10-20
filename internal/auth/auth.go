package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash"
	"math/big"
	"strconv"
	"strings"

	"github.com/crocodilestick/calibre-web-go/internal/database"
)

type contextKey string

const UserContextKey contextKey = "currentUser"

const (
	DefaultSaltLen = 16
	DefaultIter    = 600000
	KeyLen         = 32
)

// pbkdf2Key implements RFC 2898 PBKDF2 using any standard hash constructor.
func pbkdf2Key(password, salt []byte, iter, keyLen int, h func() hash.Hash) []byte {
	prf := hmac.New(h, password)
	hashLen := prf.Size()
	numBlocks := (keyLen + hashLen - 1) / hashLen

	var buf [4]byte
	dk := make([]byte, 0, numBlocks*hashLen)
	u := make([]byte, hashLen)
	ti := make([]byte, hashLen)

	for block := 1; block <= numBlocks; block++ {
		prf.Reset()
		prf.Write(salt)
		binary.BigEndian.PutUint32(buf[:], uint32(block))
		prf.Write(buf[:])
		u = prf.Sum(u[:0])
		copy(ti, u)

		for i := 2; i <= iter; i++ {
			prf.Reset()
			prf.Write(u)
			u = prf.Sum(u[:0])
			for j := range ti {
				ti[j] ^= u[j]
			}
		}
		dk = append(dk, ti...)
	}
	return dk[:keyLen]
}

// CheckPassword verifies plainPassword against a Werkzeug-compatible password hash.
func CheckPassword(hashedPassword, plainPassword string) bool {
	if hashedPassword == "" {
		return plainPassword == ""
	}

	parts := strings.Split(hashedPassword, "$")
	if len(parts) != 3 {
		return subtle.ConstantTimeCompare([]byte(hashedPassword), []byte(plainPassword)) == 1
	}

	methodParts := strings.Split(parts[0], ":")
	if len(methodParts) < 2 {
		return false
	}

	method := methodParts[0]
	hashName := methodParts[1]
	iter := 260000
	if len(methodParts) >= 3 {
		if parsedIter, err := strconv.Atoi(methodParts[2]); err == nil && parsedIter > 0 {
			iter = parsedIter
		}
	}

	if method != "pbkdf2" || hashName != "sha256" {
		return false
	}

	salt := parts[1]
	expectedHash := parts[2]

	dk := pbkdf2Key([]byte(plainPassword), []byte(salt), iter, len(expectedHash)/2, sha256.New)
	computedHex := hex.EncodeToString(dk)

	return subtle.ConstantTimeCompare([]byte(computedHex), []byte(expectedHash)) == 1
}

// HashPassword generates a Werkzeug-compatible PBKDF2:SHA256 password hash.
func HashPassword(plainPassword string) (string, error) {
	salt, err := generateRandomString(DefaultSaltLen)
	if err != nil {
		return "", err
	}

	dk := pbkdf2Key([]byte(plainPassword), []byte(salt), DefaultIter, KeyLen, sha256.New)
	return fmt.Sprintf("pbkdf2:sha256:%d$%s$%s", DefaultIter, salt, hex.EncodeToString(dk)), nil
}

// GenerateSessionKey produces a secure random hexadecimal session token.
func GenerateSessionKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func generateRandomString(n int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[num.Int64()]
	}
	return string(b), nil
}

// WithUser associates a User with the given context.
func WithUser(ctx context.Context, u *database.User) context.Context {
	return context.WithValue(ctx, UserContextKey, u)
}

// UserFromContext retrieves the User from the context, or nil if not present.
func UserFromContext(ctx context.Context) *database.User {
	if u, ok := ctx.Value(UserContextKey).(*database.User); ok {
		return u
	}
	return nil
}

// Permission bits in Calibre-Web
const (
	RoleAdmin       = 1
	RoleDownload    = 2
	RoleUpload      = 4
	RoleEdit        = 8
	RolePasswd      = 16
	RoleEditShelfs  = 32
	RoleDeleteBooks = 64
	RoleViewer      = 128
)

// HasPermission checks if a user has a specific Calibre-Web permission.
func HasPermission(u *database.User, perm int) bool {
	if u == nil {
		return false
	}
	if u.IsAdmin || (u.Role&RoleAdmin) != 0 {
		return true
	}
	return (u.Role & perm) != 0
}
