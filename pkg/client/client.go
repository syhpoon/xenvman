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
	"os"
	"strings"
	"time"

	"io/ioutil"
	"net/http"

	"github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/syhpoon/xenvman/pkg/def"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Params struct {
	ServerAddress  string
	RequestTimeout time.Duration
}

type Client struct {
	httpClient http.Client
	params     Params
}

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

func (cl *Client) MustCreateEnv(envDef *def.Env) *Env {
	env, err := cl.NewEnv(envDef)

	if err != nil {
		panic(errors.Wrapf(err, "Error creating env"))
	}

	return env
}

func (cl *Client) NewEnv(envDef *def.Env) (*Env, error) {
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

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, errors.Wrapf(err, "Error reading response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("Unexpected HTTP response %d: %s",
			resp.StatusCode, string(body))
	}

	r := apiResponse{}

	if err := json.Unmarshal(body, &r); err != nil {
		return nil, errors.Wrapf(err, "Error parsing response body")
	}

	e := &Env{}

	if r.Env != nil {
		e.Exported = r.Env
		e.httpClient = cl.httpClient
		e.serverAddress = cl.params.ServerAddress
	}

	return e, nil
}
