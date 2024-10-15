// Package extract handles asset information for file extraction management
package extract

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/seanenck/bd/internal/context"
)

const (
	inputArg     = "{{ $.Input }}"
	outputArg    = "{{ $.Output }}"
	tarCommand   = "tar"
	unzipCommand = "unzip"
	depthArgs    = "{{ $.Depth }}"
)

var (
	tarExtract      = []string{tarCommand, "xf", inputArg, "-C", outputArg, depthArgs}
	knownExtensions = map[string][]string{
		".tar.gz": tarExtract,
		".tar.xz": tarExtract,
		".zip":    {unzipCommand, depthArgs, "-d", outputArg, inputArg},
	}
	pathSep = string(os.PathSeparator)
)

type (
	// Asset handles download information and extract for asset managing
	Asset struct {
		url     string
		file    string
		tag     string
		context context.Settings
		local   struct {
			archive string
			unpack  string
			extract struct {
				command []string
				depth   bool
			}
		}
	}
	// Settings are extraction settings
	Settings struct {
		NoDepth bool     `yaml:"nodepth"`
		Command []string `yaml:"command"`
	}
)

// URL will get the download URL for the asset
func (asset *Asset) URL() string {
	return asset.url
}

// Tag will get the tag for the asset
func (asset *Asset) Tag() string {
	return asset.tag
}

// SetAppData will set the asset's data for the overall application
func (asset *Asset) SetAppData(name, workdir string, settings Settings, context context.Settings) error {
	h := sha256.New()
	if _, err := fmt.Fprintf(h, "%s%s", asset.file, asset.tag); err != nil {
		return err
	}
	hash := fmt.Sprintf("%x", h.Sum(nil))[0:7]
	asset.context = context
	asset.local.archive = filepath.Join(workdir, fmt.Sprintf("%s.%s", hash, asset.file))
	asset.local.unpack = filepath.Join(workdir, fmt.Sprintf("%s.%s", name, strings.ReplaceAll(asset.tag, "/", "_")))
	asset.local.extract.command = settings.Command
	if len(settings.Command) == 0 {
		asset.local.extract.depth = !settings.NoDepth
		for k, v := range knownExtensions {
			if strings.HasSuffix(asset.file, k) {
				asset.local.extract.command = v
				break
			}
		}
	}
	if len(asset.local.extract.command) == 0 {
		return fmt.Errorf("asset missing extractor: %s", name)
	}
	return nil
}

// NewAsset will initialize a new asset
func NewAsset(url, file, tag string) *Asset {
	a := &Asset{}
	a.url = url
	a.file = file
	a.tag = tag
	return a
}

// Unpack will get the directory to unpack to
func (asset *Asset) Unpack() string {
	return asset.local.unpack
}

// Archive will get the asset archive name
func (asset *Asset) Archive() string {
	return asset.local.archive
}

// Extract will unpack an asset
func (asset *Asset) Extract() error {
	asset.context.LogInfoSub(fmt.Sprintf("extracting: %s\n", asset.file))
	cmd := asset.local.extract.command[0]
	var args []string
	hasIn := false
	hasOut := false
	var depth []string
	if asset.local.extract.depth {
		var err error
		depth, err = asset.handleDepth(cmd)
		if err != nil {
			return err
		}
	}
	for idx, a := range asset.local.extract.command {
		if idx == 0 {
			continue
		}
		use := a
		switch a {
		case inputArg:
			hasIn = true
			use = asset.local.archive
		case outputArg:
			hasOut = true
			use = asset.local.unpack
		case depthArgs:
			args = append(args, depth...)
			continue
		}
		args = append(args, use)
	}
	if !hasIn || !hasOut {
		return fmt.Errorf("missing input/output args for extract command: %v", asset.local.extract)
	}
	if err := os.Mkdir(asset.local.unpack, 0o755); err != nil {
		return err
	}
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func (asset *Asset) handleDepth(cmd string) ([]string, error) {
	var c string
	var args []string
	var result []string
	switch cmd {
	case tarCommand:
		c = "tar"
		args = []string{"-tf"}
		result = []string{"--strip-component", "1"}
	case unzipCommand:
		c = "zipinfo"
		args = []string{"-1"}
		result = []string{"-j"}
	default:
		return nil, fmt.Errorf("unable to determine depth for command: %s", cmd)
	}
	args = append(args, asset.Archive())
	out, err := exec.Command(c, args...).Output()
	if err != nil {
		return nil, err
	}
	m := make(map[string]struct{})
	for _, line := range strings.Split(string(out), "\n") {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		if !strings.Contains(t, pathSep) {
			continue
		}
		parts := strings.Split(t, pathSep)
		m[parts[0]] = struct{}{}
		if len(m) > 1 {
			return nil, nil
		}
	}
	if len(m) == 0 {
		return nil, nil
	}
	return result, nil
}
