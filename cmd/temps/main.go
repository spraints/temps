package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/kelseyhightower/envconfig"

	"github.com/spraints/temps/pkg/temps"
)

type Config struct {
	ListenAddr string `default:"127.0.0.1:8080" split_words:"true"`
	temps.Config
}

func main() {
	var cfg Config

	if err := envconfig.Process("TEMPS", &cfg); err != nil {
		log.Fatal(err)
	}

	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, syscall.SIGINT, syscall.SIGTERM)

	svc := temps.New(cfg.Config)

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
