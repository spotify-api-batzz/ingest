package utils

import (
	"fmt"
	"os"
	"path"

	"github.com/batzz-00/goutils/logger"
)

func LoadJSON(returnType string, testDir string) func(fileName string) []byte {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return func(fileName string) []byte {
		fixture := fmt.Sprintf("%s.json", fileName)
		filePath := path.Join(wd, "integration", returnType, testDir, fixture)
		logger.Log(fmt.Sprintf("trying to fetch fixture %s\n", filePath), logger.Info)
		bytes, err := os.ReadFile(filePath)
		if err != nil {
			panic(err)
		}

		return bytes
	}
}
