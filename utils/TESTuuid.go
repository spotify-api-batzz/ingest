//go:build test
// +build test

package utils

import "fmt"

var (
	val = 0
)

func GenerateUUID() string {
	val++
	return fmt.Sprintf("%d", val)
}
