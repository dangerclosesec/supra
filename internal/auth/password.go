// internal/auth/password.go
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type PasswordConfig struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
}

type PasswordHasher struct {
	config PasswordConfig
}

func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{
		config: PasswordConfig{
			time:    1,
			memory:  64 * 1024,
			threads: 4,
			keyLen:  32,
		},
	}
}

func (p *PasswordHasher) Hash(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		p.config.time,
		p.config.memory,
		p.config.threads,
		p.config.keyLen,
	)

	// Format: $argon2id$v=19$m=65536,t=1,p=4$salt$hash
	encoded := fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		p.config.memory,
		p.config.time,
		p.config.threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

func (p *PasswordHasher) Verify(password, encodedHash string) (bool, error) {
	// Parse the encoded hash
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}

	var config PasswordConfig
	_, err := fmt.Sscanf(
		parts[3],
		"m=%d,t=%d,p=%d",
		&config.memory,
		&config.time,
		&config.threads,
	)
	if err != nil {
		return false, fmt.Errorf("invalid hash format: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("invalid salt: %w", err)
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("invalid hash: %w", err)
	}

	config.keyLen = uint32(len(decodedHash))

	// Compute hash with same parameters
	comparisonHash := argon2.IDKey(
		[]byte(password),
		salt,
		config.time,
		config.memory,
		config.threads,
		config.keyLen,
	)

	return subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1, nil
}
