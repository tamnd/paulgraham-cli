package paulgraham_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/tamnd/paulgraham-cli/paulgraham"
)

func newTestClient(ts *httptest.Server) *paulgraham.Client {
	cfg := paulgraham.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return paulgraham.NewClient(cfg)
}

func TestGetSendsUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	cfg := paulgraham.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	c = paulgraham.NewClient(cfg)

	// Use ListEssays to exercise the GET path via the client
	// articles.html returns empty list but no error
	_, _ = c.ListEssays(context.Background())
}

func TestGetRetriesOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`<html><body><a href="essays.html">Essays</a></body></html>`))
	}))
	defer srv.Close()

	cfg := paulgraham.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := paulgraham.NewClient(cfg)

	start := time.Now()
	_, err := c.ListEssays(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("retries did not back off")
	}
}

func TestListEssays(t *testing.T) {
	articlesHTML := `<html><body>
<a href="startupideas.html">How to Get Startup Ideas</a>
<a href="greatwork.html">How to Do Great Work</a>
<a href="http://external.com/page.html">External</a>
</body></html>`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(articlesHTML))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	essays, err := c.ListEssays(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(essays) != 2 {
		t.Fatalf("got %d essays, want 2", len(essays))
	}
	if essays[0].Slug != "startupideas" {
		t.Errorf("slug = %q, want startupideas", essays[0].Slug)
	}
	if essays[0].Title != "How to Get Startup Ideas" {
		t.Errorf("title = %q", essays[0].Title)
	}
	if !strings.HasSuffix(essays[0].URL, "/startupideas.html") {
		t.Errorf("url = %q", essays[0].URL)
	}
}

func TestGetEssay(t *testing.T) {
	essayHTML := `<html><head><title>How to Get Startup Ideas</title></head>
<body>
<h3>How to Get Startup Ideas</h3>
<font face=verdana>
<p>The way to get startup ideas is not to try to think of startup ideas.</p>
</font>
</body></html>`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(essayHTML))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	content, err := c.GetEssay(context.Background(), "startupideas")
	if err != nil {
		t.Fatal(err)
	}
	if content.Slug != "startupideas" {
		t.Errorf("slug = %q", content.Slug)
	}
	if content.Title != "How to Get Startup Ideas" {
		t.Errorf("title = %q", content.Title)
	}
	if !strings.Contains(content.Body, "startup ideas") {
		t.Errorf("body missing expected text, got: %q", content.Body)
	}
}
