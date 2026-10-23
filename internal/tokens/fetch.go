package tokens

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pglevy/mint/internal/errs"
	"github.com/pglevy/mint/internal/jsonv"
)

// URL is where the Sailwind design tokens are published. It is a variable so
// that builds can point it elsewhere with -ldflags "-X".
var URL = "https://raw.githubusercontent.com/pglevy/sailwind/main/public/tokens.json"

const fetchTimeout = 30 * time.Second

const networkHint = "Check your network connection and try again."

// Fetch downloads and parses the design tokens at source. It returns an
// *errs.Error on network failure, a non-2xx response or malformed JSON.
func Fetch(ctx context.Context, source string) (jsonv.Value, error) {
	ctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return nil, errs.Fetch(fmt.Sprintf("Failed to fetch tokens from %s: %s", source, reason(err)), networkHint, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errs.Fetch(fmt.Sprintf("Failed to fetch tokens from %s: %s", source, reason(err)), networkHint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		status := strings.TrimSpace(fmt.Sprintf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode)))
		return nil, errs.Fetch(fmt.Sprintf("Failed to fetch tokens from %s: HTTP %s", source, status), networkHint, nil)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errs.Fetch(fmt.Sprintf("Failed to fetch tokens from %s: %s", source, reason(err)), networkHint, err)
	}
	doc, err := jsonv.Parse(body)
	if err != nil {
		return nil, errs.Fetch(fmt.Sprintf("Failed to parse tokens from %s: %s", source, err), "The token source did not return valid JSON.", err)
	}
	if _, ok := doc.(*jsonv.Object); !ok {
		return nil, errs.Fetch(fmt.Sprintf("Failed to parse tokens from %s: expected a JSON object", source), "The token source did not return a token document.", nil)
	}
	return doc, nil
}

// reason returns err's message without the `Get "url": ` prefix that
// net/http adds, since the caller already names the URL.
func reason(err error) string {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return urlErr.Err.Error()
	}
	return err.Error()
}
