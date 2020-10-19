package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tj/go-update"
	"github.com/tj/go-update/stores/github"
	"os/exec"
	"path/filepath"
	"runtime"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates bit to the latest version",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		version := "v0.7.0"

		// open-source edition
		p := &update.Manager{
			Command: "bit",
			Store: &github.Store{
				Owner:   "chriswalz",
				Repo:    "bit",
				Version: version[1:],
			},
		}

		// fetch latest or specified release
		release, err := getLatestOrSpecified(p, version[1:])
		if err != nil {
			fmt.Println(errors.Wrap(err, "fetching latest release").Error())
			return
		}

		// no updates
		if version == release.Version {
			fmt.Println("No updates available, you're good :)")
			return
		}

		// find the tarball for this system
		a := release.FindTarball(runtime.GOOS, runtime.GOARCH)
		if a == nil {
			fmt.Println(fmt.Errorf("failed to find a binary for %s %s", runtime.GOOS, runtime.GOARCH))
			return
		}

		// download tarball to a tmp dir
		tarball, err := a.Download()
		if err != nil {
			fmt.Println(errors.Wrap(err, "downloading tarball"))
			return
		}

		// determine path
		path, err := exec.LookPath("bit")
		if err != nil {
			fmt.Println(errors.Wrap(err, "looking up executable path"))
			return
		}
		dst := filepath.Dir(path)

		// install it
		if err := p.InstallTo(tarball, dst); err != nil {
			fmt.Println(errors.Wrap(err, "installing"))
			return
		}

		fmt.Printf("Updated bit %s to %s in %s", version, release.Version, dst)

	},
	Args: cobra.NoArgs,
}

func init() {
	ShellCmd.AddCommand(updateCmd)
}

// getLatestOrSpecified returns the latest or specified release.
func getLatestOrSpecified(s update.Store, version string) (*update.Release, error) {
	if version == "" {
		return getLatest(s)
	}

	return s.GetRelease(version)
}

// getLatest returns the latest release, error, or nil when there is none.
func getLatest(s update.Store) (*update.Release, error) {
	releases, err := s.LatestReleases()

	if err != nil {
		return nil, errors.Wrap(err, "fetching releases")
	}

	if len(releases) == 0 {
		return nil, nil
	}

	return releases[0], nil
}