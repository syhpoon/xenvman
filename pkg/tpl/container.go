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

package tpl

type ContainerFileMounts struct {
	Path     string
	Bytes    []byte
	ReadOnly bool
}

type ContainerMounts struct {
	Files map[string]ContainerFileMounts
}

type Container struct {
	name    string
	image   string
	environ map[string]string
	cmd     []string
	ports   []uint16
	mounts  ContainerMounts
}

func (cont *Container) SetEnv(k, v string) {
	cont.environ[k] = v
}

func (cont *Container) SetCmd(cmd ...string) {
	cont.cmd = cmd
}

func (cont *Container) SetPorts(ports ...uint16) {
	cont.ports = ports
}

func (cont *Container) Name() string {
	return cont.name
}

func (cont *Container) Image() string {
	return cont.image
}

func (cont *Container) Ports() []uint16 {
	return cont.ports
}

func (cont *Container) Environ() map[string]string {
	return cont.environ
}

func (cont *Container) Cmd() []string {
	return cont.cmd
}
