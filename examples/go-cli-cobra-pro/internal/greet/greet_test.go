package greet_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"{{.module_name}}/internal/greet"
)

func TestGreeting(t *testing.T) {
	got := greet.Greeting("Alice")
	require.Equal(t, "Hello, Alice!", got)
}
