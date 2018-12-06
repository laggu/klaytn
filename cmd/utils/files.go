package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

func WriteFile(content, filePath, fileName string) {
	err := os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		fmt.Printf("Failed to create folder %v failed: %v\n", filePath, err)
		os.Exit(-1)
	}

	err = ioutil.WriteFile(path.Join(filePath, fileName), []byte(content), 0644)
	if err != nil {
		fmt.Printf("Failed to write %v file: %v\n", fileName, err)
		os.Exit(-1)
	}
}
