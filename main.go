package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

const REVEAL_JS_URL = "git@github.com:hakimel/reveal.js.git"

func main() {
	if err := Main(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func Main(args []string) error {
	dir, err := GitCloneReveal()
	if err != nil {
		return err
	}
	files, err := GrepCopyTargets(dir)
	if err != nil {
		return err
	}
	// TODO: index.html を何とかする
	// TODO: md mode
	fmt.Println(files)

	return nil
}

// Returns cloned dir
func GitCloneReveal() (string, error) {
	dir, err := ioutil.TempDir("", "reveal-init-")
	if err != nil {
		return "", errors.Wrap(err, "Fail to create tmp dir")
	}
	out, err := exec.Command("git", "clone", "--depth", "1", REVEAL_JS_URL, dir).CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "git clone is failed. out: %s", string(out))
	}

	return dir, nil
}

func GrepCopyTargets(dir string) ([]string, error) {
	cmd := exec.Command("git", "ls-files")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "git ls-files is failed")
	}

	files := strings.Split(string(out), "\n")

	res := make([]string, 0, len(files))
	dontCopiedFiles := []string{
		".gitignore",
		".travis.yml",
		"CONTRIBUTING.md",
		"Gruntfile.js",
		"README.md",
		"bower.json",
		"demo.html",
		"package.json",
	}
	for _, file := range files {
		if strings.HasPrefix(file, "test/") ||
			strings.HasPrefix(file, "css/theme/source/") ||
			strings.HasPrefix(file, "css/theme/template/") ||
			strings.HasPrefix(file, "css/theme/template/") ||
			ContainStringSlice(file, dontCopiedFiles) {

			continue
		}
		res = append(res, file)
	}
	return res, nil
}

func ContainStringSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
