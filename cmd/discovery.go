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
	"strings"
	"sync"
	"syscall"
	"time"

	"encoding/json"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/syhpoon/xenvman/pkg/discovery"
	"github.com/syhpoon/xenvman/pkg/logger"
)

var discLog = logger.GetLogger("xenvman.cmd.discovery")

var (
	flagDnsAddr     string
	flagHttpAddr    string
	flagRecursors   []string
	flagOwnDomain   string
	flagMappingFile string
)

var discCmd = &cobra.Command{
	Use:   "discovery",
	Short: "Run discovery agent",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())

		mapping, err := loadMapping(flagMappingFile)

		if err != nil {
			discLog.Errorf("Error loading domain mapping: %s", err)

			os.Exit(1)
		}

		httpListener, err := net.Listen("tcp", flagHttpAddr)

		if err != nil {
			discLog.Errorf("Error listening on %s: %s", flagHttpAddr, err)

			os.Exit(1)
		}

		// Setup DNS server
		dnsParams := discovery.DnsServerParams{
			Addr:      flagDnsAddr,
			DomainMap: mapping,
			Recursors: processRecursors(flagRecursors),
			OwnDomain: flagOwnDomain,
			Ctx:       ctx,
		}

		dnsServer := discovery.NewDnsServer(dnsParams)

		// Setup HTTP server
		httpParams := discovery.HttpServerParams{
			Listener:     httpListener,
			WriteTimeout: 1 * time.Minute,
			ReadTimeout:  1 * time.Minute,
			DnsServer:    dnsServer,
			Ctx:          ctx,
		}

		httpServer := discovery.NewHttpServer(httpParams)

		wg := &sync.WaitGroup{}
		wg.Add(2)

		errch := make(chan error)

		go dnsServer.Run(wg, errch)
		go httpServer.Run(wg, errch)

		waitDiscovery(ctx, cancel, wg, errch)
	},
}

func waitDiscovery(ctx context.Context, cancel func(),
	wg *sync.WaitGroup, errch <-chan error) {

	c := make(chan os.Signal, 1)

	signal.Notify(c,
		os.Interrupt,
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
			break LOOP
		case err := <-errch:
			discLog.Errorf("Error running discovery agent: %+v", err)

			cancel()

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
}

func loadMapping(file string) (map[string]string, error) {
	m := map[string]string{}
	f, err := os.Open(file)

	if err != nil {
		return nil, err
	}

	d := json.NewDecoder(f)

	err = d.Decode(&m)

	if err == nil {
		b, _ := json.MarshalIndent(m, "", "  ")
		discLog.Infof("Loaded domains: %s", string(b))
	}

	return m, err
}

func processRecursors(recs []string) []string {
	newList := make([]string, len(recs))

	for i, rec := range recs {
		if len(strings.Split(rec, ":")) == 1 {
			newList[i] = rec + ":53"
		} else {
			newList[i] = rec
		}
	}

	return newList
}

func init() {
	discCmd.Flags().StringVarP(&flagDnsAddr,
		"dns-addr", "d", ":53", "DNS listen address")

	discCmd.Flags().StringVarP(&flagHttpAddr,
		"http-addr", "t", ":8080", "HTTP listen address")

	discCmd.Flags().StringVarP(&flagOwnDomain,
		"domain", "n", ".xenv.", "Internal domain")

	discCmd.Flags().StringVarP(&flagMappingFile,
		"map", "m", "", "File containing mapping between hosts and IP addresses encoded as a JSON object")

	discCmd.Flags().StringSliceVarP(&flagRecursors,
		"recursors", "r", []string{"1.1.1.1:53"}, "DNS recursors")
}
