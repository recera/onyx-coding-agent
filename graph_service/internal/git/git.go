package git

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-git/go-git/v5"
)

// CloneRepository clones a git repository to a temporary directory and returns the path.
func CloneRepository(url string) (string, error) {
	dir, err := ioutil.TempDir("", "clone-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	_, err = git.PlainClone(dir, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout, // Optional: print progress to stdout
	})
	if err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	fmt.Printf("Repository cloned to: %s\n", dir)
	return dir, nil
}
