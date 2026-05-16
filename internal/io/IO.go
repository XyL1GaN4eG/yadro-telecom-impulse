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

func Print(msg string) error {
	_, err := fmt.Fprintln(writer, msg)
	return err
}

func PrintError(err error) error {
	_, err = fmt.Fprintln(os.Stderr, err)
	return err
}
