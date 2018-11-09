/*
 MIT License

 Copyright (c) 2018 Max Kuznetsov <syhpoon@syhpoon.ca>

 Permission is hereby granted, free of charge, to any person obtaining a copy
 of this software and associated documentation files (the "Software"), to deal
 in the Software without restriction, including without limitation the rights
 to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 copies of the Software, and to permit persons to whom the Software is
 furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included in all
 copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 SOFTWARE.
*/

package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"syscall"

	"github.com/pkg/errors"

	"os/signal"

	"github.com/spf13/cobra"
	"github.com/syhpoon/xenvman/pkg/config"
	"github.com/syhpoon/xenvman/pkg/conteng"
	"github.com/syhpoon/xenvman/pkg/lib"
	"github.com/syhpoon/xenvman/pkg/logger"
	"github.com/syhpoon/xenvman/pkg/server"
)

var runLog = logger.GetLogger("xenvman.cmd.run")

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run xenvman server",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		_ = config.BindPFlag("listen", cmd.Flag("listen"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())

		// Start API server
		params := server.DefaultParams(ctx)

		cengCtx, cengCancel := context.WithCancel(context.Background())

		if contEng, err := buildContEng(); err != nil {
			runLog.Errorf("Error building container engine: %s", err)
		} else {
			params.ContEng = contEng
		}

		listener, err := net.Listen("tcp", config.GetString("listen"))

		if err != nil {
			runLog.Errorf("Unable to start listener: %s", err)

			os.Exit(1)
		}

		params.Listener = listener
		params.ExportAddress = config.GetString("export_address")
		params.BaseTplDir = config.GetString("tpl.base_dir")
		params.BaseWsDir = config.GetString("tpl.ws_dir")
		params.BaseMountDir = config.GetString("tpl.mount_dir")
		params.TLSCertFile = config.GetString("tls.cert")
		params.TLSKeyFile = config.GetString("tls.key")
		params.CengCtx = cengCtx

		// Ports
		prange, err := parsePorts()

		if err != nil {
			runLog.Errorf("Error parsing ports: %s", err)

			os.Exit(1)
		} else {
			params.PortRange = prange
		}

		// Auth backend
		authb, err := parseAuthBackend()

		if err != nil {
			runLog.Errorf("Error parsing auth backend: %+v", err)

			os.Exit(1)
		} else {
			runLog.Infof("Using auth backend: %s", authb.String())

			params.AuthBackend = authb
		}

		srv := server.New(params)

		wg := &sync.WaitGroup{}

		wg.Add(1)

		errch := make(chan error)

		go srv.Run(wg, errch)

		wait(ctx, cancel, cengCancel, wg, errch)
	},
}

func buildContEng() (conteng.ContainerEngine, error) {
	ceng := config.GetString("container_engine")

	switch ceng {
	case "docker":
		runLog.Infof("Using Docker container engine")

		params := conteng.DockerEngineParams{}

		return conteng.NewDockerEngine(params)
	default:
		return nil, fmt.Errorf("Unknown container engine type: %s", ceng)
	}
}

func wait(ctx context.Context, cancel, cengCancel func(),
	wg *sync.WaitGroup, errch <-chan error) {
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)
	signal.Notify(c,
		syscall.SIGTERM,
		syscall.SIGABRT,
		syscall.SIGPIPE,
		syscall.SIGBUS,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
		syscall.SIGQUIT)

LOOP:
	for {
		select {
		case <-ctx.Done():
			break
		case err := <-errch:
			runLog.Errorf("Error running server: %+v", err)

			break LOOP
		case sig := <-c:
			switch sig {
			case os.Interrupt, syscall.SIGTERM, syscall.SIGABRT,
				syscall.SIGPIPE, syscall.SIGBUS, syscall.SIGUSR1, syscall.SIGUSR2,
				syscall.SIGQUIT:
				cancel()

				break LOOP
			}
		}
	}

	wg.Wait()
	cengCancel()
}

func parseAuthBackend() (server.AuthBackend, error) {
	b := config.GetString("api_auth")

	switch b {
	case "":
		return nil, nil
	case "basic":
		return parseAuthBackendBasic()
	default:
		return nil, errors.Errorf("Unknown auth backend type: %s", b)
	}
}

func parseAuthBackendBasic() (server.AuthBackend, error) {
	creds := config.GetStringMapString("auth_basic")

	if len(creds) == 0 {
		return nil, nil
	}

	return &server.AuthBackendBasic{Credentials: creds}, nil
}

func parsePorts() (*lib.PortRange, error) {
	ports := config.GetStrings("ports_range")

	if len(ports) != 2 {
		return nil, fmt.Errorf("Expected two ports for a range, got %d", len(ports))
	}

	pMin, err := strconv.ParseUint(ports[0], 0, 16)

	if err != nil {
		return nil, fmt.Errorf("Error parsing port %s: %s", ports[0], err)
	}

	pMax, err := strconv.ParseUint(ports[1], 0, 16)

	if err != nil {
		return nil, fmt.Errorf("Error parsing port %s: %s", ports[1], err)
	}

	if pMax < pMin {
		return nil, fmt.Errorf("Port %d should be greater than %d", pMax, pMin)
	}

	return lib.NewPortRange(uint16(pMin), uint16(pMax)), nil
}

func init() {
	runCmd.Flags().StringP("listen", "l", ":9876", "Listen address")
}
