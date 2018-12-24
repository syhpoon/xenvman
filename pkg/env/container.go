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

type container struct {
	ports map[uint16]uint16
	ip    string
	cont  *tpl.Container
}

func container2interpolate(cont *tpl.Container,
	ports map[uint16]uint16, ip string) *container {

	return &container{
		cont:  cont,
		ports: ports,
		ip:    ip,
	}
}

func (cont *container) IP() string {
	return cont.ip
}

func (cont *container) Hostname() string {
	return cont.cont.Hostname()
}

func (cont *container) Name() string {
	return cont.cont.Name()
}

func (cont *container) GetLabel(label string) string {
	return cont.cont.GetLabel(label)
}

// Return an external port for the given internal one
func (cont *container) ExposedPort(port int) uint16 {
	eport, ok := cont.ports[uint16(port)]

	if !ok {
		panic(fmt.Sprintf("Port %d is not exposed for %s", port, cont.Name()))
	}

	return eport
}
