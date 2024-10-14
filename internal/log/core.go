// Package log handles basic logging
package log

import "fmt"

// Write will print an indented message
func Write(msg string) {
	fmt.Printf("  %s", msg)
}
