package utils

import (
	"fmt"
	"os"
)

func MustGetEnv(key string) string {
	envVar := os.Getenv(key)
	if envVar == "" {
		panic(fmt.Sprintf("Could not get %s from .env!", key))
	}
	return envVar
}
