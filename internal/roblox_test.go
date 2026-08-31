package internal

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func resp(status int, body string, headers http.Header) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: headers}
}

func TestGetGroups(t *testing.T) {
	client := &RobloxClient{http: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v2/groups" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if got := r.URL.Query().Get("groupIds"); got != "100,101,102" {
			t.Fatalf("groupIds=%s", got)
		}
		return resp(200, `{"data":[{"id":100,"name":"Test","owner":null,"publicEntryAllowed":true,"isLocked":false}]}`, http.Header{}), nil
	})}}
	groups, status, retry, err := client.GetGroups(context.Background(), []int{100, 101, 102})
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 || retry != 0 || len(groups) != 1 || groups[0].ID != 100 {
		t.Fatalf("unexpected result: %#v %d %s", groups, status, retry)
	}
}

func TestGetGroupsEmpty(t *testing.T) {
	c := NewRobloxClient(time.Second)
	_, status, _, err := c.GetGroups(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if status != 400 {
		t.Fatalf("status=%d", status)
	}
}

func TestGetGroups429(t *testing.T) {
	h := http.Header{}
	h.Set("Retry-After", "3")
	c := &RobloxClient{http: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return resp(429, "", h), nil })}}
	_, status, retry, err := c.GetGroups(context.Background(), []int{1})
	if err != nil {
		t.Fatal(err)
	}
	if status != 429 || retry != 3*time.Second {
		t.Fatalf("status=%d retry=%s", status, retry)
	}
}

func TestGetGroups429Fallback(t *testing.T) {
	c := &RobloxClient{http: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return resp(429, "", http.Header{}), nil })}}
	_, status, retry, err := c.GetGroups(context.Background(), []int{1})
	if err != nil {
		t.Fatal(err)
	}
	if status != 429 || retry != time.Second {
		t.Fatalf("status=%d retry=%s", status, retry)
	}
}

func TestGetGroupsInvalidJSON(t *testing.T) {
	c := &RobloxClient{http: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return resp(200, "{invalid", http.Header{}), nil })}}
	_, _, _, err := c.GetGroups(context.Background(), []int{1})
	if err == nil {
		t.Fatal("expected decode error")
	}
}
func TestGetGroupsNetworkError(t *testing.T) {
	c := &RobloxClient{http: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, errors.New("network failure") })}}
	_, _, _, err := c.GetGroups(context.Background(), []int{1})
	if err == nil || !strings.Contains(err.Error(), "network failure") {
		t.Fatalf("err=%v", err)
	}
}
