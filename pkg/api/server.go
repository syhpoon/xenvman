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

package api

import (
	"context"
	"net"
	"sync"
	"time"

	"net/http"

	"github.com/gorilla/mux"
	"github.com/syhpoon/xenvman/pkg/logger"
	"github.com/syhpoon/xenvman/pkg/repo"
	"encoding/json"
	"io/ioutil"
)

var serverLog = logger.GetLogger("xenvman.cmd.api")

type ServerParams struct {
	Listener     net.Listener
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
	Repos        map[string]repo.Repo
	Ctx          context.Context
}

func DefaultServerParams(ctx context.Context) ServerParams {
	return ServerParams{
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
		Ctx:          ctx,
	}
}

type Server struct {
	router *mux.Router
	server http.Server
	params ServerParams
}

func NewServer(params ServerParams) *Server {
	router := mux.NewRouter()

	return &Server{
		router: router,
		server: http.Server{
			Handler:      router,
			WriteTimeout: params.WriteTimeout,
			ReadTimeout:  params.ReadTimeout,
		},
		params: params,
	}
}

func (s *Server) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	// API endpoints
	s.setupHandlers()

	// Shutdown http server upon global signal
	go func() {
		<-s.params.Ctx.Done()
		s.server.Shutdown(s.params.Ctx)
	}()

	serverLog.Infof("Starting API server at %s",
		s.params.Listener.Addr().String())

	s.server.Serve(s.params.Listener)
}

func (s *Server) setupHandlers() {
	// POST /api/v1/env - Create a new environment
	s.router.HandleFunc("/api/v1/env", s.createEnvHandler).
		Methods(http.MethodPost)

	//TODO: Show api doc on root
	//s.router.HandleFunc("/", s.handleRoot)
}

// Body: envDef structure
func (s *Server) createEnvHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)

	_ = err

	edef := envDef{}

	err = json.Unmarshal(body, &edef)
}
