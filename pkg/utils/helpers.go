package utils

import (
	"log"
	"os"
)

// LogError logs an error message to the standard error output.
func LogError(err error) {
	if err != nil {
		log.Printf("Error: %v", err)
	}
}

// CheckError checks an error and logs it if not nil, then exits the program.
func CheckError(err error) {
	if err != nil {
		LogError(err)
		os.Exit(1)
	}
}

// PrintMessage prints a message to the standard output.
func PrintMessage(message string) {
	log.Println(message)
}