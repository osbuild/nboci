package nboci

import (
	"fmt"
	"os"
	"strings"
)

var Verbose bool

func Debug(messages ...string) {
	if !Verbose {
		return
	}

	Print(messages...)
}

func Debugf(format string, args ...any) {
	if !Verbose {
		return
	}

	Printf(format, args...)
}

func DebugErr(err error, messages ...string) {
	if !Verbose {
		return
	}

	fmt.Printf("%s: %s\n", strings.Join(messages, " "), err.Error())
}

func Print(messages ...string) {
	fmt.Print(strings.Join(messages, " "))
	fmt.Printf("\n")
}

func Printf(format string, args ...any) {
	fmt.Printf(format, args)
}

func Error(messages ...string) {
	fmt.Fprintf(os.Stderr, "%s\n", strings.Join(messages, " "))
}

func Errorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args)
	fmt.Fprintf(os.Stderr, "%s\n", msg)
}

func ErrorErr(err error, messages ...string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", strings.Join(messages, " "), err.Error())
}

func Fatal(messages ...string) {
	fmt.Fprintf(os.Stderr, "%s\n", strings.Join(messages, " "))
	os.Exit(1)
}

func Fatalf(format string, args ...any) {
	msg := fmt.Sprintf(format, args)
	fmt.Fprintf(os.Stderr, "%s\n", msg)
	os.Exit(1)
}

func FatalErr(err error, messages ...string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", strings.Join(messages, " "), err.Error())
	os.Exit(1)
}

func FatalfErr(err error, format string, args ...any) {
	msg := fmt.Sprintf(format, args)
	fmt.Fprintf(os.Stderr, "%s: %s\n", msg, err.Error())
	os.Exit(1)
}

// OutputWriter writes Go standard library logger to stdout/stderr.
type OutputWriter struct{}

func (slw OutputWriter) Write(p []byte) (n int, err error) {
	Print(strings.TrimSpace(string(p)))

	return len(p), nil
}
