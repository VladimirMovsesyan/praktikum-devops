package middleware

import (
	"compress/gzip"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	writer io.Writer
}

func (gw gzipWriter) Write(p []byte) (int, error) {
	return gw.writer.Write(p)
}

func Compress(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(rw http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(rw, r)
				return
			}

			newWriter := gzip.NewWriter(rw)
			defer newWriter.Close()

			writer := gzipWriter{
				ResponseWriter: rw,
				writer:         newWriter,
			}

			writer.Header().Set("Content-Encoding", "gzip")
			next.ServeHTTP(writer, r)
		},
	)
}

func Decompress(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(rw http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
				next.ServeHTTP(rw, r)
				return
			}

			reader, err := gzip.NewReader(r.Body)
			if err != nil {
				log.Println(err)
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer reader.Close()

			r.Body = reader
			next.ServeHTTP(rw, r)
		},
	)
}

func SubnetCheck(subnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(rw http.ResponseWriter, r *http.Request) {
				if subnet == "" {
					next.ServeHTTP(rw, r)
					return
				}

				ipHeader := r.Header.Get("X-Real-IP")
				ip := net.ParseIP(ipHeader)

				_, ipv4Net, err := net.ParseCIDR(subnet)
				if err != nil {
					log.Println(err)
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				if !ipv4Net.Contains(ip) {
					log.Printf("Client IP: '%s' is not in subnet: '%s'\n", ipHeader, subnet)
					rw.WriteHeader(http.StatusForbidden)
					return
				}

				next.ServeHTTP(rw, r)
			},
		)
	}
}
