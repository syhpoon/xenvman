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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIPNext(t *testing.T) {
	ipn, err := ParseNet("10.0.0.0/30")
	require.Nil(t, err)

	ip1 := ipn.NextIP()
	require.NotNil(t, ip1)
	require.Equal(t, ip1.To4(), net.IP{10, 0, 0, 1}.To4())

	ip2 := ipn.NextIP()
	require.NotNil(t, ip2)
	require.Equal(t, ip2.To4(), net.IP{10, 0, 0, 2}.To4())

	ip3 := ipn.NextIP()
	require.Nil(t, ip3)
}

func TestNetsOverlap(t *testing.T) {
	_, n1, _ := net.ParseCIDR("10.0.0.0/24")
	_, n2, _ := net.ParseCIDR("10.0.0.0/30")
	require.True(t, NetsOverlap(n1, n2))

	_, n1, _ = net.ParseCIDR("10.0.0.0/30")
	_, n2, _ = net.ParseCIDR("10.0.0.4/30")
	require.False(t, NetsOverlap(n1, n2))
}
