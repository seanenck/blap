// Package core handles asset information for file extraction management
package core

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/seanenck/blap/internal/util"
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
	// Resource handles download information and extract for asset managing
	Resource struct {
		URL   string
		File  string
		Tag   string
		Paths struct {
			set     bool
			Archive string
			Unpack  string
		}
		extract Extraction
	}
)

// ID attempt to get a reasonable id for file system parsing
func (asset *Resource) ID() (string, error) {
	h := sha256.New()
	if _, err := fmt.Fprintf(h, "%s%s%s", asset.File, asset.Tag, asset.URL); err != nil {
		return "", err
	}
	hash := fmt.Sprintf("%x", h.Sum(nil))
	return hash[0:7], nil
}

// SetAppData will set the asset's data for the overall application
func (asset *Resource) SetAppData(name, workdir string, settings Extraction) error {
	if name == "" || workdir == "" {
		return errors.New("name and directory are required")
	}
	if asset.Tag == "" || asset.File == "" || asset.URL == "" {
		return errors.New("asset not initialized properly")
	}
	hash, err := asset.ID()
	if err != nil {
		return err
	}
	asset.Paths.set = true
	asset.Paths.Archive = filepath.Join(workdir, util.CleanFileName(fmt.Sprintf("%s.%s", hash, asset.File)))
	asset.Paths.Unpack = filepath.Join(workdir, util.CleanFileName(fmt.Sprintf("%s.%s.%s", hash, name, asset.Tag)))
	asset.extract = settings
	asset.extract.NoDepth = true
	if len(settings.Command) == 0 {
		asset.extract.NoDepth = settings.NoDepth
		for k, v := range knownExtensions {
			if strings.HasSuffix(asset.File, k) {
				asset.extract.Command = []Resolved{}
				for _, value := range v {
					asset.extract.Command = append(asset.extract.Command, Resolved(value))
				}
				break
			}
		}
	}
	return nil
}

// Extract will unpack an asset
func (asset *Resource) Extract(opts util.Runner) error {
	if !asset.Paths.set {
		return errors.New("asset not set for extraction")
	}
	if len(asset.extract.Command) == 0 {
		return errors.New("asset has no extraction command")
	}
	cmd := asset.extract.Command[0]
	var args []string
	hasIn := false
	hasOut := false
	var depth []string
	if !asset.extract.NoDepth {
		var err error
		depth, err = asset.handleDepth(cmd.String(), opts)
		if err != nil {
			return err
		}
	}
	for idx, a := range asset.extract.Command {
		if idx == 0 {
			continue
		}
		use := a.String()
		switch a {
		case inputArg:
			hasIn = true
			use = asset.Paths.Archive
		case outputArg:
			hasOut = true
			use = asset.Paths.Unpack
		case depthArgs:
			args = append(args, depth...)
			continue
		}
		args = append(args, use)
	}
	if !hasIn || !hasOut {
		return fmt.Errorf("missing input/output args for extract command: %s %s", inputArg, outputArg)
	}
	if err := os.Mkdir(asset.Paths.Unpack, 0o755); err != nil {
		return err
	}
	return opts.RunCommand(cmd.String(), args...)
}

func (asset *Resource) handleDepth(cmd string, opts util.Runner) ([]string, error) {
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
	args = append(args, asset.Paths.Archive)
	out, err := opts.Output(c, args...)
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
