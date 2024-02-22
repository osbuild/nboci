package nboci

import (
	"fmt"
	"os"
)

func ExitWithError(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", msg, err.Error())
	os.Exit(1)
}

func ExitWithErrorf(format string, err error, args ...string) {
	msg := fmt.Sprintf(format, args)
	fmt.Fprintf(os.Stderr, "%s: %s\n", msg, err.Error())
	os.Exit(1)
}

func ExitWithErrorMsg(msg string) {
	fmt.Fprintf(os.Stderr, "%s\n", msg)
	os.Exit(1)
}
