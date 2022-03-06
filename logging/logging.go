// Package logging provides logging functionality and configuration.
package logging

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"github.com/threecommaio/opc/core"
)

var (
	ErrParseLogLevel = errors.New("failed to parse log level")
)

// SetLevel sets the log level
func SetLevel(loglevel string) error {
	l, err := log.ParseLevel(loglevel)
	if err != nil {
		return fmt.Errorf("%w: %s - %s", ErrParseLogLevel, loglevel, err)
	}

	log.SetLevel(l)

	return nil
}

type ExtraFieldHook struct {
	service string
	env     string
	pid     int
}

func NewExtraFieldHook(service string, env string) *ExtraFieldHook {
	return &ExtraFieldHook{
		service: service,
		env:     env,
	}
}

func (h *ExtraFieldHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *ExtraFieldHook) Fire(entry *logrus.Entry) error {
	entry.Data["service"] = h.service
	entry.Data["env"] = h.env
	return nil
}

// Init set service name and environment
func Init(service, env string) error {
	log.SetOutput(ioutil.Discard) // Send all logs to nowhere by default

	log.AddHook(&writer.Hook{ // Send logs with level higher than or equal to warning to stderr
		Writer: os.Stderr,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
			log.WarnLevel,
		},
	})
	log.AddHook(&writer.Hook{ // Send trace, debug, and info logs to stdout
		Writer: os.Stdout,
		LogLevels: []log.Level{
			log.InfoLevel,
			log.DebugLevel,
			log.TraceLevel,
		},
	})

	switch env {
	case core.Production:
		gin.SetMode(gin.ReleaseMode)
		log.SetFormatter(&log.JSONFormatter{
			FieldMap: log.FieldMap{
				log.FieldKeyMsg:   "message",
				log.FieldKeyLevel: "severity",
			},
		})
	default:
		log.SetFormatter(&log.TextFormatter{
			DisableColors: false,
			FullTimestamp: true,
			ForceColors:   true,
			PadLevelText:  true,
		})
	}

	log.AddHook(NewExtraFieldHook(service, env))

	log.Print("logs for [debug,info] -> [stdout], [else] -> [stderr]")
	log.Printf("production mode: %t", core.Environment() == core.Production)

	return nil
}
