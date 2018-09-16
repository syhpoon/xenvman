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

package def

type ContainerFileMounts struct {
	Path     string `json:"path"`
	Bytes    []byte `json:"bytes"`
	ReadOnly bool   `json:"readonly"`
}

type ContainerMounts struct {
	Files map[string]ContainerFileMounts
}

type Container struct {
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Environ map[string]string `json:"environ,omitempty"`
	Cmd     []string          `json:"cmd,omitempty"`
	Ports   []uint16          `json:"ports,omitempty"`
	Mounts  ContainerMounts   `json:"mounts,omitempty"`
}
