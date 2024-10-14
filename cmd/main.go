package main

import (
	"fmt"
	"os"

	"github.com/seanenck/bd/internal/core"
	"github.com/seanenck/bd/internal/fetch"
)

func main() {
	cfg := core.Configuration{}
	cfg.Directory = "~/.local/fs"
	app := core.Application{}
	app.Mode = "github"
	app.Name = "rg"
	app.Binaries.Destination = "~/.local/bin"
	app.Binaries.Files = []string{"rg/rg"}
	app.Remote.Upstream = "BurntSushi/ripgrep"
	app.Remote.Asset = "x86_64-unknown-linux-(.+?).tar.gz$"
	//cfg.Applications = append(cfg.Applications, app)
	appt := core.Application{}
	appt.Mode = "tagged"
	appt.Name = "go"
	appt.Remote.Upstream = "https://github.com/golang/go"
	// "no v"
	appt.Remote.Download = "https://go.dev/dl/{{ $.Tag }}.linux-amd64.tar.gz"
	appt.Remote.Filters = []string{"refs/tags/weekly", "refs/tags/release", "[0-9]rc[0-9]"}
	appt.Binaries.Destination = "~/.local/bin/developer"
	appt.Binaries.Files = append(appt.Binaries.Files, "go/bin/go")
	cfg.Applications = append(cfg.Applications, appt)
	if err := cfg.Process(fetch.ResourceFetcher{}); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
	}
}
