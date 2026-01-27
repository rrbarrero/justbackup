package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/crypto"
)

func DecryptCommand() {
	decryptCmd := flag.NewFlagSet("decrypt", flag.ExitOnError)
	filePtr := decryptCmd.String("file", "", "Path to the encrypted .tar.gz.enc file (required)")
	outPtr := decryptCmd.String("out", "", "Path to the output .tar.gz file (required)")
	idPtr := decryptCmd.String("id", "", "Backup ID (required)")
	keyPtr := decryptCmd.String("key", "", "Master encryption key (required)")

	if len(os.Args) < 3 {
		fmt.Println("Usage: justbackup decrypt --file <path> --out <path> --id <backup-id> --key <master-key>")
		decryptCmd.PrintDefaults()
		os.Exit(1)
	}

	decryptCmd.Parse(os.Args[2:])

	if *filePtr == "" || *outPtr == "" || *idPtr == "" || *keyPtr == "" {
		fmt.Println("Missing required flags")
		decryptCmd.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("Deriving key for backup %s...\n", *idPtr)
	key, err := crypto.DeriveKey(*keyPtr, *idPtr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to derive key: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Decrypting %s to %s...\n", *filePtr, *outPtr)
	if err := crypto.DecryptFile(*filePtr, *outPtr, key); err != nil {
		fmt.Fprintf(os.Stderr, "Decryption failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Decryption successful!")
	fmt.Printf("You can now extract the file using: tar -xzvf %s\n", *outPtr)
}
