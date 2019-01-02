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

package server

import (
	"context"
	"net"
	"sync"
	"time"

	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/syhpoon/xenvman/pkg/conteng"
	"github.com/syhpoon/xenvman/pkg/def"
	"github.com/syhpoon/xenvman/pkg/env"
	"github.com/syhpoon/xenvman/pkg/lib"
	"github.com/syhpoon/xenvman/pkg/logger"
)

var serverLog = logger.GetLogger("xenvman.pkg.server.server")

type Params struct {
	Listener         net.Listener
	WriteTimeout     time.Duration
	ReadTimeout      time.Duration
	ContEng          conteng.ContainerEngine
	PortRange        *lib.PortRange
	BaseTplDir       string
	BaseWsDir        string
	BaseMountDir     string
	ExportAddress    string
	TLSCertFile      string
	TLSKeyFile       string
	AuthBackend      AuthBackend
	Ctx              context.Context
	CengCtx          context.Context
	DefaultKeepalive time.Duration
}

func DefaultParams(ctx context.Context) Params {
	return Params{
		WriteTimeout: 5 * time.Minute,
		ReadTimeout:  5 * time.Minute,
		Ctx:          ctx,
	}
}

type Server struct {
	router *mux.Router
	server http.Server
	params Params
	envs   map[string]*env.Env
	sync.RWMutex
}

func New(params Params) *Server {
	router := mux.NewRouter()

	return &Server{
		router: router,
		server: http.Server{
			Handler:      router,
			WriteTimeout: params.WriteTimeout,
			ReadTimeout:  params.ReadTimeout,
		},
		params: params,
		envs:   map[string]*env.Env{},
	}
}

func (s *Server) Run(wg *sync.WaitGroup, errch chan<- error) {
	// API endpoints
	s.setupHandlers()

	ctx, cancel := context.WithCancel(s.params.Ctx)
	defer cancel()

	// Shutdown http server upon global signal
	go func() {
		<-ctx.Done()
		_ = s.server.Shutdown(ctx)

		s.Lock()

		for _, e := range s.envs {
			_ = e.Terminate()
		}

		s.Unlock()
		wg.Done()
	}()

	useTls := s.params.TLSCertFile != "" && s.params.TLSKeyFile != ""

	mode := ""

	if useTls {
		mode = " [TLS]"
	}

	serverLog.Infof("Starting xenvman server%s at %s",
		mode, s.params.Listener.Addr().String())

	var err error

	if useTls {
		err = s.server.ServeTLS(s.params.Listener,
			s.params.TLSCertFile, s.params.TLSKeyFile)
	} else {
		err = s.server.Serve(s.params.Listener)
	}

	if err != nil {
		errch <- errors.WithStack(err)
	}
}

func (s *Server) setupHandlers() {
	hf := func(h http.HandlerFunc) http.HandlerFunc {
		if s.params.AuthBackend != nil {
			return apiAuthMiddleware(h, s.params.AuthBackend)
		} else {
			return h
		}
	}

	// POST /api/v1/env - Create a new environment
	s.router.HandleFunc("/api/v1/env", hf(s.createEnvHandler)).
		Methods(http.MethodPost)

	// DELETE /api/v1/env/{id} - Delete an environment
	s.router.HandleFunc("/api/v1/env/{id}", hf(s.deleteEnvHandler)).
		Methods(http.MethodDelete)

	// PATCH /api/v1/env/{id} - Patch an environment
	s.router.HandleFunc("/api/v1/env/{id}", hf(s.patchEnvHandler)).
		Methods(http.MethodPatch)

	// POST /api/v1/env/{id}/keepalive - Keep alive an environment
	s.router.HandleFunc("/api/v1/env/{id}/keepalive",
		hf(s.keepaliveEnvHandler)).Methods(http.MethodPost)

	// Prometheus metrics
	s.router.Handle("/metrics", promhttp.Handler())

	//TODO: API doc
	//s.router.HandleFunc("/apidoc", s.handleRoot)

	//TODO: Web UI on root?
	//s.router.HandleFunc("/", s.handleRoot)
}

