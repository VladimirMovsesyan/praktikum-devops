package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func Get(data, key string) string {
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write([]byte(data))
	return hex.EncodeToString(hash.Sum(nil))
}

func Valid(hash, data, key string) bool {
	return hash == Get(data, key)
}
