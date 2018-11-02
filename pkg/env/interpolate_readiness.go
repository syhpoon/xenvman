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

import "fmt"

// This one is used to interpolate readiness check parameters
type readinessInterpolator struct {
	externalAddress string
	ports           map[string]map[uint16]uint16
}

// Return an external address associated with xenvman api server
func (ri *readinessInterpolator) ExternalAddress() string {
	return ri.externalAddress
}

func (ri *readinessInterpolator) ExposedContainerPort(container string,
	port int) uint16 {
	p, ok := ri.ports[container]

	if !ok {
		panic(fmt.Sprintf("No ports for container %s", container))
	}

	eport, ok := p[uint16(port)]

	if !ok {
		panic(fmt.Sprintf("Port %d is not exposed for %s", port, container))
	}

	return eport
}