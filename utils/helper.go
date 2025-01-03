package utils

import (
	"fmt"
	"os"
)

func P(a string) {
	fmt.Print(a)
	os.Exit(1)
}
