package temps

type Option func(*Temps)

func WithTagListSecret(secret string) Option {
	return func(t *Temps) {
		t.secret = secret
	}
}
