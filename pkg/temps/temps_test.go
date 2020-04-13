package temps

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetWSURL(t *testing.T) {
	examples := map[string]struct {
		request       *http.Request
		expectedWSURL string
	}{
		"dev": {
			request: &http.Request{
				Host: "127.0.0.1:8080",
				URL:  &url.URL{Path: "/"},
			},
			expectedWSURL: "ws://127.0.0.1:8080/live",
		},
		"prod": {
			request: &http.Request{
				Host: "temps.whatever.com",
				URL:  &url.URL{Path: "/"},
			},
			expectedWSURL: "ws://temps.whatever.com/live",
		},
	}

	for label, ex := range examples {
		t.Run(label, func(tt *testing.T) {
			assert.Equal(t, ex.expectedWSURL, getWSURL(ex.request))
		})
	}
}
