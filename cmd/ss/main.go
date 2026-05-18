package main

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/kimmykuang/selfskill/internal/cli"
	webstatic "github.com/kimmykuang/selfskill/web"
)

var version = "dev"

func main() {
	staticFS, err := fs.Sub(webstatic.StaticFS, "static")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load static files:", err)
		os.Exit(1)
	}
	rootCmd := cli.NewRootCmd(version, staticFS)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
