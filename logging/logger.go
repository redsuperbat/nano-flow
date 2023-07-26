package logging

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

const (
	LOG_FORMAT_PLAIN = "plain"
	LOG_FORMAT_JSON  = "json"
)

func getFormat() string {
	format, ok := os.LookupEnv("NANO_LOG_FORMAT")
	if !ok {
		return LOG_FORMAT_JSON
	}
	return format

}

func New() *zap.SugaredLogger {
	format := getFormat()

	if format == LOG_FORMAT_JSON {
		logger, err := zap.NewProduction()
		if err != nil {
			panic(err)
		}
		return logger.Sugar()
	}
	if format == LOG_FORMAT_PLAIN {
		logger, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		return logger.Sugar()
	}

	errMsg := fmt.Sprintf("invalid log format 'NANO_LOG_FORMAT' expected '%s'|'%s' got %s", LOG_FORMAT_PLAIN, LOG_FORMAT_JSON, format)
	panic(errMsg)
}
