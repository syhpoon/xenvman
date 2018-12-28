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

package tpl

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"io/ioutil"
	"net/http"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/syhpoon/xenvman/pkg/lib"
	"github.com/syhpoon/xenvman/pkg/logger"
)

var readinessHttpLog = logger.GetLogger("xenvman.pkg.tpl.readiness_http")

func readinessCheckHttpInit(params map[string]interface{}) ReadinessCheck {
	httpParams := &readinessCheckHttpParams{}

	if err := mapstructure.Decode(params, httpParams); err != nil {
		panic(fmt.Sprintf("Invalid HTTP readiness check params: %+v",
			errors.WithStack(err)))
	}

	return &readinessCheckHttp{
		params:  *httpParams,
		headers: make([]map[string]*regexp.Regexp, len(httpParams.Headers)),
	}
}

type readinessCheckHttpParams struct {
	Url string `mapstructure:"url"`
	// At least one code must match
	Codes []int `mapstructure:"codes"`
	// [Header name -> value regexp]
	// Values within the same map are matched in a conjuctive way (AND)
	// Values from different maps are matched in a disjunctive way (OR)
	Headers       []map[string]string `mapstructure:"headers"`
	Body          string              `mapstructure:"body"`
	RetryLimit    int                 `mapstructure:"retry_limit"`
	RetryInterval string              `mapstructure:"retry_interval"`
}

type readinessCheckHttp struct {
	params        readinessCheckHttpParams
	body          *regexp.Regexp
	headers       []map[string]*regexp.Regexp
	retryInterval time.Duration
}

func (rh *readinessCheckHttp) init() error {
	var err error

	if rh.params.Body != "" {
		if rh.body, err = regexp.Compile(rh.params.Body); err != nil {
			return errors.WithStack(err)
		}
	}

	for i, hdrs := range rh.params.Headers {
		if rh.headers[i] == nil {
			rh.headers[i] = map[string]*regexp.Regexp{}
		}

		for k, v := range hdrs {
			if rh.headers[i][k], err = regexp.Compile(v); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	if rh.params.RetryInterval != "" {
		dur, err := time.ParseDuration(rh.params.RetryInterval)

		if err != nil {
			return errors.WithStack(err)
		} else {
			rh.retryInterval = dur
		}
	} else {
		rh.retryInterval = 2 * time.Second
	}

	if rh.params.RetryLimit == 0 {
		rh.params.RetryLimit = 5
	}

	return nil
}

func (rh *readinessCheckHttp) InterpolateParameters(data interface{}) error {
	var err error

	if url, err := lib.Interpolate(rh.params.Url, data); err != nil {
		return errors.WithStack(err)
	} else {
		rh.params.Url = url
	}

	if rh.params.Body != "" {
		if body, err := lib.Interpolate(rh.params.Body, data); err != nil {
			return errors.WithStack(err)
		} else {
			rh.params.Body = body
		}
	}

	for _, hdrs := range rh.params.Headers {
		for k, v := range hdrs {
			if hdrs[k], err = lib.Interpolate(v, data); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return rh.init()
}

func (rh *readinessCheckHttp) Wait(ctx context.Context, success bool) bool {
	cl := http.Client{Timeout: 2 * time.Second}

	i := 0

	for rh.params.RetryLimit == 0 || i < rh.params.RetryLimit {
		if i > 0 {
			select {
			case <-ctx.Done():
				return false
			case <-time.After(rh.retryInterval):
			}
		}

		i += 1
		resp, err := cl.Get(rh.params.Url)

		if (success && err != nil) || (!success && err == nil) {
			continue
		}

		// Match status codes
		if len(rh.params.Codes) > 0 {
			found := false

			for _, code := range rh.params.Codes {
				if (success && resp.StatusCode == code) ||
					(!success && resp.StatusCode != code) {
					found = true
					break
				}
			}

			if !found {
				continue
			}
		}

		// Match body
		if rh.body != nil {
			body, err := ioutil.ReadAll(resp.Body)

			if err != nil {
				readinessHttpLog.Debugf("Error reading response body: %s", err)

				continue
			}

			match := rh.body.Match(body)

			if (success && !match) || (!success && match) {
				continue
			}
		}

		_ = resp.Body.Close()

		//TODO: Match headers

		return true
	}

	return false
}

func (rh *readinessCheckHttp) String() string {
	return "http"
}
