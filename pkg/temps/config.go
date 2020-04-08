package temps

type Config struct {
	Secret                string `required:"true"`
	WundergroundAPIKey    string `required:"true" split_words:"true"`
	WundergroundStationID string `default:"KINKIRKL2"`
}
