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

	"os/signal"

	"github.com/spf13/cobra"
	"github.com/syhpoon/xenvman/pkg/api"
	"github.com/syhpoon/xenvman/pkg/config"
	"github.com/syhpoon/xenvman/pkg/conteng"
	"github.com/syhpoon/xenvman/pkg/lib"
	"github.com/syhpoon/xenvman/pkg/logger"
)

var apiLog = logger.GetLogger("xenvman.cmd.api")

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "API server",
}

var apiRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run API server",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		_ = config.BindPFlag("api.listen", cmd.Flag("listen"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())

		fmt.Printf("%s\n", config.GetString("api.listen"))
		fmt.Printf("%s\n", config.GetString("container-engine"))
		return

		// Start API server
		apiParams := api.DefaultServerParams(ctx)

		cengCtx, centCancel := context.WithCancel(context.Background())

		if contEng, err := buildContEng(); err != nil {
			apiLog.Errorf("Error building container engine: %s", err)
		} else {
			apiParams.ContEng = contEng
		}

		listener, err := net.Listen("tcp", config.GetString("api.listen"))

		if err != nil {
			apiLog.Errorf("Unable to start listener: %s", err)

			os.Exit(1)
		}

		apiParams.Listener = listener
		apiParams.ExportAddress = config.GetString("api.export-address")
		apiParams.BaseTplDir = config.GetString("tpl.base-dir")
		apiParams.BaseWsDir = config.GetString("tpl.ws-dir")
		apiParams.BaseMountDir = config.GetString("tpl.mount-dir")
		apiParams.CengCtx = cengCtx

		// Ports
		prange, err := parsePorts()

		if err != nil {
			apiLog.Errorf("Error parsing ports: %s", err)

			os.Exit(1)
		} else {
			apiParams.PortRange = prange
		}

		apiServer := api.NewServer(apiParams)

		wg := &sync.WaitGroup{}

		wg.Add(1)

		go apiServer.Run(wg)

		wait(ctx, cancel, centCancel, wg)
	},
}

func buildContEng() (conteng.ContainerEngine, error) {
	ceng := config.GetString("container-engine")

	switch ceng {
	case "docker":
		apiLog.Infof("Using Docker container engine")

		params := conteng.DockerEngineParams{}

		return conteng.NewDockerEngine(params)
	default:
		return nil, fmt.Errorf("Unknown container engine type: %s", ceng)
	}
}

func wait(ctx context.Context, cancel, centCancel func(), wg *sync.WaitGroup) {
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
	centCancel()
}

func parsePorts() (*lib.PortRange, error) {
	ports := config.GetStrings("ports-range")

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
	apiCmd.AddCommand(apiRunCmd)

	apiRunCmd.Flags().StringP("listen", "l", ":9876", "Listen address")
}
