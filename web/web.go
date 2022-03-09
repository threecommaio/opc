// Package web provides functionality for HTTP, Templating, and RPC
package web

import (
	// force imports.
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/Masterminds/sprig"
	_ "github.com/dghubble/sling"
	"github.com/dustin/go-humanize"
	_ "github.com/foolin/goview"
	sentry "github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	_ "github.com/gobuffalo/validate"
	_ "github.com/gorilla/schema"
	_ "github.com/gotailwindcss/tailwind"
	_ "github.com/joncalhoun/form"
	log "github.com/sirupsen/logrus"
	"github.com/threecommaio/opc/core"
	"github.com/threecommaio/opc/version"
	ginlogrus "github.com/toorop/gin-logrus"
	_ "google.golang.org/grpc"
)

var (
	ErrSentryDisabled = errors.New("sentry disabled in development environment")
	ErrSentryEmpty    = errors.New("sentry dsn is empty")
)

// Srv is the web server
type Srv struct {
	cfg       SrvConfig
	quit      chan os.Signal
	sentryDSN string
	server    *http.Server
}

// Option is used for configuring features of the webserver
type Option func(*Srv)

// SrvConfig is the configuration for the web server
type SrvConfig struct {
	ListenAddress string
	ReadTimeout   string
	WriteTimeout  string
}

func configureSentry(dsn string) error {
	debug := os.Getenv("SENTRY_DEBUG")
	sr := os.Getenv("SENTRY_SAMPLE_RATE")
	if sr == "" {
		sr = "0.2" // 20% sample rate is the default
	}
	sampleRate, err := strconv.ParseFloat(sr, 64)
	if err != nil {
		log.Fatalf("failed to parse SENTRY_SAMPLE_RATE: %s", err)
	}
	if sampleRate < 0 || sampleRate > 1 {
		log.Fatal("sentry sample rate must be between 0 and 1.0")
	}

	if core.Environment() == core.Development {
		return ErrSentryDisabled
	}

	if dsn == "" {
		return ErrSentryEmpty
	}

	// To initialize Sentry's handler, you need to initialize Sentry itself beforehand
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Release:          version.Release(),
		Environment:      core.Environment(),
		Debug:            debug == "true",
		TracesSampleRate: sampleRate,
	}); err != nil {
		return err
	}

	return nil
}

// New creates the webserver
func New(cfg SrvConfig, opts ...Option) (Srv, *gin.Engine, error) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Setup the gin router
	router := gin.New()
	router.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))
	router.Use(ginlogrus.Logger(log.StandardLogger()), gin.Recovery())
	// attach healthcheck
	router.GET("/health", Healthz())

	readTimeout, err := time.ParseDuration(cfg.ReadTimeout)
	if err != nil {
		return Srv{}, nil, fmt.Errorf("failed to parse read timeout: %w", err)
	}
	writeTimeout, err := time.ParseDuration(cfg.WriteTimeout)
	if err != nil {
		return Srv{}, nil, fmt.Errorf("failed to parse write timeout: %w", err)
	}

	server := &http.Server{
		Addr:           cfg.ListenAddress,
		Handler:        router,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 24 * humanize.KiByte,
	}

	srv := &Srv{
		cfg:    cfg,
		server: server,
	}

	opts = append(opts, WithQuit(quit))
	// Loop through each option
	for _, opt := range opts {
		opt(srv)
	}

	return *srv, router, nil
}

// WithQuit sets the quit channel to receive a signal to stop the bot
func WithQuit(quit chan os.Signal) Option {
	return func(s *Srv) {
		s.quit = quit
	}
}

// WithSentryDSN sets the sentry dsn
func WithSentryDSN(dsn string) Option {
	return func(s *Srv) {
		s.sentryDSN = dsn
	}
}

// Start starts the web server
func (s *Srv) Start() error {
	// Flush buffered events before the program terminates
	// Set the timeout to the maximum duration the program can afford to wait
	defer sentry.Flush(5 * time.Second)

	log.Infof("build release: %s", version.Release())

	// configure sentry
	err := configureSentry(s.sentryDSN)
	if err != nil {
		switch true {
		case errors.Is(err, ErrSentryDisabled):
			log.Warn("sentry disabled in development environment")
		case errors.Is(err, ErrSentryEmpty):
			log.Warn("sentry dsn is empty for this project")
		default:
			return err
		}
	} else {
		log.Info("sentry initialized")
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := s.server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("listen: %s\n", err)
		}
	}()

	<-s.quit
	if err := s.Stop(); err != nil {
		return err
	}

	return nil
}

// Stop stops the web server
func (s *Srv) Stop() error {
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to gracefully shutdown server: %w", err)
	}

	log.Info("Server exiting")

	return nil
}

// IsError checks if err and aborts with json 500 error
func IsError(c *gin.Context, err error) bool {
	if err != nil {
		log.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"status": false, "message": err.Error()})

		return true // signal that there was an error and the caller should return
	}

	return false // no error, can continue
}

// IsError401 checks if err and aborts with json 401 error
func IsError401(c *gin.Context, err error) bool {
	if err != nil {
		log.Warn(err)
		c.AbortWithStatusJSON(http.StatusUnauthorized,
			gin.H{"status": false, "message": err.Error()})

		return true // signal that there was an error and the caller should return
	}

	return false // no error, can continue
}
