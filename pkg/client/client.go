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

package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/syhpoon/xenvman/pkg/def"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Params struct {
	ServerAddress  string
	RequestTimeout time.Duration
}

// Client structure represents a logical session with xenvman API server
type Client struct {
	httpClient http.Client
	params     Params
}

// Create a new xenvman client
// If params.ServerAddress is not set, a value from
// XENV_API_SERVER environment variable will be used
func New(params Params) *Client {
	hcl := http.Client{
		Timeout: params.RequestTimeout,
	}

	if params.ServerAddress == "" {
		params.ServerAddress = os.Getenv("XENV_API_SERVER")
	}

	if params.ServerAddress == "" {
		params.ServerAddress = "http://localhost:9876"
	}

	if strings.HasSuffix(params.ServerAddress, "/") {
		params.ServerAddress = params.ServerAddress[:len(params.ServerAddress)-1]
	}

	return &Client{
		httpClient: hcl,
		params:     params,
	}
}

// Create xenvman client or panic otherwise
func (cl *Client) MustCreateEnv(envDef *def.InputEnv) *Env {
	env, err := cl.NewEnv(envDef)

	if err != nil {
		panic(errors.Wrapf(err, "Error creating env"))
	}

	return env
}

// Create a new environment
func (cl *Client) NewEnv(envDef *def.InputEnv) (*Env, error) {
	url := fmt.Sprintf("%s/api/v1/env", cl.params.ServerAddress)

	b, err := json.Marshal(envDef)

	if err != nil {
		return nil, errors.Wrapf(err, "Error marshaling request body")
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))

	if err != nil {
		return nil, errors.Wrapf(err, "Error creating HTTP request to %s", url)
	}

	resp, err := cl.httpClient.Do(req)

	if err != nil {
		return nil, errors.Wrapf(err, "Error making HTTP request to %s", url)
	}

	outEnv := def.OutputEnv{}
	e := &Env{}

	if err := fetch(resp, &outEnv); err != nil {
		return nil, errors.WithStack(err)
	}

	e.OutputEnv = &outEnv
	e.httpClient = cl.httpClient
	e.serverAddress = cl.params.ServerAddress

	return e, nil
}

// List currently active environments
func (cl *Client) ListEnvs() ([]*def.OutputEnv, error) {
	url := fmt.Sprintf("%s/api/v1/env", cl.params.ServerAddress)

	resp, err := cl.httpClient.Get(url)

	if err != nil {
		return nil, errors.Wrapf(err, "Error making HTTP request to %s", url)
	}

	var e []*def.OutputEnv

	if err := fetch(resp, &e); err != nil {
		return nil, errors.WithStack(err)
	}

	return e, nil
}

// List available templates
func (cl *Client) ListTemplates() (map[string]*def.TplInfo, error) {
	url := fmt.Sprintf("%s/api/v1/tpl", cl.params.ServerAddress)

	resp, err := cl.httpClient.Get(url)

	if err != nil {
		return nil, errors.Wrapf(err, "Error making HTTP request to %s", url)
	}

	r := map[string]*def.TplInfo{}

	if err := fetch(resp, &r); err != nil {
		return nil, errors.WithStack(err)
	}

	return r, nil
}

// Get environment info
func (cl *Client) GetEnvInfo(id string) (*def.OutputEnv, error) {
	url := fmt.Sprintf("%s/api/v1/env/%s", cl.params.ServerAddress, id)

	resp, err := cl.httpClient.Get(url)

	if err != nil {
		return nil, errors.Wrapf(err, "Error making HTTP request to %s", url)
	}

	var e *def.OutputEnv

	if err := fetch(resp, &e); err != nil {
		return nil, errors.WithStack(err)
	}

	return e, nil
}

func fetch(resp *http.Response, dst interface{}) error {
	//noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("Unexpected HTTP response %d: %s",
			resp.StatusCode, resp.Status)
	}

	if dst != nil {
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return errors.Wrapf(err, "Error reading response body")
		}

		r := def.ApiResponse{}

		if err := json.Unmarshal(body, &r); err != nil {
			return errors.Wrapf(err, "Error parsing response body")
		}

		if r.Data == nil {
			return errors.Errorf("Returned data is nil")
		}

		if err := mapstructure.Decode(r.Data, dst); err != nil {
			return errors.Wrapf(err, "Error decoding response data")
		}
	}

	return nil
}
