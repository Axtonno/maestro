// m35metadata records only review-relevant local Ollama model metadata.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type show struct {
	License, Modelfile, Parameters, Template, System string
	Capabilities                                     []string
	Details                                          struct {
		Family            string `json:"family"`
		ParameterSize     string `json:"parameter_size"`
		QuantizationLevel string `json:"quantization_level"`
	}
}
type record struct {
	Name, Digest, Family, ParameterSize, Quantization, TemplateSHA256, SystemSHA256, Parameters string
	Size                                                                                        int64
	Capabilities                                                                                []string
	ApacheLicense                                                                               bool
}

func main() {
	client := http.Client{Timeout: 30 * time.Second}
	var tags struct {
		Models []struct {
			Name, Digest string
			Size         int64
			Details      struct {
				Family            string `json:"family"`
				ParameterSize     string `json:"parameter_size"`
				QuantizationLevel string `json:"quantization_level"`
			}
		}
	}
	get(&client, "/api/tags", nil, &tags)
	byName := map[string]record{}
	for _, x := range tags.Models {
		byName[x.Name] = record{Name: x.Name, Digest: x.Digest, Size: x.Size, Family: x.Details.Family, ParameterSize: x.Details.ParameterSize, Quantization: x.Details.QuantizationLevel}
	}
	names := []string{"qwen2.5-coder:7b", "qwen2.5-coder:14b", "granite-code:8b-instruct"}
	out := []record{}
	for _, name := range names {
		x, ok := byName[name]
		if !ok {
			panic("missing " + name)
		}
		var s show
		get(&client, "/api/show", map[string]string{"model": name}, &s)
		x.TemplateSHA256 = hash(s.Template)
		x.SystemSHA256 = hash(s.System)
		x.Parameters = s.Parameters
		x.Capabilities = s.Capabilities
		x.ApacheLicense = strings.Contains(s.License, "Apache License")
		out = append(out, x)
	}
	b, err := json.MarshalIndent(struct {
		Provider, CollectedAt string
		Models                []record
	}{"ollama-0.33.1", time.Now().UTC().Format(time.RFC3339), out}, "", "  ")
	must(err)
	must(os.WriteFile("docs/reports/milestone-35-model-metadata.json", append(b, '\n'), 0600))
	fmt.Println("wrote 3 frozen metadata records")
}
func get(c *http.Client, path string, payload any, target any) {
	var body *bytes.Reader
	if payload == nil {
		body = bytes.NewReader(nil)
	} else {
		b, e := json.Marshal(payload)
		must(e)
		body = bytes.NewReader(b)
	}
	method := "GET"
	if payload != nil {
		method = "POST"
	}
	r, e := http.NewRequest(method, "http://127.0.0.1:11434"+path, body)
	must(e)
	if payload != nil {
		r.Header.Set("Content-Type", "application/json")
	}
	resp, e := c.Do(r)
	must(e)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic(resp.Status)
	}
	must(json.NewDecoder(resp.Body).Decode(target))
}
func hash(s string) string { x := sha256.Sum256([]byte(s)); return hex.EncodeToString(x[:]) }
func must(e error) {
	if e != nil {
		panic(e)
	}
}
