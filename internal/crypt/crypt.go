package crypt

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"io"
	"log"
	"net/http"
	"os"
)

type Crypter interface {
	Encrypt(data []byte) ([]byte, error)
	GetDecryptMiddleware() func(next http.Handler) http.Handler
}

type crypt struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func New(opts ...func(c *crypt) error) (Crypter, error) {
	c := &crypt{}
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	return c, nil
}

func WithPrivateKey(keyPath string) func(c *crypt) error {
	return func(c *crypt) error {
		data, err := os.ReadFile(keyPath)
		block, _ := pem.Decode(data)

		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			log.Println(err)
			return err
		}

		c.privateKey = key
		return nil
	}
}

func WithPublicKey(keyPath string) func(c *crypt) error {
	return func(c *crypt) error {
		data, err := os.ReadFile(keyPath)
		block, _ := pem.Decode(data)

		key, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			log.Println(err)
			return err
		}

		c.publicKey = key
		return nil
	}
}

func (c *crypt) Encrypt(data []byte) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, c.publicKey, data, []byte(""))
}

func (c *crypt) GetDecryptMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(rw http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					log.Println(err)
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, c.privateKey, body, []byte(""))
				if err != nil {
					log.Println(err)
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				r.Body = io.NopCloser(bytes.NewReader(decrypted))
				r.ContentLength = int64(len(decrypted))

				next.ServeHTTP(rw, r)
			},
		)
	}
}
