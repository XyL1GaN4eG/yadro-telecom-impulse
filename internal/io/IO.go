package io

import (
	"fmt"
	"io"
	"os"
)

var writer io.Writer = os.Stdout

func SetOutput(w io.Writer) {
	writer = w
}

func Output(msg string) error {
	_, err := fmt.Fprintln(writer, msg)
	return err
}
