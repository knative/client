package logger

import (
	"fmt"
	"io"
	"os"
)

type Config struct {
	QuietMode bool
}

func NewLogger(config Config) io.Writer {
	if config.QuietMode {
		return io.Discard
	}

	return io.Writer(os.Stdout)
}

func Log(writer io.Writer, message string) {
	fmt.Fprintf(writer, message)
}
