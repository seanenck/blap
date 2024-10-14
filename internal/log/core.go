package log

import "fmt"

func Write(msg string) {
	fmt.Printf("  %s", msg)
}
