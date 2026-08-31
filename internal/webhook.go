package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type WebhookClient struct {
	url    string
	client *http.Client
	queue  chan int
}
type webhookPayload struct {
	Content string `json:"content"`
}

func NewWebhookClient(url string, timeout time.Duration) *WebhookClient {
	transport := &http.Transport{MaxIdleConns: 4, MaxIdleConnsPerHost: 4, IdleConnTimeout: 90 * time.Second, ForceAttemptHTTP2: true}
	w := &WebhookClient{url: url, client: &http.Client{Transport: transport, Timeout: timeout}, queue: make(chan int, 64)}
	go w.run()
	return w
}

func (w *WebhookClient) Queue(id int) {
	select {
	case w.queue <- id:
	default:
	}
}

func (w *WebhookClient) run() {
	for id := range w.queue {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_ = w.send(ctx, id)
		cancel()
	}
}

func (w *WebhookClient) send(ctx context.Context, id int) error {
	data, err := json.Marshal(webhookPayload{Content: fmt.Sprintf("Hit: https://www.roblox.com/groups/group.aspx?gid=%s", strconv.Itoa(id))})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned HTTP %d", resp.StatusCode)
	}
	return nil
}
