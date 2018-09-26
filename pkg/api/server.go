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

	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/syhpoon/xenvman/pkg/def"
	"github.com/syhpoon/xenvman/pkg/env"
	"github.com/syhpoon/xenvman/pkg/lib"

	"github.com/syhpoon/xenvman/pkg/conteng"
	"github.com/syhpoon/xenvman/pkg/logger"
)

var serverLog = logger.GetLogger("xenvman.pkg.api.server")

type ServerParams struct {
	Listener     net.Listener
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
	ContEng      conteng.ContainerEngine
	PortRange    *lib.PortRange
	BaseTplDir   string
	BaseWsDir    string
	Ctx          context.Context
}

func DefaultServerParams(ctx context.Context) ServerParams {
	return ServerParams{
		WriteTimeout: 5 * time.Minute,
		ReadTimeout:  60 * time.Second,
		Ctx:          ctx,
	}
}

type Server struct {
	router *mux.Router
	server http.Server
	params ServerParams
	envs   map[string]*env.Env
	sync.RWMutex
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
		envs:   map[string]*env.Env{},
	}
}

func (s *Server) Run(wg *sync.WaitGroup) {
	// API endpoints
	s.setupHandlers()

	// Shutdown http server upon global signal
	go func() {
		<-s.params.Ctx.Done()
		_ = s.server.Shutdown(s.params.Ctx)

		s.Lock()

		for _, e := range s.envs {
			_ = e.Terminate()
		}

		s.Unlock()

		wg.Done()
	}()

	serverLog.Infof("Starting API server at %s",
		s.params.Listener.Addr().String())

	if err := s.server.Serve(s.params.Listener); err != nil {
		panic(err.Error())
	}
}

func (s *Server) setupHandlers() {
	// POST /api/v1/env - Create a new environment
	s.router.HandleFunc("/api/v1/env", s.createEnvHandler).
		Methods(http.MethodPost)

	// DELETE /api/v1/env/{id} - Delete an environment
	s.router.HandleFunc("/api/v1/env/{id}", s.deleteEnvHandler).
		Methods(http.MethodDelete)

	// POST /api/v1/env/{id}/keepalive - Keep alive an environment
	s.router.HandleFunc("/api/v1/env/{id}/keepalive", s.keepaliveEnvHandler).
		Methods(http.MethodPost)

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

	edef := def.Env{}

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
		EnvDef:     &edef,
		ContEng:    s.params.ContEng,
		PortRange:  s.params.PortRange,
		BaseTplDir: s.params.BaseTplDir,
		BaseWsDir:  s.params.BaseWsDir,
		Ctx:        s.params.Ctx,
	})

	if err != nil {
		serverLog.Errorf("Error creating env: %+v", err)

		ApiSendMessage(w, http.StatusBadRequest, "Error creating env: %s", err)

		return
	}

	s.Lock()
	s.envs[e.Id] = e
	s.Unlock()

	ApiSendData(w, http.StatusOK, e)
}

func (s *Server) deleteEnvHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	vars := mux.Vars(req)
	id := vars["id"]

	s.RLock()
	env, ok := s.envs[id]
	s.RUnlock()

	if !ok {
		serverLog.Errorf("Env not found", id)

		ApiSendMessage(w, http.StatusNotFound, "Env not found")

		return
	}

	err := env.Terminate()

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

func (s *Server) keepaliveEnvHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	vars := mux.Vars(req)
	id := vars["id"]

	s.RLock()
	env, ok := s.envs[id]
	s.RUnlock()

	if !ok {
		serverLog.Errorf("Env not found", id)

		ApiSendMessage(w, http.StatusNotFound, "Env not found")

		return
	}

	if !env.IsAlive() {
		serverLog.Errorf("Env is terminating %s", id)

		s.Lock()
		delete(s.envs, id)
		s.Unlock()

		ApiSendMessage(w, http.StatusNotFound, "Env is terminating")

		return
	}

	env.KeepAlive()

	ApiSendMessage(w, http.StatusOK, "")
}
