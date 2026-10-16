package atsutil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
)

const maxResponseBytes = 5 << 20

var (
	breakTags = regexp.MustCompile(`(?i)<\s*(br|/p|/div|/li|/h[1-6])\s*/?>`)
	allTags   = regexp.MustCompile(`<[^>]+>`)
	spaces    = regexp.MustCompile(`[ \t]+`)
	newlines  = regexp.MustCompile(`\n{3,}`)
)

func GetJSON(ctx context.Context, client *http.Client, endpoint string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "job-application-assistant/0.1")

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("fetch ATS posting: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		return fmt.Errorf("ATS returned HTTP %d", response.StatusCode)
	}

	decoder := json.NewDecoder(io.LimitReader(response.Body, maxResponseBytes))
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode ATS response: %w", err)
	}
	return nil
}

func HTMLToText(value string) string {
	value = breakTags.ReplaceAllString(value, "\n")
	value = allTags.ReplaceAllString(value, "")
	value = html.UnescapeString(value)
	lines := strings.Split(strings.ReplaceAll(value, "\r", ""), "\n")
	for index, line := range lines {
		lines[index] = strings.TrimSpace(spaces.ReplaceAllString(line, " "))
	}
	return strings.TrimSpace(newlines.ReplaceAllString(strings.Join(lines, "\n"), "\n\n"))
}

func ClientOrDefault(client *http.Client) *http.Client {
	if client == nil {
		return http.DefaultClient
	}
	return client
}

func DisplayName(slug string) string {
	words := strings.Fields(strings.NewReplacer("-", " ", "_", " ").Replace(slug))
	for index, word := range words {
		if word != "" {
			words[index] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

var ErrInvalidURL = errors.New("invalid or incomplete ATS job URL")
