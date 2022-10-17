package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/kelseyhightower/envconfig"

	"github.com/spraints/temps/pkg/filestore"
	"github.com/spraints/temps/pkg/memorystore"
	"github.com/spraints/temps/pkg/templates"
	"github.com/spraints/temps/pkg/temps"
	"github.com/spraints/temps/pkg/types"
	"github.com/spraints/temps/pkg/wu"
)

type Config struct {
	ListenAddr            string  `split_words:"true" default:"127.0.0.1:8080"`
	TagListSecret         string  `split_words:"true" required:"true"`
	WundergroundAPIKey    string  `split_words:"true"`
	WundergroundStationID string  `split_words:"true" default:"KINKIRKL2"`
	FakeOutdoorTemp       float64 `split_words:"true"`

	PublicPath      string `split_words:"true" default:"public"`
	TemplatesPath   string `split_words:"true" default:"templates"`
	ReloadTemplates bool   `split_words:"true" default:"false"`

	DataDir string `split_words:"true"`
}

func main() {
	var cfg Config

	if err := envconfig.Process("TEMPS", &cfg); err != nil {
		log.Fatal(err)
	}

	assetTag := fmt.Sprint(time.Now().Unix())

	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, syscall.SIGINT, syscall.SIGTERM)

	var weather temps.WeatherClient
	if cfg.WundergroundAPIKey != "" && cfg.WundergroundStationID != "" {
		weather = wu.New(cfg.WundergroundAPIKey, cfg.WundergroundStationID)
	} else {
		log.Printf("No weather underground API key or station ID was provided, using fixed outdoor temp (%.0f)", cfg.FakeOutdoorTemp)
		weather = fakeWeather(cfg.FakeOutdoorTemp)
	}

	var store temps.Store
	if cfg.DataDir != "" {
		store = filestore.New(path.Join(cfg.DataDir, "temps.json"))
	} else {
		store = memorystore.New()
	}

	svc := temps.New(
		weather,
		store,
		templates.New(cfg.TemplatesPath, cfg.ReloadTemplates, assetTag),
		temps.WithTagListSecret(cfg.TagListSecret),
	)

	var shutdown sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	run(ctx, &shutdown, "poll current temperature", svc.Poll)
	run(ctx, &shutdown, "http server on "+cfg.ListenAddr, newHTTPServer(&cfg, svc))
	<-stopSignal
	cancel()
	wait(10*time.Second, &shutdown)
}

func run(ctx context.Context, wg *sync.WaitGroup, label string, runFn func(context.Context)) {
	log.Printf("%s: starting", label)
	wg.Add(1)
	go func() {
		defer wg.Done()
		runFn(ctx)
		log.Printf("%s: finished", label)
	}()
}

func wait(timeout time.Duration, wg *sync.WaitGroup) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		wg.Wait()
	}()
	select {
	case <-done:
		return
	case <-time.After(timeout):
		log.Print("shut down before all threads finished")
	}
}

type service interface {
	Register(mux chi.Router)
}

func newHTTPServer(cfg *Config, services ...service) func(context.Context) {
	mux := chi.NewRouter()

	for _, svc := range services {
		svc.Register(mux)
	}
	mux.Mount("/", http.FileServer(http.Dir(cfg.PublicPath)))

	server := http.Server{
		Addr:    cfg.ListenAddr,
		Handler: mux,
	}

	return func(ctx context.Context) {
		go func() {
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatal(err)
			}
		}()
		<-ctx.Done()
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("failed to shut down HTTP server cleanly: %v", err)
		}

	}
}

type fakeWeather float64

func (f fakeWeather) GetCurrentConditions(ctx context.Context) (*wu.Conditions, error) {
	return &wu.Conditions{Temperature: types.Fahrenheit(f)}, nil
}
