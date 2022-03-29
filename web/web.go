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
	"syscall"
	"time"

	_ "github.com/Masterminds/sprig"
	_ "github.com/dghubble/sling"
	"github.com/dustin/go-humanize"
	_ "github.com/foolin/goview"
	"github.com/gin-gonic/gin"
	_ "github.com/gobuffalo/validate"
	_ "github.com/gorilla/schema"
	_ "github.com/gotailwindcss/tailwind"
	_ "github.com/joncalhoun/form"
	log "github.com/sirupsen/logrus"
	"github.com/threecommaio/opc/version"
	ginlogrus "github.com/toorop/gin-logrus"
	_ "google.golang.org/grpc"
)

// Srv is the web server
type Srv struct {
	cfg    SrvConfig
	quit   chan os.Signal
	server *http.Server
}

// Option is used for configuring features of the webserver
type Option func(*Srv)

// SrvConfig is the configuration for the web server
type SrvConfig struct {
	ListenAddress string
	ReadTimeout   string
	WriteTimeout  string
}

// New creates the webserver
func New(cfg SrvConfig, opts ...Option) (Srv, *gin.Engine, error) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Setup the gin router
	router := gin.New()
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

// Start starts the web server
func (s *Srv) Start() error {
	log.Infof("build release: %s", version.Release())

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		log.Infof("starting web server on %s", s.cfg.ListenAddress)
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
