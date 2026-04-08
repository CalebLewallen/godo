package updater

import (
	"context"
	"fmt"
	"os"

	selfupdate "github.com/creativeprojects/go-selfupdate"
)

// RunUpdate checks GitHub releases for a newer version and replaces the binary.
func RunUpdate(currentVersion string) {
	fmt.Println("Checking for updates...")

	ctx := context.Background()
	repo := selfupdate.ParseSlug("CalebLewallen/godo")

	updater, err := selfupdate.NewUpdater(selfupdate.Config{})
	if err != nil {
		fmt.Printf("Error creating updater: %v\n", err)
		return
	}

	latest, found, err := updater.DetectLatest(ctx, repo)
	if err != nil {
		fmt.Printf("Error checking for updates: %v\n", err)
		return
	}

	if !found {
		fmt.Printf("Already up to date (version %s)\n", currentVersion)
		return
	}

	if currentVersion != "dev" && latest.LessOrEqual(currentVersion) {
		fmt.Printf("Already up to date (version %s)\n", currentVersion)
		return
	}

	fmt.Printf("Updating to version %s...\n", latest.Version())
	exe, err := os.Executable()
	if err != nil {
		fmt.Printf("Error getting executable path: %v\n", err)
		return
	}
	if err := updater.UpdateTo(ctx, latest, exe); err != nil {
		fmt.Printf("Update failed: %v\n", err)
		return
	}

	fmt.Printf("Successfully updated to version %s\n", latest.Version())
	fmt.Println("Please restart godo to use the new version.")
}
