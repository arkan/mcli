package github

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"syscall"

	"github.com/google/go-github/github"
	"github.com/hashicorp/go-version"
	"golang.org/x/oauth2"
)

type Github struct {
	client *github.Client
	owner  string
	repo   string
}

type Option func(g *Github)

func New(token string, owner string, repo string, opts ...Option) *Github {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	gh := &Github{
		client: github.NewClient(tc),
		owner:  owner,
		repo:   repo,
	}
	for _, opt := range opts {
		opt(gh)
	}
	return gh
}

func (g *Github) IsTokenValid() bool {
	organizations, _, err := g.client.Organizations.List(context.Background(), "", &github.ListOptions{})
	if err != nil {
		return false
	}
	for _, o := range organizations {
		if o.GetLogin() == g.owner {
			return true
		}
	}
	return false
}

func (c *Github) DownloadNewVersionIfNeeded(binaryName, currentVersion string) error {
	latest, _, err := c.client.Repositories.GetLatestRelease(context.Background(), c.owner, c.repo)
	if err != nil {
		return err
	}

	if len(latest.Assets) == 0 {
		return nil
	}

	cVersion, err := version.NewVersion(currentVersion)
	if err != nil {
		return err
	}
	lVersion, err := version.NewVersion(latest.GetTagName())
	if err != nil {
		return err
	}

	if cVersion.GreaterThanOrEqual(lVersion) {
		return nil
	}

	goOS := runtime.GOOS
	goArch := runtime.GOARCH

	for _, a := range latest.Assets {
		if strings.Contains(strings.ToLower(a.GetName()), strings.ToLower(goOS)) {
			if strings.Contains(strings.ToLower(a.GetName()), strings.ToLower(goArch)) {
				fmt.Printf("* Upgrading from %s to %s... \n", currentVersion, lVersion.String())
				fmt.Printf("* Downloading %s...\n", *a.Name)

				r, _, err := c.client.Repositories.DownloadReleaseAsset(context.Background(), c.owner, c.repo, a.GetID())
				if err != nil {
					return err
				}

				dest, err := os.Executable()
				if err != nil {
					return err
				}

				newDest := dest + ".new"
				if err := downloadFile(r, binaryName, newDest); err != nil {
					return err
				}

				oldDest := dest + ".old"
				if err := os.Rename(dest, oldDest); err != nil {
					return err
				}

				if err := os.Rename(newDest, dest); err != nil {
					return err
				}

				// Removing backup.
				_ = os.Remove(oldDest)

				fmt.Printf("* %s successfully updated to %s.\n* Restarting %s ...\n\n", binaryName, lVersion.String(), binaryName)

				// The update completed, we can now restart the application without requiring any user action.
				if err := syscall.Exec(dest, os.Args, os.Environ()); err != nil {
					return err
				}

				os.Exit(0)
			}
		}
	}

	return fmt.Errorf("release %s doesn't contain binary for %s@%s", lVersion.String(), goOS, goArch)
}

func downloadFile(r io.ReadCloser, binaryName string, dest string) error {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(zr)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if header.Typeflag == tar.TypeReg && header.Name == binaryName {
			f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o777)
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.CopyN(f, tarReader, header.Size); err != nil {
				return err
			}

			return nil
		}
	}

	return fmt.Errorf("executable file not found in the release")
}
