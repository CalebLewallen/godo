package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/CalebLewallen/godo/internal/db"
	"github.com/CalebLewallen/godo/internal/ui"
	"github.com/CalebLewallen/godo/internal/updater"
	tea "github.com/charmbracelet/bubbletea"
)

// version is injected at build time via:
//
//	go build -ldflags "-X main.version=v0.1.0"
var version = "dev"

func main() {
	updateFlag := flag.Bool("update", false, "update godo to the latest version")
	uninstallFlag := flag.Bool("uninstall", false, "remove godo and all associated data")
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("godo %s\n", version)
		os.Exit(0)
	}

	if *updateFlag {
		updater.RunUpdate(version)
		os.Exit(0)
	}

	if *uninstallFlag {
		updater.RunUninstall()
		os.Exit(0)
	}

	database := db.Open()
	defer database.Close()
	database.RunMigrations()

	p := tea.NewProgram(
		ui.NewAppModel(database),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "godo: %v\n", err)
		os.Exit(1)
	}
}
