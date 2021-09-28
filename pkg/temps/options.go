package temps

import "net/http"

type Option func(*Temps)

func WithTagListSecret(secret string) Option {
	return func(t *Temps) {
		t.secret = secret
	}
}

func WithDefaultHandler(handler http.Handler) Option {
	return func(t *Temps) {
		t.defaultHandler = handler
	}
}
