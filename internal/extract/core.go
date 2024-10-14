package extract

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/seanenck/bd/internal/log"
)

type Asset struct {
	url   string
	file  string
	tag   string
	local struct {
		archive string
		unpack  string
		extract []string
	}
}

func (asset *Asset) URL() string {
	return asset.url
}

func (asset *Asset) SetAppData(name, workdir string) {
	asset.local.archive = filepath.Join(workdir, asset.file)
	asset.local.unpack = filepath.Join(workdir, fmt.Sprintf("%s.%s", name, asset.tag))
	for k, v := range knownExtensions {
		if strings.HasSuffix(asset.file, k) {
			asset.local.extract = v
			break
		}
	}
}

func NewAsset(url, file, tag string) *Asset {
	a := &Asset{}
	a.url = url
	a.file = file
	a.tag = tag
	return a
}

func (asset *Asset) Unpack() string {
	return asset.local.unpack
}

func (asset *Asset) Archive() string {
	return asset.local.archive
}

var knownExtensions = map[string][]string{
	".tar.gz": {"tar", "xf"},
}

func (asset *Asset) HasExtractor() bool {
	return len(asset.local.extract) > 0
}

func (asset *Asset) Extract() error {
	log.Write(fmt.Sprintf("extracting: %s\n", asset.file))
	cmd := asset.local.extract[0]
	var args []string
	if len(asset.local.extract) > 1 {
		args = asset.local.extract[1:]
	}
	args = append(args, asset.local.archive)
	if err := os.Mkdir(asset.local.unpack, 0o755); err != nil {
		return err
	}
	c := exec.Command(cmd, args...)
	c.Dir = asset.local.unpack
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
