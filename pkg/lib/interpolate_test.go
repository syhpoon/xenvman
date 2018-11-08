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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInterpolateBasic(t *testing.T) {
	template := `{{.A}} - {{.B}}`
	data := map[string]interface{}{
		"A": "aaa",
		"B": "bbb",
	}

	r, err := Interpolate(template, data)
	require.Nil(t, err)
	require.Equal(t, r, `aaa - bbb`)
}

func TestInterpolateRange(t *testing.T) {
	template := `{{range .Items}}{{.Name}}{{end}}`
	data := map[string]interface{}{
		"Items": []map[string]interface{}{
			{"Name": "x"},
			{"Name": "y"},
			{"Name": "z"},
		},
	}

	r, err := Interpolate(template, data)
	require.Nil(t, err)
	require.Equal(t, r, `xyz`)
}
