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

	"github.com/syhpoon/xenvman/pkg/tpl"
)

type interpolator struct {
	externalAddress string
	self            *tpl.Container
	selfPorts       map[uint16]uint16
	containers      []*tpl.Container
}

// Return a Container instance for the given template
func (ip *interpolator) Self() *tpl.Container {
	return ip.self
}

// Return an external address associated with xenvman api server
func (ip *interpolator) ExternalAddress() string {
	return ip.externalAddress
}

// Return an external port for the given internal one
func (ip *interpolator) SelfExposedPort(port int) uint16 {
	eport, ok := ip.selfPorts[uint16(port)]

	if !ok {
		panic(fmt.Sprintf("Port %d is not exposed for %s", port, ip.self.Name()))
	}

	return eport
}

// Return a list of containers which have one of the provided labels set
// A label is considered set when it has any non-empty label value
func (ip *interpolator) ContainersWithLabels(labels ...string) []*tpl.Container {
	var res []*tpl.Container

	ls := map[string]bool{}

	for _, l := range labels {
		ls[l] = true
	}

	for _, c := range ip.containers {
		for label := range c.Labels() {
			if ls[label] {
				res = append(res, c)
				break
			}
		}
	}

	return res
}

// Return a container possessing a given label.
// Empty value matches any label value
// If more than one containers match, one of them is returned in arbitrary order
func (ip *interpolator) ContainerWithLabel(label, value string) *tpl.Container {
	for _, c := range ip.containers {
		for l, v := range c.Labels() {
			if label == l {
				if value == "" || value == v {
					return c
				}
			}
		}
	}

	return nil
}