// Body: envDef structure
func (s *Server) createEnvHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)

	if err != nil {
		serverLog.Errorf("Error reading request body: %s", err)

		ApiSendMessage(w, http.StatusBadRequest, "Error reading request body: %s", err)

		return
	}

	edef := def.InputEnv{}

	if err = json.Unmarshal(body, &edef); err != nil {
		serverLog.Errorf("Error decoding request body: %s", err)

		ApiSendMessage(w, http.StatusBadRequest, "Error decoding request body: %s", err)

		return
	}

	if validErr := edef.Validate(); validErr != nil {
		serverLog.Errorf("Error validating env def: %s", validErr)

		ApiSendMessage(w, http.StatusBadRequest, "Error validating env: %s",
			validErr)

		return
	}

	e, err := env.NewEnv(env.Params{
		EnvDef:           &edef,
		ContEng:          s.params.ContEng,
		PortRange:        s.params.PortRange,
		BaseTplDir:       s.params.BaseTplDir,
		BaseWsDir:        s.params.BaseWsDir,
		BaseMountDir:     s.params.BaseMountDir,
		ExportAddress:    s.params.ExportAddress,
		DefaultKeepAlive: def.Duration(s.params.DefaultKeepalive),
		Ctx:              s.params.CengCtx,
	})

	if err != nil {
		serverLog.Errorf("Error creating env: %+v", err)

		ApiSendMessage(w, http.StatusBadRequest, "Error creating env: %s", err)

		return
	}

	s.Lock()
	s.envs[e.Id()] = e
	s.Unlock()

	ApiSendData(w, http.StatusOK, e.Export())
}

func (s *Server) deleteEnvHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	vars := mux.Vars(req)
	id := vars["id"]

	s.RLock()
	e, ok := s.envs[id]
	s.RUnlock()

	if !ok {
		serverLog.Errorf("Env not found", id)

		ApiSendMessage(w, http.StatusNotFound, "Env not found")

		return
	}

	err := e.Terminate()

	if err != nil {
		serverLog.Errorf("Error terminating env %s: %+v", id, err)

		ApiSendMessage(w, http.StatusBadRequest, "Error terminating env: %s", err)

		return
	}

	s.Lock()
	delete(s.envs, id)
	s.Unlock()

	ApiSendMessage(w, http.StatusOK, "Env deleted")
}

func (s *Server) patchEnvHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	vars := mux.Vars(req)
	id := vars["id"]

	s.RLock()
	e, ok := s.envs[id]
	s.RUnlock()

	if !ok {
		serverLog.Errorf("Env not found", id)

		ApiSendMessage(w, http.StatusNotFound, "Env not found")

		return
	}

	patchDef := def.PatchEnv{}

	err := json.NewDecoder(req.Body).Decode(&patchDef)

	if err != nil {
		serverLog.Errorf("Error parsing patch request body for %s: %s",
			id, err)

		ApiSendMessage(w, http.StatusBadRequest, "Error parsing request body")

		return
	}

	// 1. Check if there are containers to stop
	if len(patchDef.StopContainers) > 0 {
		if err := e.StopContainers(patchDef.StopContainers); err != nil {

			serverLog.Errorf("Error stopping containers for %s: %+v", id, err)

			ApiSendMessage(w, http.StatusBadRequest, "Error stopping containers")

			return
		}
	}

	// 2. Check if there are containers to restart
	if len(patchDef.RestartContainers) > 0 {
		if err := e.RestartContainers(patchDef.RestartContainers); err != nil {
			serverLog.Errorf("Error restarting containers for %s: %+v",
				id, err)

			ApiSendMessage(w, http.StatusBadRequest,
				"Error restarting containers")

			return
		}
	}

	// 3. Check if there are new templates to add
	if len(patchDef.Templates) > 0 {
		if err := e.ApplyTemplates(patchDef.Templates, false, true); err != nil {
			serverLog.Errorf("Error adding new templates for %s: %+v",
				id, err)

			ApiSendMessage(w, http.StatusBadRequest,
				"Error adding new templates")

			return
		}
	}

	ApiSendMessage(w, http.StatusOK, "Env updated")
}

func (s *Server) keepaliveEnvHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	vars := mux.Vars(req)
	id := vars["id"]

	s.RLock()
	e, ok := s.envs[id]
	s.RUnlock()

	if !ok {
		serverLog.Errorf("Env not found", id)

		ApiSendMessage(w, http.StatusNotFound, "Env not found")

		return
	}

	if !e.IsAlive() {
		serverLog.Errorf("Env is terminating %s", id)

		s.Lock()
		delete(s.envs, id)
		s.Unlock()

		ApiSendMessage(w, http.StatusNotFound, "Env is terminating")

		return
	}

	e.KeepAlive()

	ApiSendMessage(w, http.StatusOK, "")
}
