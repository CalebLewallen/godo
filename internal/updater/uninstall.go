package updater

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/CalebLewallen/godo/internal/db"
)

// RunUninstall prompts the user and removes the binary and data directory.
func RunUninstall() {
	fmt.Print("Remove godo and all data? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" {
		fmt.Println("Uninstall cancelled.")
		return
	}

	dataDir := db.DataDirPath()
	fmt.Printf("Removing data directory: %s\n", dataDir)
	if err := os.RemoveAll(dataDir); err != nil {
		fmt.Printf("Warning: could not remove data directory: %v\n", err)
	}

	exe, err := os.Executable()
	if err != nil {
		fmt.Printf("Warning: could not determine executable path: %v\n", err)
		return
	}

	fmt.Printf("Removing binary: %s\n", exe)
	if err := os.Remove(exe); err != nil {
		fmt.Printf("Warning: could not remove binary (on Windows, delete manually): %v\n", err)
		return
	}

	fmt.Println("godo has been uninstalled.")
}
