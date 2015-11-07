package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mitchellh/go-homedir"
)

var (
	ShortGitHubURI = regexp.MustCompile(`^[\w\-.]+/[\w\-.]+$`)
	GitHubURI      = regexp.MustCompile(`^github.com/[\w\-.]+/[\w\-.]+$`)
	BitbucketURI   = regexp.MustCompile(`^bitbucket.org/[\w\-.]+/[\w\-.]+$`)
)

func ToSourceURI(uri string) string {
	switch {
	case ShortGitHubURI.MatchString(uri):
		return "https://github.com/" + uri
	case GitHubURI.MatchString(uri):
		return "https://" + uri
	case BitbucketURI.MatchString(uri):
		return "https://" + uri
	default:
		return uri
	}
}

var (
	home, errInit = homedir.Dir()
	dotvim        = filepath.Join(home, ".vim")
)

func ToDestinationPath(uri, filetype string) string {
	name := filepath.Base(uri)
	if filetype == "" {
		return filepath.Join(dotvim, "bundle", name)
	}
	return filepath.Join(dotvim, "ftbundle", filetype, name)
}

func ListPackages(filetype string) {
	var path string
	if filetype == "" {
		path = filepath.Join(dotvim, "bundle")
	} else {
		path = filepath.Join(dotvim, "ftbundle", filetype)
	}

	// Ignore err for filetype doesn't exist.
	files, _ := ioutil.ReadDir(path)
	for _, file := range files {
		fmt.Println(file.Name())
	}
}

type Package struct {
	src string
	dst string
}

func NewPackage(uri, filetype string) *Package {
	return &Package{
		src: ToSourceURI(uri),
		dst: ToDestinationPath(uri, filetype),
	}
}

func (p *Package) installed() bool {
	_, err := os.Stat(p.dst)
	return err == nil
}

func (p *Package) Install() error {
	if p.installed() {
		return nil
	}

	if _, err := exec.LookPath("git"); err != nil {
		return err
	}
	errBuf := bytes.NewBuffer(make([]byte, 0))
	c := exec.Command("git", "clone", p.src, p.dst)
	c.Stderr = errBuf
	if err := c.Run(); err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(errBuf.String()))
	}
	return nil
}

func (p *Package) Remove() error {
	if p.installed() {
		return os.RemoveAll(p.dst)
	}
	return nil
}

func (p *Package) Update() error {
	if err := p.Remove(); err != nil {
		return err
	}
	return p.Install()
}
