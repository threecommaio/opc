// Package logging provides logging functionality and configuration.
package logging

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
)

var (
	errParseLogLevel = errors.New("failed to parse log level")
)

func init() {
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
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
		ForceColors:   true,
	})
}

// SetLevel sets the log level
func SetLevel(loglevel string) error {
	l, err := log.ParseLevel(loglevel)
	if err != nil {
		return fmt.Errorf("%w: %s - %s", errParseLogLevel, loglevel, err)
	}

	log.SetLevel(l)

	return nil
}
