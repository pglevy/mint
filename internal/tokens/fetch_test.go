package tokens

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pglevy/mint/internal/errs"
)

func serve(t *testing.T, handler http.HandlerFunc) string {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv.URL
}

func fetchError(t *testing.T, err error) *errs.Error {
	t.Helper()
	var e *errs.Error
	if !errors.As(err, &e) || e.Kind != "FetchError" {
		t.Fatalf("error = %#v, want a FetchError", err)
	}
	return e
}

func TestFetchParsesTheDocument(t *testing.T) {
	url := serve(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"color":{"red":{"$value":"#f00","$type":"color"}}}`))
	})
	doc, err := Fetch(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	store := NewStore()
	if counts := store.Load(doc); counts[Color] != 1 {
		t.Errorf("counts = %v", counts)
	}
}

func TestFetchFollowsRedirects(t *testing.T) {
	var base string
	base = serve(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/final" {
			w.Write([]byte(`{}`))
			return
		}
		http.Redirect(w, r, base+"/final", http.StatusFound)
	})
	if _, err := Fetch(context.Background(), base+"/start"); err != nil {
		t.Fatal(err)
	}
}

func TestFetchIgnoresAByteOrderMark(t *testing.T) {
	url := serve(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("\xef\xbb\xbf{}"))
	})
	if _, err := Fetch(context.Background(), url); err != nil {
		t.Fatal(err)
	}
}

func TestFetchHTTPError(t *testing.T) {
	url := serve(t, func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) })
	_, err := Fetch(context.Background(), url)
	e := fetchError(t, err)
	if want := "Failed to fetch tokens from " + url + ": HTTP 404 Not Found"; e.Message != want {
		t.Errorf("message = %q, want %q", e.Message, want)
	}
	if e.Suggestion != "Check your network connection and try again." {
		t.Errorf("suggestion = %q", e.Suggestion)
	}
}

func TestFetchUsesTheCanonicalStatusTextNotTheServers(t *testing.T) {
	url := serve(t, func(w http.ResponseWriter, r *http.Request) {
		// Raw status line with a custom reason phrase.
		conn, buf, _ := w.(http.Hijacker).Hijack()
		defer conn.Close()
		buf.WriteString("HTTP/1.1 503 Try Later Please\r\nContent-Length: 0\r\n\r\n")
		buf.Flush()
	})
	_, err := Fetch(context.Background(), url)
	if e := fetchError(t, err); !strings.HasSuffix(e.Message, "HTTP 503 Service Unavailable") {
		t.Errorf("message = %q", e.Message)
	}
}

func TestFetchConnectionFailure(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close() // nothing is listening any more
	_, err := Fetch(context.Background(), url)
	e := fetchError(t, err)
	if !strings.HasPrefix(e.Message, "Failed to fetch tokens from "+url+": ") {
		t.Errorf("message = %q", e.Message)
	}
	// The reason replaces net/http's `Get "url": ` prefix, which would repeat the URL.
	if strings.Contains(e.Message, `Get "`) {
		t.Errorf("message repeats the request: %q", e.Message)
	}
	if e.Err == nil {
		t.Error("cause was dropped")
	}
}

func TestFetchRejectsDocumentsThatAreNotTokenFiles(t *testing.T) {
	for _, body := range []string{"", "not json", "<!DOCTYPE html>", "[]", "null", `"x"`, "42", `{"color":`} {
		url := serve(t, func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(body)) })
		_, err := Fetch(context.Background(), url)
		if e := fetchError(t, err); !strings.HasPrefix(e.Message, "Failed to parse tokens from "+url) {
			t.Errorf("body %q: message = %q", body, e.Message)
		}
	}
}

func TestFetchHonoursContextCancellation(t *testing.T) {
	url := serve(t, func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Fetch(ctx, url); err == nil {
		t.Fatal("expected an error")
	} else {
		fetchError(t, err)
	}
}
