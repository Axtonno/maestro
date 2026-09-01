package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type capture struct {
	Method         string          `json:"method"`
	Path           string          `json:"path"`
	SHA256         string          `json:"sha256"`
	Body           json.RawMessage `json:"body"`
	ResponseStatus int             `json:"response_status"`
	ResponseSHA256 string          `json:"response_sha256"`
	ResponseBody   json.RawMessage `json:"response_body"`
}

func main() {
	listen := flag.String("listen", "127.0.0.1:11435", "relay listen address")
	upstreamValue := flag.String("upstream", "http://127.0.0.1:11434", "Ollama origin")
	output := flag.String("output", "m23-relay.jsonl", "capture output")
	flag.Parse()

	upstream, err := url.Parse(*upstreamValue)
	if err != nil || upstream.Scheme == "" || upstream.Host == "" {
		log.Fatal("invalid upstream")
	}
	file, err := os.OpenFile(*output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	client := &http.Client{Timeout: 10 * time.Minute}

	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(io.LimitReader(request.Body, 16<<20))
		if err != nil {
			http.Error(writer, "relay read failure", http.StatusBadGateway)
			return
		}
		var entry *capture
		if request.URL.Path == "/api/chat" {
			digest := sha256.Sum256(body)
			entry = &capture{Method: request.Method, Path: request.URL.Path, SHA256: hex.EncodeToString(digest[:]), Body: append([]byte(nil), body...)}
		}
		target := *upstream
		target.Path = request.URL.Path
		target.RawQuery = request.URL.RawQuery
		forward, err := http.NewRequestWithContext(request.Context(), request.Method, target.String(), bytes.NewReader(body))
		if err != nil {
			http.Error(writer, "relay request failure", http.StatusBadGateway)
			return
		}
		forward.Header = request.Header.Clone()
		response, err := client.Do(forward)
		if err != nil {
			http.Error(writer, "relay upstream failure", http.StatusBadGateway)
			return
		}
		defer response.Body.Close()
		responseBody, err := io.ReadAll(io.LimitReader(response.Body, 16<<20))
		if err != nil {
			http.Error(writer, "relay response failure", http.StatusBadGateway)
			return
		}
		if entry != nil {
			digest := sha256.Sum256(responseBody)
			entry.ResponseStatus = response.StatusCode
			entry.ResponseSHA256 = hex.EncodeToString(digest[:])
			entry.ResponseBody = append([]byte(nil), responseBody...)
			if err := json.NewEncoder(file).Encode(entry); err != nil {
				http.Error(writer, "relay capture failure", http.StatusBadGateway)
				return
			}
			if err := file.Sync(); err != nil {
				http.Error(writer, "relay sync failure", http.StatusBadGateway)
				return
			}
		}
		for name, values := range response.Header {
			for _, value := range values {
				writer.Header().Add(name, value)
			}
		}
		writer.WriteHeader(response.StatusCode)
		_, _ = writer.Write(responseBody)
	})

	fmt.Printf("m23 relay listening on %s\n", *listen)
	server := &http.Server{Addr: *listen, Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	log.Fatal(server.ListenAndServe())
}
