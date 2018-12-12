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

package lib

import (
	"fmt"
	"net"
	"sync"

	"github.com/pkg/errors"
)

type Port = uint16

type PortRange struct {
	min  Port
	max  Port
	next Port
	sync.Mutex
}

func NewPortRange(min, max Port) *PortRange {
	return &PortRange{
		min:  min,
		next: min,
		max:  max,
	}
}

func (pr *PortRange) NextPort() (Port, error) {
	pr.Lock()
	defer pr.Unlock()

	var span uint16

	for {
		if span >= pr.max-pr.min {
			return 0, errors.Errorf("No free ports found in range %d-%d: %d",
				pr.min, pr.max, span)
		}

		// Reset back to min, some ports may have been freed by now
		if pr.next > pr.max {
			pr.next = pr.min
		}

		l, err := net.Listen("tcp", fmt.Sprintf(":%d", pr.next))

		if err != nil {
			span++

			if IsErrAddrInUse(err) {
				continue
			}

			return 0, err
		}

		span = 0

		pr.next++

		port := l.Addr().(*net.TCPAddr).Port
		_ = l.Close()

		return uint16(port), nil
	}
}
