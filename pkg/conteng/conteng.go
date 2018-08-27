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

package conteng

import "io"

type NetworkId = string

type RunContainerParams struct {
	NetworkId NetworkId
	IP        string
	Hosts     map[string]string // hostname -> IP
	Ports     map[uint16]uint16 // container port -> host port
}

type ContainerEngine interface {
	CreateNetwork(name string) (NetworkId, string, error)
	BuildImage(tag string, buildContext io.Reader) error
	RemoveImage(tag string) error
	RunContainer(name, tag string, params RunContainerParams) (string, error)
	// Stop and remove
	RemoveContainer(id string) error
	RemoveNetwork(id string) error

	Terminate()
}
