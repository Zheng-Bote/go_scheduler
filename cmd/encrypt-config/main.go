package main

import (
	"encoding/json"
	"fmt"

	"log"
	"os"
	"syscall"

	"go-scheduler/internal/config"
	"go-scheduler/internal/crypto"

	"golang.org/x/term"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: encrypt-config <input_json> <output_encrypted>")
		os.Exit(1)
	}

	inputPath := os.Args[1]
	outputPath := os.Args[2]

	// Read input JSON
	plaintext, err := os.ReadFile(inputPath)
	if err != nil {
		log.Fatalf("Failed to read input file: %v", err)
	}

	// Validate JSON
	var dbConfig config.DBConfig
	if err := json.Unmarshal(plaintext, &dbConfig); err != nil {
		log.Fatalf("Invalid JSON format: %v", err)
	}

	// Ask for password
	fmt.Print("Enter password to encrypt: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("\nFailed to read password: %v", err)
	}
	fmt.Println()

	// Encrypt
	ciphertext, err := crypto.Encrypt(plaintext, bytePassword)
	if err != nil {
		log.Fatalf("Encryption failed: %v", err)
	}

	// Save
	if err := os.WriteFile(outputPath, ciphertext, 0600); err != nil {
		log.Fatalf("Failed to write encrypted file: %v", err)
	}

	fmt.Printf("Successfully encrypted %s to %s\n", inputPath, outputPath)
}
