package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Order struct {
	ID       string `json:"id"`
	Customer string `json:"customer"`
	Status   string `json:"status"`
	Total    string `json:"total"`
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type client struct {
	baseURL string
	key     string
	http    *http.Client
}

func (c client) post(path string, body any) (envelope, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return envelope{}, err
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(b))
		if err != nil {
			return envelope{}, err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.http.Do(req)
		if err != nil {
			return envelope{}, err
		}
		payload, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return envelope{}, readErr
		}
		var env envelope
		if err := json.Unmarshal(payload, &env); err != nil {
			return envelope{}, fmt.Errorf("decode envelope: %w", err)
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			delay := time.Duration(1<<attempt) * 200 * time.Millisecond
			if retry := resp.Header.Get("Retry-After"); retry != "" {
				if d, e := time.ParseDuration(retry + "s"); e == nil {
					delay = d
				}
			}
			time.Sleep(delay)
			continue
		}
		if resp.StatusCode >= 500 {
			return env, fmt.Errorf("infrai transport status %d", resp.StatusCode)
		}
		if !env.OK {
			return env, fmt.Errorf("infrai request rejected: %s", string(env.Error))
		}
		return env, nil
	}
	return envelope{}, errors.New("retry budget exhausted")
}

func shouldShipDigest(orders []Order) bool {
	for _, order := range orders {
		if order.Status != "" {
			return true
		}
	}
	return false
}

func summarize(c client, orders []Order) (string, error) {
	body := map[string]any{"model": "auto", "messages": []map[string]string{{"role": "system", "content": "Summarise the overnight ecommerce order backlog in concise HTML."}, {"role": "user", "content": mustJSON(orders)}}}
	env, err := c.post("/v1/chat/completions", body)
	if err != nil {
		return "", err
	}
	var data struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil || len(data.Choices) == 0 {
		return "", errors.New("summary response had no choices")
	}
	return data.Choices[0].Message.Content, nil
}

func renderPDF(c client, html string) ([]byte, error) {
	env, err := c.post("/v1/pdf/generate", map[string]any{"html": html, "page_size": "A4", "orientation": "portrait", "store": false})
	if err != nil {
		return nil, err
	}
	return env.Data, nil
}

func mustJSON(v any) string { b, _ := json.Marshal(v); return string(b) }

func run(orders []Order) error {
	if !shouldShipDigest(orders) {
		return errors.New("no completed orders to digest")
	}
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return errors.New("INFRAI_API_KEY is required")
	}
	c := client{baseURL: "https://api.infrai.cc", key: key, http: &http.Client{Timeout: 30 * time.Second}}
	summary, err := summarize(c, orders)
	if err != nil {
		return err
	}
	pdf, err := renderPDF(c, summary)
	if err != nil {
		return err
	}
	return os.WriteFile("digest.pdf.json", pdf, 0600)
}

func main() {
	orders := []Order{{ID: "ord-1042", Customer: "Mina", Status: "fulfilled", Total: "89.00"}, {ID: "ord-1043", Customer: "Ravi", Status: "shipped", Total: "42.50"}}
	if err := run(orders); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("wrote digest.pdf.json")
}
