package temps

type Option func(*Temps)

func WithTagListSecret(secret string) Option {
	return func(t *Temps) {
		t.secret = secret
	}
}

func WithWU(weather WeatherClient) Option {
	return func(t *Temps) {
		t.weather = weather
	}
}
