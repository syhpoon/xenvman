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

package api

import (
	"fmt"
	"time"

	"encoding/json"

	"github.com/pkg/errors"
	"github.com/syhpoon/xenvman/pkg/lib"
)

type imageDef struct {
	Name       string                 `json:"name"`
	Repo       string                 `json:"repo"`
	Provider   string                 `json:"provider"`
	Parameters map[string]interface{} `json:"parameters"`
}

type containerDef struct {
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Env     map[string]string `json:"env"`
	Volumes map[string]string `json:"volumes"`
	Ports   []uint16          `json:"ports,omitempty"`
}

type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d)
}

func (d Duration) String() string {
	return time.Duration(d).String()
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))

		return nil
	case string:
		tmp, err := time.ParseDuration(value)

		if err != nil {
			return err
		} else {
			*d = Duration(tmp)
		}

		return nil
	default:
		return errors.New("invalid duration")
	}
}

type envDef struct {
	// Environment name
	Name string `json:"name"`
	// Environment description
	Description string `json:"description,omitempty"`
	// Images to provision
	Images []*imageDef `json:"images"`
	// Containers to create
	Containers []*containerDef `json:"containers"`

	// Additional env options
	Options struct {
		KeepAlive Duration `json:"keep_alive,omitempty"`
	} `json:"options"`
}

func (ed *envDef) Validate() error {
	if ed.Name == "" {
		return fmt.Errorf("Name is empty")
	}

	return nil
}

func (ed *envDef) assignIps(sub string,
	conts []*containerDef) (map[string]string, map[string]string, error) {

	ipn, err := lib.ParseNet(sub)

	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// Skip gateway
	_ = ipn.NextIP()

	ips := map[string]string{}
	hosts := map[string]string{}

	for _, cont := range conts {
		ip := ipn.NextIP()

		if ip == nil {
			return nil, nil, errors.Errorf("Unable to assign IP to container: network %s exhausted", sub)
		}

		ips[cont.Name] = ip.String()
		hosts[cont.Name] = ip.String()

		envLog.Debugf("Assigned IP %s to %s", ips[cont.Name], cont.Name)
	}

	return ips, hosts, nil
}
