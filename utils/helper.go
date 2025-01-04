package utils

import (
	"fmt"
	"os"
)

func Dump(a any) {
	fmt.Print(a)
	os.Exit(1)
}
