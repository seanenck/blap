// Package extract handles asset information for file extraction management
package extract

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/seanenck/bd/internal/log"
)

const (
	inputArg     = "{INPUT}"
	outputArg    = "{OUTPUT}"
	tarCommand   = "tar"
	unzipCommand = "unzip"
)

var (
	knownExtensions = map[string][]string{
		".tar.gz": {tarCommand, "xf", inputArg, "-C", outputArg},
		".zip":    {unzipCommand, "-d", outputArg, inputArg},
	}
	pathSep = string(os.PathSeparator)
)

// Asset handles download information and extract for asset managing
type Asset struct {
	url   string
	file  string
	tag   string
	local struct {
		archive string
		unpack  string
		extract struct {
			command []string
			depth   bool
		}
	}
}

// URL will get the download URL for the asset
func (asset *Asset) URL() string {
	return asset.url
}

// SetAppData will set the asset's data for the overall application
func (asset *Asset) SetAppData(name, workdir string, findDepth bool, extraction []string) error {
	asset.local.archive = filepath.Join(workdir, asset.file)
	asset.local.unpack = filepath.Join(workdir, fmt.Sprintf("%s.%s", name, asset.tag))
	asset.local.extract.command = extraction
	if len(extraction) == 0 {
		asset.local.extract.depth = findDepth
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
	log.Write(fmt.Sprintf("extracting: %s\n", asset.file))
	cmd := asset.local.extract.command[0]
	var args []string
	if asset.local.extract.depth {
		d, err := asset.handleDepth(cmd)
		if err != nil {
			return err
		}
		args = append(args, d...)
	}
	hasIn := false
	hasOut := false
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
