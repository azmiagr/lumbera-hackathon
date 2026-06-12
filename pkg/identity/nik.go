package identity

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"strings"
)

func NormalizeNIK(nik string) string {
	return strings.TrimSpace(nik)
}

func HashNIK(nik string) (string, error) {
	secret := os.Getenv("NIK_HASH_SECRET")
	if strings.TrimSpace(secret) == "" {
		return "", errors.New("NIK_HASH_SECRET is not configured")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(NormalizeNIK(nik)))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func EncryptNIK(nik string) (string, error) {
	rawKey := os.Getenv("NIK_ENCRYPTION_KEY")
	key, err := base64.StdEncoding.DecodeString(rawKey)
	if err != nil || len(key) != 32 {
		return "", errors.New("NIK_ENCRYPTION_KEY must be base64 encoded 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(NormalizeNIK(nik)), nil)
	payload := append(nonce, ciphertext...)
	return base64.StdEncoding.EncodeToString(payload), nil
}
