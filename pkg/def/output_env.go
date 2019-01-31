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

import (
	"github.com/pkg/errors"
)

type ContainerData struct {
	Id       string `json:"id"`
	Hostname string `json:"hostname"`
	// <internal port> -> <external port>
	Ports map[string]int `json:"ports"`
}

func NewContainerData(id, hostname string) *ContainerData {
	return &ContainerData{
		Id:       id,
		Hostname: hostname,
		Ports:    map[string]int{},
	}
}

type TplData struct {
	// Container name -> <internal port> -> "<host>:<external-port>"
	Containers map[string]*ContainerData `json:"containers"`
}

type OutputEnv struct {
	Id              string                `json:"id"`
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	WsDir           string                `json:"ws_dir" mapstructure:"ws_dir"`
	MountDir        string                `json:"mount_dir" mapstructure:"mount_dir"`
	NetId           string                `json:"net_id" mapstructure:"net_id"`
	Created         string                `json:"created"`
	Keepalive       string                `json:"keep_alive" mapstructure:"keep_alive"`
	ExternalAddress string                `json:"external_address" mapstructure:"external_address"`
	Templates       map[string][]*TplData `json:"templates"` // tpl name -> [TplData]
}

func (e *OutputEnv) GetContainer(tplName string, tplIdx int,
	contName string) (*ContainerData, error) {
	tpls, ok := e.Templates[tplName]

	if !ok {
		return nil, errors.Errorf("Template not found: %s", tplName)
	}

	if tplIdx >= len(tpls) {
		return nil, errors.Errorf("Template %s index %d not found",
			tplName, tplIdx)
	}

	cont, ok := tpls[tplIdx].Containers[contName]

	if !ok {
		return nil, errors.Errorf("Container not found: %s", contName)
	}

	return cont, nil
}
