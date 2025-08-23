package greet

import "fmt"

// Greeting returns a friendly greeting for the given name.
func Greeting(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}
