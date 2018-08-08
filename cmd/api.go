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
	"net"
	"os"
	"sync"
	"syscall"

	"os/signal"

	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/syhpoon/xenvman/pkg/api"
	"github.com/syhpoon/xenvman/pkg/config"
	"github.com/syhpoon/xenvman/pkg/conteng"
	"github.com/syhpoon/xenvman/pkg/logger"
	"github.com/syhpoon/xenvman/pkg/repo"
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
		config.BindPFlag("api.listen", cmd.Flag("listen"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())

		// Start API server
		apiParams := api.DefaultServerParams(ctx)

		if repos, err := parseRepos(); err != nil {
			apiLog.Errorf("Error parsing repos: %+v", err)

			cancel()

			return
		} else {
			apiParams.Repos = repos
		}

		if contEng, err := buildContEng(ctx); err != nil {
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
		apiServer := api.NewServer(apiParams)

		wg := &sync.WaitGroup{}

		wg.Add(1)

		go apiServer.Run(wg)

		wait(ctx, cancel, wg)
	},
}

func buildContEng(ctx context.Context) (conteng.ContainerEngine, error) {
	ceng := config.GetString("api.container-engine")

	switch ceng {
	case "docker":
		apiLog.Infof("Using Docker container engine")

		params := conteng.DockerEngineParams{
			Ctx: ctx,
		}

		return conteng.NewDockerEngine(params)
	default:
		return nil, fmt.Errorf("Unknown container engine: %s", ceng)
	}
}

func wait(ctx context.Context, cancel func(), wg *sync.WaitGroup) {
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
}

func parseRepos() (map[string]repo.Repo, error) {
	rawRepos := config.Get("repo")
	repos := map[string]repo.Repo{}

	switch rr := rawRepos.(type) {
	case []interface{}:
		for _, rawRepo := range rr {
			m := rawRepo.(map[string]interface{})

			name, ok := m["name"].(string)

			if !ok || name == "" {
				return nil, errors.Errorf("Empty name for %v", m)
			}

			switch m["type"] {
			case "shell":
				if r, err := configureShellRepo(m); err != nil {
					return nil, err
				} else {
					apiLog.Infof("Configured shell repo %s: %s", name, r.String())

					repos[name] = r
				}

			default:
				return nil, errors.Errorf("Invalid repo type: %s", m["type"])
			}
		}
	default:
		return nil, errors.Errorf(
			"`repo` expected to be an array of tables, got: %T", rawRepos)
	}

	return repos, nil
}

func configureShellRepo(raw map[string]interface{}) (repo.Repo, error) {
	params := repo.ShellRepoParams{}

	err := mapstructure.Decode(raw, &params)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to decode shell repo params")
	}

	return repo.NewShellRepo(params)
}

func init() {
	apiCmd.AddCommand(apiRunCmd)

	apiRunCmd.Flags().StringP("listen", "l", ":9876", "Listen address")
}
