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

package discovery

import (
	"context"
	"net"
	"sync"
	"time"

	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/syhpoon/xenvman/pkg/logger"
	httpLib "github.com/syhpoon/xenvman/pkg/server"
)

var httpLog = logger.GetLogger("xenvman.pkg.discovery.http_server")

type HttpServerParams struct {
	Listener     net.Listener
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
	DnsServer    *DnsServer
	Ctx          context.Context
}

type HttpServer struct {
	router *mux.Router
	server http.Server
	params HttpServerParams
}

func NewHttpServer(params HttpServerParams) *HttpServer {
	router := mux.NewRouter()

	return &HttpServer{
		router: router,
		server: http.Server{
			Handler:      router,
			WriteTimeout: params.WriteTimeout,
			ReadTimeout:  params.ReadTimeout,
		},
		params: params,
	}
}

func (srv *HttpServer) Run(wg *sync.WaitGroup, errch chan<- error) {
	// API endpoints
	srv.setupHandlers()

	ctx, cancel := context.WithCancel(srv.params.Ctx)
	defer cancel()

	// Shutdown http server upon global signal
	go func() {
		<-ctx.Done()
		_ = srv.server.Shutdown(ctx)

		wg.Done()
	}()

	dnsLog.Infof("Starting HTTP server at %s", srv.params.Listener.Addr())

	err := srv.server.Serve(srv.params.Listener)

	if err != nil {
		errch <- errors.WithStack(err)
	}
}

func (srv *HttpServer) setupHandlers() {
	// PATCH /api/v1/domains - Add new domains
	srv.router.HandleFunc("/api/v1/domains", srv.updateDomainsHandler).
		Methods(http.MethodPatch)

	// DELETE /api/v1/domains - Delete domains
	srv.router.HandleFunc("/api/v1/domains", srv.deleteDomainsHandler).
		Methods(http.MethodDelete)

	// GET /api/v1/domains - List domains
	srv.router.HandleFunc("/api/v1/domains", srv.listDomainsHandler).
		Methods(http.MethodGet)

	// GET /health - Health endpoint
	srv.router.HandleFunc("/health", srv.healthHandler).Methods(http.MethodGet)
}

// Body: {<domain>: <ip>}
func (srv *HttpServer) updateDomainsHandler(
	w http.ResponseWriter, req *http.Request) {

	defer req.Body.Close()

	dec := json.NewDecoder(req.Body)

	domains := map[string]string{}

	if err := dec.Decode(&domains); err != nil {
		httpLog.Errorf("Error decoding request body: %s", err)

		httpLib.ApiSendMessage(w, http.StatusBadRequest,
			"Error decoding request body: %s", err)

		return
	}

	if len(domains) > 0 {
		srv.params.DnsServer.updateDomains(domains)
	}

	httpLib.ApiSendData(w, http.StatusOK, nil)
}

// Body: [<domain>]
func (srv *HttpServer) deleteDomainsHandler(
	w http.ResponseWriter, req *http.Request) {

	defer req.Body.Close()

	dec := json.NewDecoder(req.Body)

	var domains []string

	if err := dec.Decode(&domains); err != nil {
		httpLog.Errorf("Error decoding request body: %s", err)

		httpLib.ApiSendMessage(w, http.StatusBadRequest,
			"Error decoding request body: %s", err)

		return
	}

	if len(domains) > 0 {
		srv.params.DnsServer.deleteDomains(domains)
	}

	httpLib.ApiSendData(w, http.StatusOK, nil)
}

func (srv *HttpServer) listDomainsHandler(
	w http.ResponseWriter, req *http.Request) {
	httpLib.ApiSendData(w, http.StatusOK, srv.params.DnsServer.getDomains())
}

func (srv *HttpServer) healthHandler(
	w http.ResponseWriter, req *http.Request) {
	httpLib.ApiSendData(w, http.StatusOK, nil)
}
