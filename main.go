package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/syscod3/quickserve/internal/app"
	"github.com/syscod3/quickserve/internal/netinfo"
	"github.com/syscod3/quickserve/internal/upnp"
)

func main() {
	cfg := app.Config{}
	flag.StringVar(&cfg.Dir, "dir", ".", "directory to serve")
	flag.IntVar(&cfg.Port, "port", 8000, "local TCP port; use 0 to select an available port")
	flag.BoolVar(&cfg.UPnP, "upnp", false, "request a public TCP port mapping using UPnP IGD")
	flag.IntVar(&cfg.UPnPPort, "upnp-port", 0, "external UPnP port; 0 uses the selected local port")
	flag.DurationVar(&cfg.UPnPLease, "upnp-lease", time.Hour, "UPnP lease duration; 0 requests a permanent mapping")
	flag.BoolVar(&cfg.Version, "version", false, "print version information and exit")
	flag.Parse()

	if cfg.Version {
		app.PrintVersion(os.Stdout, app.CurrentBuildInfo())
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	runner := app.NewRunner(cfg, netinfo.DefaultProvider(), upnp.NewDefaultManager())
	started, errc := runner.Start(ctx, os.Stdout)
	select {
	case err := <-errc:
		if err != nil {
			fmt.Fprintf(os.Stderr, "quickserve: %v\n", err)
			os.Exit(1)
		}
	case <-started.Ready:
	}

	err := <-errc
	if err != nil {
		fmt.Fprintf(os.Stderr, "quickserve: %v\n", err)
		os.Exit(1)
	}
}
