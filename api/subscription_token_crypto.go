package api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"sublink/middlewares"
)

func subscriptionTokenCipher() (cipher.AEAD, error) {
	secret := middlewares.SecretBytes()
	if len(secret) == 0 {
		return nil, errors.New("JWT secret is not configured")
	}
	keyMaterial := append([]byte("sublink-subscription-token-v1:"), secret...)
	key := sha256.Sum256(keyMaterial)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func encryptSubscriptionToken(token string) (string, error) {
	aead, err := subscriptionTokenCipher()
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := aead.Seal(nonce, nonce, []byte(token), nil)
	return base64.RawURLEncoding.EncodeToString(sealed), nil
}

func decryptSubscriptionToken(encoded string) (string, error) {
	aead, err := subscriptionTokenCipher()
	if err != nil {
		return "", err
	}
	sealed, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	nonceSize := aead.NonceSize()
	if len(sealed) < nonceSize {
		return "", errors.New("invalid encrypted subscription token")
	}
	token, err := aead.Open(nil, sealed[:nonceSize], sealed[nonceSize:], nil)
	if err != nil {
		return "", err
	}
	return string(token), nil
}
