// Пакет crypto содержит методы шифрования
// сообщений с помощью ключей
package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// Encrypt зашифровывает сообщение с помощью ключа из файла
func Encrypt(msg []byte, keyPath string) ([]byte, error) {
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("Encrypt: read file error %w", err)
	}

	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, fmt.Errorf("Encrypt: find PEM data error %w", err)
	}

	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Encrypt: parse public key failed %w", err)
	}

	encryptedBytes, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		key,
		msg,
		nil)
	if err != nil {
		return nil, fmt.Errorf("Encrypt: encrypt message error %w", err)
	}

	return encryptedBytes, nil
}

// Decrypt расшифровывает сообщение ключом из файла
func Decrypt(msg []byte, keyPath string) ([]byte, error) {
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("Encrypt: read file error %w", err)
	}

	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, fmt.Errorf("Encrypt: find PEM data error %w", err)
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Encrypt: parse public key failed %w", err)
	}

	decryptedBytes, err := key.Decrypt(
		nil,
		msg,
		&rsa.OAEPOptions{
			Hash: crypto.SHA256,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("Decrypt: decrypt message error %w", err)
	}

	return decryptedBytes, nil
}
