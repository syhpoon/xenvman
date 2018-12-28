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
	"net"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/syhpoon/xenvman/pkg/lib"
	"github.com/syhpoon/xenvman/pkg/logger"
)

var readinessNetLog = logger.GetLogger("xenvman.pkg.tpl.readiness_net")

func readinessCheckNetInit(params map[string]interface{}) ReadinessCheck {
	tcpParams := &readinessCheckNetParams{}

	if err := mapstructure.Decode(params, tcpParams); err != nil {
		panic(fmt.Sprintf("Invalid net readiness check params: %+v",
			errors.WithStack(err)))
	}

	return &readinessCheckNet{
		params: *tcpParams,
	}
}

type readinessCheckNetParams struct {
	Protocol      string `mapstructure:"protocol"`
	Address       string `mapstructure:"address"`
	RetryLimit    int    `mapstructure:"retry_limit"`
	RetryInterval string `mapstructure:"retry_interval"`
}

type readinessCheckNet struct {
	params        readinessCheckNetParams
	retryInterval time.Duration
}

func (rnet *readinessCheckNet) init() error {
	if rnet.params.RetryInterval != "" {
		dur, err := time.ParseDuration(rnet.params.RetryInterval)

		if err != nil {
			return errors.WithStack(err)
		} else {
			rnet.retryInterval = dur
		}
	} else {
		rnet.retryInterval = 2 * time.Second
	}

	if rnet.params.RetryLimit == 0 {
		rnet.params.RetryLimit = 5
	}

	return nil
}

func (rnet *readinessCheckNet) InterpolateParameters(data interface{}) error {
	if address, err := lib.Interpolate(rnet.params.Address, data); err != nil {
		return errors.WithStack(err)
	} else {
		rnet.params.Address = address
	}

	return rnet.init()
}

func (rnet *readinessCheckNet) Wait(ctx context.Context, success bool) bool {
	for i := 0; rnet.params.RetryLimit == 0 || i < rnet.params.RetryLimit; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return false
			case <-time.After(rnet.retryInterval):
			}
		}

		con, err := net.Dial(rnet.params.Protocol, rnet.params.Address)

		if (success && err != nil) || (!success && err == nil) {
			continue
		}

		if con != nil {
			_ = con.Close()
		}

		return true
	}

	return false
}

func (rnet *readinessCheckNet) String() string {
	return "net"
}
