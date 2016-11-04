package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/ogier/pflag"
	"github.com/pkg/errors"
)

const REVEAL_JS_URL = "git@github.com:hakimel/reveal.js.git"

func main() {
	if err := Main(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

func Main(args []string) error {
	confArgs, err := LoadConfigFile()
	if err != nil {
		return err
	}

	c, err := ParseArgs(append(confArgs, args...))
	if err != nil {
		return err
	}

	if !Exists(c.DstDir) {
		err := os.Mkdir(c.DstDir, 0777)
		if err != nil {
			return errors.Wrap(err, "Creating target directory is failed")
		}
	}

	if c.SrcDir == "" {
		dir, err := GitCloneReveal()
		if err != nil {
			return err
		}
		c.SrcDir = dir
		defer os.RemoveAll(dir)
	}

	files, err := GrepCopyTargets(c.SrcDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		err := CopyFile(
			path.Join(c.SrcDir, file),
			path.Join(c.DstDir, file),
		)
		if err != nil {
			return errors.Wrap(err, "Copying file is failed")
		}
	}

	return nil
}

type Config struct {
	SrcDir string
	DstDir string
}

func ParseArgs(args []string) (*Config, error) {
	c := new(Config)
	fs := pflag.NewFlagSet("reveal-init", pflag.ExitOnError)
	fs.StringVarP(&c.SrcDir, "dir", "d", "", "exist reveal.js directory")

	err := fs.Parse(args)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing is failed")
	}

	if len(fs.Args()) <= 1 {
		return nil, errors.Errorf("Please specify target directory as an argument")
	} else {
		c.DstDir = fs.Arg(1)
	}
	return c, nil
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
		"",
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

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func LoadConfigFile() ([]string, error) {
	path, err := homedir.Expand("~/.config/reveal-init")
	if err != nil {
		return nil, errors.Wrap(err, "Expand path is failed")
	}
	if !Exists(path) {
		return []string{}, nil
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	s := strings.Trim(string(b), "\n")
	return strings.Split(s, " "), nil
}
