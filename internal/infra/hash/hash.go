// Пакет hash содержит объекты и методы для формирования хеша.
package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/pavlegich/metrics-alerting/internal/entities"
)

// Sign формирует хеш из сообщения ключа.
func Sign(msg []byte, key []byte) ([]byte, error) {
	value := sha256.Sum256(msg)
	h := hmac.New(sha256.New, key)
	if _, err := h.Write(value[:]); err != nil {
		return nil, fmt.Errorf("Sign: write hash failed %w", err)
	}
	return h.Sum(nil), nil
}

// SigningResponseWriter содержит реализацию http.ResponseWriter.
type SigningResponseWriter struct {
	http.ResponseWriter
}

// Write реализует запись ответа, формирование и размещение хеша
// в заголовок ответа.
func (r *SigningResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	if err != nil {
		return size, fmt.Errorf("Write: response write %w", err)
	}
	hash, err := Sign([]byte(""), []byte(entities.Key))
	if err != nil {
		return size, fmt.Errorf("Write: sign message failed %w", err)
	}
	r.Header().Set("HashSHA256", hex.EncodeToString(hash))
	return size, nil
}
