package utils

import (
	"bufio"
	"log"
	"os"
)

func ReadFileByScanner(filePath string, visitor func(scanner *bufio.Scanner)) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	visitor(bufio.NewScanner(file))
}
