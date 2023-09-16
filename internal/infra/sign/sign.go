package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

func Sign(msg []byte, key []byte) ([]byte, error) {
	value := sha256.Sum256(msg)
	h := hmac.New(sha256.New, key)
	if _, err := h.Write(value[:]); err != nil {
		return nil, fmt.Errorf("Sign: write hash failed %w", err)
	}
	return h.Sum(nil), nil
}
