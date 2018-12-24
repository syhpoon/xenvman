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
	"github.com/syhpoon/xenvman/pkg/lib"
	"github.com/syhpoon/xenvman/pkg/tpl"
)

func TestInterpolateEnvContainersWithLabels(t *testing.T) {
	input := `{{- range .ContainersWithLabels "exposed" -}}
name = "{{.GetLabel "exposed-service-name"}}"
address = "{{.Hostname}}_{{.GetLabel "exposed-service-port" }}"
{{- end -}}`

	res := `name = "c1"
address = "c1.0.tpl.xenv_2000"`

	c1 := tpl.NewContainer("c1", "tpl", 0)
	c1.SetLabel("exposed", "true")
	c1.SetLabel("exposed-service-name", "c1")
	c1.SetLabel("exposed-service-port", "2000")

	c2 := tpl.NewContainer("c2", "tpl", 1)

	conts := []*tpl.Container{c1, c2}

	ip := &interpolator{
		containers: conts,
		ports:      make(ports),
	}

	out, err := lib.Interpolate(input, ip)

	require.Nil(t, err)
	require.Equal(t, out, res)
}

func TestInterpolateEnvContainerWithLabel(t *testing.T) {
	input := `
{{- with .ContainerWithLabel "auth" "" -}}
auth = "{{.Hostname}}:{{.GetLabel "auth-port"}}"
{{- end -}}`

	res := `auth = "c1.0.tpl.xenv:2000"`

	c1 := tpl.NewContainer("c1", "tpl", 0)
	c1.SetLabel("auth", "true")
	c1.SetLabel("auth-port", "2000")

	c2 := tpl.NewContainer("c2", "tpl", 1)

	conts := []*tpl.Container{c1, c2}

	ip := &interpolator{
		containers: conts,
		ports:      make(ports),
	}

	out, err := lib.Interpolate(input, ip)

	require.Nil(t, err)
	require.Equal(t, out, res)
}
