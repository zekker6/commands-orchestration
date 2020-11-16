package updater

import (
	"bufio"
	"fmt"
	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"log"
	"os"
)

func DoSelfUpdate(currentVersion string) {
	latest, found, err := selfupdate.DetectLatest("zekker6/commands-orchestration")
	if err != nil {
		log.Println("Error occurred while detecting version:", err)
		return
	}

	if currentVersion != "" {
		v := semver.MustParse(currentVersion)
		if !found || latest.Version.LTE(v) {
			log.Println("Current version is the latest")
			return
		}
	}

	fmt.Print("Do you want to update to ", latest.Version, "? (y/n): ")
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil || (input != "y\n" && input != "n\n") {
		log.Println("Invalid input")
		return
	}
	if input == "n\n" {
		return
	}

	exe, err := os.Executable()
	if err != nil {
		log.Println("Could not locate executable path")
		return
	}
	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		log.Println("Error occurred while updating binary:", err)
		return
	}
	log.Println("Successfully updated to version", latest.Version, latest.ReleaseNotes)
}
