package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mitchellh/go-homedir"
)

var (
	ShortGitHubURI = regexp.MustCompile(`^[\w\-.]+/[\w\-.]+$`)
)

func ToSourceURI(uri string) (string, error) {
	switch {
	case ShortGitHubURI.MatchString(uri):
		return "https://github.com/" + uri, nil
	default:
		return uri, nil
	}
}

var (
	home, errInit = homedir.Dir()
	dotvim        = filepath.Join(home, ".vim")
)

func ToDestinationPath(uri, filetype string) (string, error) {
	name := filepath.Base(uri)
	if filetype == "" {
		return filepath.Join(dotvim, "bundle", name), nil
	}
	return filepath.Join(dotvim, "ftbundle", filetype, name), nil
}

type Package struct {
	src string
	dst string
}

func NewPackage(uri, filetype string) (*Package, error) {
	src, err := ToSourceURI(uri)
	if err != nil {
		return nil, err
	}
	dst, err := ToDestinationPath(uri, filetype)
	if err != nil {
		return nil, err
	}
	return &Package{
		src: src,
		dst: dst,
	}, nil
}

func (p *Package) toInstallCommand() *exec.Cmd {
	return exec.Command("git", "clone", p.src, p.dst)
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

	errMessage := bytes.NewBuffer(make([]byte, 0))

	installcmd := p.toInstallCommand()
	installcmd.Stderr = errMessage
	if err := installcmd.Run(); err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(errMessage.String()))
	}
	return nil
}

func (p *Package) Remove() error {
	if !p.installed() {
		return nil
	}
	return os.RemoveAll(p.dst)
}

func (p *Package) Update() error {
	if p.installed() {
		p.Remove()
	}
	return p.Install()
}
