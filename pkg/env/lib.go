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

package env

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/syhpoon/xenvman/pkg/lib"
	"github.com/syhpoon/xenvman/pkg/tpl"
)

func newEnvId(name string) string {
	t := time.Now().Format("20060102150405")

	return fmt.Sprintf("%s-%s-%s", name, t, lib.NewIdShort())
}

func assignIps(ipn *lib.Net, conts []*tpl.Container) ([]string, map[string]string, error) {
	ips := make([]string, len(conts))
	hosts := map[string]string{}

	for i, cont := range conts {
		ip := ipn.NextIP()
		hostname := cont.Hostname()

		if ip == nil {
			return nil, nil, errors.Errorf("Unable to assign IP to container: network %s exhausted", ipn.Sub())
		}

		ips[i] = ip.String()
		hosts[hostname] = ip.String()

		envLog.Debugf("Assigned IP %s to %s", ips[i], hostname)
	}

	return ips, hosts, nil
}
