package utils

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

func Value[T any](first T, _ ...any) T {
	return first
}

// Config reads an environment variable first, then a local .config JSON file.
func Config(name string) (value string) {
	if value = os.Getenv(strings.ToUpper(name)); value != "" {
		return value
	}

	file, err := os.ReadFile(".config")
	if err != nil {
		return ""
	}

	config := map[string]json.RawMessage{}
	if err = json.Unmarshal(file, &config); err != nil {
		log.Printf("parse .config: %v", err)
		return ""
	}

	raw, ok := config[name]
	if !ok {
		return ""
	}
	if json.Unmarshal(raw, &value) == nil {
		return value
	}
	return string(raw)
}

var reLinks = regexp.MustCompile(`https?://(?:[^/\s]+(?:\.|\b))*(/[^\s!$'()*,:;\[\]]*)?`)

func ExtractUrls(s string) []*url.URL {
	matches := reLinks.FindAllString(s, -1)
	urls := make([]*url.URL, 0, len(matches))
	occur := make(map[string]struct{})

	for _, match := range matches {
		u, err := url.Parse(match)
		if err != nil {
			continue
		}

		escaped := u.EscapedPath()
		for strings.Contains(escaped, "//") {
			escaped = strings.ReplaceAll(escaped, "//", "/")
		}
		if path, err := url.PathUnescape(escaped); err == nil {
			u.RawPath = escaped
			u.Path = path
		}

		key := u.String()
		if _, ok := occur[key]; !ok {
			occur[key] = struct{}{}
			urls = append(urls, u)
		}
	}
	return urls
}

func NewClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = 200
	transport.MaxConnsPerHost = 10
	transport.MaxIdleConnsPerHost = 10
	transport.IdleConnTimeout = 5 * time.Minute

	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

var Client = NewClient()

var reRefreshMeta = regexp.MustCompile(`(?im)<meta\s.*?http-equiv\s*=\s*['"\s]*?refresh['"\s]*?.*?>`)
var reRefreshURL = regexp.MustCompile(`(?i);\s*URL=(.+?)['"\s]`)

func SeekLocation(u *url.URL) *url.URL {
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := Client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if location := resp.Header.Get("Location"); location != "" {
		parsed, err := url.Parse(location)
		if err == nil && parsed.String() != u.String() {
			return parsed
		}
		return nil
	}

	match := reRefreshMeta.FindSubmatch(Value(io.ReadAll(resp.Body)))
	if len(match) < 1 {
		return nil
	}
	match = reRefreshURL.FindSubmatch(match[0])
	if len(match) < 2 {
		return nil
	}
	parsed, err := url.Parse(string(match[1]))
	if err != nil || parsed.String() == u.String() {
		return nil
	}
	return parsed
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandString(n int) string {
	result := make([]byte, n)
	for i := range result {
		result[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(result)
}
