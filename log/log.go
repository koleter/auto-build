package log

import (
	"os"

	"github.com/subchen/go-log"
	"github.com/subchen/go-log/formatters"
	"github.com/subchen/go-log/writers"
)

func init() {
	log.Default.Out = os.Stdout
	log.Default.Formatter = new(formatters.TextFormatter)
}

func SetLogFileName(name string, level string) {
	log.Default.Out = &writers.FixedSizeFileWriter{
		Name:     name,
		MaxSize:  10 * 1024 * 1024, // 10m
		MaxCount: 10,
	}

	if len(level) == 0 {
		level = "DEBUG"
	}
	log.Default.Level, _ = log.ParseLevel(level)
}
