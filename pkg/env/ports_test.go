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
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/syhpoon/xenvman/pkg/tpl"
)

func TestPorts(t *testing.T) {
	p := make(ports)

	cont1tpl01 := tpl.NewContainer("c1", "tpl1", 0)
	cont2tpl01 := tpl.NewContainer("c2", "tpl1", 0)
	cont1tpl11 := tpl.NewContainer("c1", "tpl1", 1)
	cont1tpl02 := tpl.NewContainer("c1", "tpl2", 0)
	cont1tpl03 := tpl.NewContainer("c1", "tpl3", 0)

	p.add(cont1tpl01, map[uint16]uint16{1: 1})
	require.Len(t, p["tpl1"], 1)
	require.Len(t, p["tpl1"][0], 1)

	p.add(cont2tpl01, map[uint16]uint16{2: 1})
	require.Len(t, p["tpl1"], 1)
	require.Len(t, p["tpl1"][0], 2)

	p.add(cont1tpl11, map[uint16]uint16{10: 10})
	require.Len(t, p["tpl1"], 2)
	require.Len(t, p["tpl1"][1], 1)

	p.add(cont1tpl02, map[uint16]uint16{1: 2})
	p.add(cont1tpl03, map[uint16]uint16{1: 3})

	require.Len(t, p["tpl2"], 1)
	require.Len(t, p["tpl3"], 1)

	require.Equal(t, p["tpl1"][0]["c1"][1], uint16(1))
	require.Equal(t, p["tpl1"][0]["c2"][2], uint16(1))
	require.Equal(t, p["tpl1"][1]["c1"][10], uint16(10))
	require.Equal(t, p["tpl2"][0]["c1"][1], uint16(2))
	require.Equal(t, p["tpl3"][0]["c1"][1], uint16(3))
}
