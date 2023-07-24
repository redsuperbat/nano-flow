package logging

import (
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
		logger, _ := zap.NewProduction()
		return logger.Sugar()
	}
	if format == LOG_FORMAT_PLAIN {
		logger, _ := zap.NewDevelopment()
		return logger.Sugar()
	}

	panic("invalid log format 'NANO_LOG_FORMAT'")
}
