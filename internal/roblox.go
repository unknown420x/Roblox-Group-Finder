package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultMinID = 1_000_000
	DefaultMaxID = 1_150_000
)

type RobloxClient struct{ http *http.Client }

func NewRobloxClient(timeout time.Duration) *RobloxClient {
	transport := &http.Transport{MaxIdleConns: 64, MaxIdleConnsPerHost: 64, MaxConnsPerHost: 64, IdleConnTimeout: 90 * time.Second, ForceAttemptHTTP2: true}
	return &RobloxClient{http: &http.Client{Transport: transport, Timeout: timeout}}
}

func (c *RobloxClient) GetGroups(ctx context.Context, ids []int) ([]Group, int, time.Duration, error) {
	if len(ids) == 0 {
		return nil, http.StatusBadRequest, 0, nil
	}
	encoded := make([]string, len(ids))
	for i, id := range ids {
		encoded[i] = strconv.Itoa(id)
	}
	q := url.Values{}
	q.Set("groupIds", strings.Join(encoded, ","))
	endpoint := fmt.Sprintf("https://groups.roblox.com/v2/groups?%s", q.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, 0, 0, err
	}
	req.Header.Set("User-Agent", "AleksGroupFinder/1.0")
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, resp.StatusCode, retryAfter(resp), nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.StatusCode, 0, nil
	}
	var result GroupsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, resp.StatusCode, 0, fmt.Errorf("decode Roblox response: %w", err)
	}
	return result.Data, resp.StatusCode, 0, nil
}

func retryAfter(resp *http.Response) time.Duration {
	value := strings.TrimSpace(resp.Header.Get("Retry-After"))
	if value == "" {
		if reset := strings.TrimSpace(resp.Header.Get("x-ratelimit-reset")); reset != "" {
			if seconds, err := strconv.Atoi(reset); err == nil && seconds >= 0 {
				if seconds > 60 {
					seconds = 60
				}
				return time.Duration(seconds) * time.Second
			}
		}
		return time.Second
	}
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		if seconds > 60 {
			seconds = 60
		}
		return time.Duration(seconds) * time.Second
	}
	if date, err := http.ParseTime(value); err == nil {
		d := time.Until(date)
		if d < 0 {
			d = 0
		}
		if d > 60*time.Second {
			d = 60 * time.Second
		}
		return d
	}
	return time.Second
}
