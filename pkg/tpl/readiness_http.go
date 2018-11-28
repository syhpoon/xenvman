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
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

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

func (rtcp *readinessCheckHttp) init() error {
	var err error

	if rtcp.params.Body != "" {
		if rtcp.body, err = regexp.Compile(rtcp.params.Body); err != nil {
			return errors.WithStack(err)
		}
	}

	for i, hdrs := range rtcp.params.Headers {
		if rtcp.headers[i] == nil {
			rtcp.headers[i] = map[string]*regexp.Regexp{}
		}

		for k, v := range hdrs {
			if rtcp.headers[i][k], err = regexp.Compile(v); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	if rtcp.params.RetryInterval != "" {
		dur, err := time.ParseDuration(rtcp.params.RetryInterval)

		if err != nil {
			return errors.WithStack(err)
		} else {
			rtcp.retryInterval = dur
		}
	} else {
		rtcp.retryInterval = 2 * time.Second
	}

	if rtcp.params.RetryLimit == 0 {
		rtcp.params.RetryLimit = 5
	}

	return nil
}

func (rtcp *readinessCheckHttp) InterpolateParameters(data interface{}) error {
	var err error

	if url, err := lib.Interpolate(rtcp.params.Url, data); err != nil {
		return errors.WithStack(err)
	} else {
		rtcp.params.Url = url
	}

	if rtcp.params.Body != "" {
		if body, err := lib.Interpolate(rtcp.params.Body, data); err != nil {
			return errors.WithStack(err)
		} else {
			rtcp.params.Body = body
		}
	}

	for _, hdrs := range rtcp.params.Headers {
		for k, v := range hdrs {
			if hdrs[k], err = lib.Interpolate(v, data); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return rtcp.init()
}

func (rtcp *readinessCheckHttp) WaitUntilReady(ctx context.Context) bool {
	cl := http.Client{Timeout: 2 * time.Second}

	i := 0

	for rtcp.params.RetryLimit == 0 || i < rtcp.params.RetryLimit {
		if i > 0 {
			select {
			case <-ctx.Done():
				return false
			case <-time.After(rtcp.retryInterval):
			}
		}

		i += 1
		resp, err := cl.Get(rtcp.params.Url)

		if err != nil {
			readinessHttpLog.Warningf("Error running HTTP readiness check to %s: %s", rtcp.params.Url, err)

			continue
		}

		// Match status codes
		if len(rtcp.params.Codes) > 0 {
			found := false

			for _, code := range rtcp.params.Codes {
				if resp.StatusCode == code {
					found = true
					break
				}
			}

			if !found {
				continue
			}
		}

		// Match body
		if rtcp.body != nil {
			body, err := ioutil.ReadAll(resp.Body)

			if err != nil {
				readinessHttpLog.Debugf("Error reading response body: %s", err)

				continue
			}

			if !rtcp.body.Match(body) {
				readinessHttpLog.Debugf("Body mismatch: %s", rtcp.body.String())

				continue
			}
		}

		resp.Body.Close()

		//TODO: Match headers

		return true
	}

	return false
}

func (rtcp *readinessCheckHttp) String() string {
	return "http"
}
