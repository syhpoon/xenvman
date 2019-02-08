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

	"github.com/pkg/errors"
	"github.com/syhpoon/xenvman/pkg/def"
)

type Env struct {
	*def.OutputEnv

	httpClient    http.Client
	serverAddress string
}

// Terminate/Delete environment
func (env *Env) Terminate() error {
	url := fmt.Sprintf("%s/api/v1/env/%s", env.serverAddress, env.Id)

	req, err := http.NewRequest(http.MethodDelete, url, nil)

	if err != nil {
		return errors.Wrapf(err, "Error creating HTTP request to %s", url)
	}

	resp, err := env.httpClient.Do(req)

	if err != nil {
		return errors.Wrapf(err, "Error making HTTP request to %s", url)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)

		return errors.Errorf("Unexpected HTTP response %d: %s",
			resp.StatusCode, string(body))
	}

	return nil
}

// Patch environment
func (env *Env) Patch(patch *def.PatchEnv) (*Env, error) {
	b, err := json.Marshal(patch)

	if err != nil {
		return nil, errors.Wrapf(err, "Error marshaling request body")
	}

	url := fmt.Sprintf("%s/api/v1/env/%s", env.serverAddress, env.Id)

	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(b))

	if err != nil {
		return nil, errors.Wrapf(err, "Error creating HTTP request to %s", url)
	}

	resp, err := env.httpClient.Do(req)

	if err != nil {
		return nil, errors.Wrapf(err, "Error making HTTP request to %s", url)
	}

	var e *def.OutputEnv

	if err := fetch(resp, &e); err != nil {
		return nil, errors.WithStack(err)
	}

	env.OutputEnv = e

	return env, nil
}

// Send a keepalive message
func (env *Env) Keepalive() error {
	url := fmt.Sprintf("%s/api/v1/env/%s/keepalive", env.serverAddress, env.Id)

	req, err := http.NewRequest(http.MethodPost, url, nil)

	if err != nil {
		return errors.Wrapf(err, "Error creating HTTP request to %s", url)
	}

	resp, err := env.httpClient.Do(req)

	if err != nil {
		return errors.Wrapf(err, "Error making HTTP request to %s", url)
	}

	if err := fetch(resp, nil); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (env *Env) String() string {
	b, _ := json.MarshalIndent(env, "", "   ")

	return string(b)
}
