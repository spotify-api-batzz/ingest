package utils

import (
	"fmt"
	"os"
	"strconv"
)

func MustGetEnv(key string) string {
	envVar := os.Getenv(key)
	if envVar == "" {
		panic(fmt.Sprintf("Could not get %s from .env!", key))
	}
	return envVar
}

func MustGetEnvInt(key string) int {
	envVar := os.Getenv(key)
	if envVar == "" {
		panic(fmt.Sprintf("Could not get %s from .env!", key))
	}

	intVal, err := strconv.Atoi(envVar)
	if err != nil {
		panic(err)
	}

	return intVal
}
