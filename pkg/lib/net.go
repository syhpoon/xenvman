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
	"net"
	"os"
	"syscall"

	"encoding/binary"

	"github.com/pkg/errors"
)

type Net struct {
	ip     net.IP
	lastIP net.IP
	ipnet  *net.IPNet
}

// Only works for IP4 for now
func ParseNet(sub string) (*Net, error) {
	ip, ipnet, err := net.ParseCIDR(sub)

	if err != nil {
		return nil, errors.Wrapf(err, "Error parsing net: %s", sub)
	}

	lastip := make(net.IP, len(ip.To4()))

	binary.BigEndian.PutUint32(lastip,
		binary.BigEndian.Uint32(ip.To4())|^binary.BigEndian.Uint32(
			net.IP(ipnet.Mask).To4()))

	return &Net{
		ip:     ip.To4(),
		lastIP: lastip.To4(),
		ipnet:  ipnet,
	}, nil
}

func (n *Net) NextIP() net.IP {
	n.ip[3]++

	if n.ip.Equal(n.lastIP) || !n.ipnet.Contains(n.ip) {
		return nil
	}

	return n.ip
}

func NetsOverlap(n1, n2 *net.IPNet) bool {
	return n1.Contains(n2.IP) || n2.Contains(n1.IP)
}

func IsErrAddrInUse(err error) bool {
	if err, ok := err.(*net.OpError); ok {
		if err, ok := err.Err.(*os.SyscallError); ok {
			return err.Err == syscall.EADDRINUSE
		}
	}

	return false
}
