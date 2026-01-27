package commands

import (
	"flag"
	"fmt"
	"os"
	"syscall"

	"github.com/rrbarrero/justbackup/internal/cli/client"
	"github.com/rrbarrero/justbackup/internal/cli/config"
	"golang.org/x/term"
)

func BootstrapCommand() {
	// 1. Parse Arguments
	bootstrapCmd := flag.NewFlagSet("bootstrap", flag.ExitOnError)
	host := bootstrapCmd.String("host", "", "Hostname or IP of the target host (required)")
	user := bootstrapCmd.String("user", "", "SSH User (default: current user)")
	name := bootstrapCmd.String("name", "", "Name alias for the host (required)")
	port := bootstrapCmd.Int("port", 22, "SSH Port")

	if len(os.Args) < 2 {
		fmt.Println("Usage: justbackup bootstrap --host <host> --name <name> [--user <user>] [--port <port>]")
		os.Exit(1)
	}
	bootstrapCmd.Parse(os.Args[2:])

	if *host == "" || *name == "" {
		fmt.Println("Error: --host and --name are required.")
		bootstrapCmd.Usage()
		os.Exit(1)
	}

	// 2. Resolve User
	if *user == "" {
		userRetriever := NewOSUserRetriever()
		username, err := userRetriever.GetCurrentUsername()
		if err != nil {
			fmt.Printf("Error getting current user: %v\n", err)
			os.Exit(1)
		}
		*user = username
		fmt.Printf("No user provided, using current user: %s\n", *user)
	}

	// 3. Prompt for SSH Password
	fmt.Printf("Enter SSH password for %s@%s: ", *user, *host)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("\nError reading password: %v\n", err)
		return
	}
	fmt.Println() // Newline after password input
	password := string(bytePassword)

	// 4. Initialize Services
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\nRun 'justbackup config' to configure the CLI.\n", err)
		return
	}

	apiClient := client.NewClient(cfg)
	apiService := NewAPIService(apiClient)
	sshService := NewSSHService()
	bootstrapService := NewBootstrapService(sshService, apiService)

	// 5. Execute Bootstrap
	fmt.Println("Starting bootstrap process...")
	params := BootstrapParams{
		Name:     *name,
		Host:     *host,
		User:     *user,
		Port:     *port,
		Password: password,
	}

	if err := bootstrapService.Execute(params); err != nil {
		fmt.Printf("Bootstrap failed: %v\n", err)
		return
	}

	fmt.Printf("Host '%s' bootstrapped and registered successfully!\n", *name)
}
