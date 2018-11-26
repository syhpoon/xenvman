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

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/syhpoon/xenvman/pkg/conteng"
	"github.com/syhpoon/xenvman/pkg/lib"
	"github.com/syhpoon/xenvman/pkg/logger"
)

var contLog = logger.GetLogger("xenvman.pkg.tpl.container")

type Container struct {
	envId             string
	tplName           string
	tplIdx            int
	name              string
	image             string
	cmd               []string
	ports             []uint16
	dataDir           string
	mountDir          string
	mounts            []*conteng.ContainerFileMount
	environ           map[string]string
	labels            map[string]string
	needInterpolating map[string]bool
	ctx               context.Context
}

func NewContainer(name, tplName string, tplIdx int) *Container {
	return &Container{
		tplName:           tplName,
		tplIdx:            tplIdx,
		name:              name,
		environ:           map[string]string{},
		labels:            map[string]string{},
		needInterpolating: map[string]bool{},
	}
}

func (cont *Container) SetEnv(k, v string) {
	checkCancelled(cont.ctx)
	cont.environ[k] = v
}

func (cont *Container) SetLabel(k string, v interface{}) {
	checkCancelled(cont.ctx)

	val := ""

	switch vv := v.(type) {
	case string:
		val = vv
	default:
		val = fmt.Sprintf("%v", vv)
	}

	cont.labels[k] = val
}

func (cont *Container) SetCmd(cmd ...string) {
	checkCancelled(cont.ctx)
	cont.cmd = cmd
}

func (cont *Container) SetPorts(ports ...uint16) {
	checkCancelled(cont.ctx)
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

func (cont *Container) Labels() map[string]string {
	return cont.labels
}

func (cont *Container) GetLabel(label string) string {
	return cont.labels[label]
}

func (cont *Container) ToInterpolate() []string {
	var r []string

	for k := range cont.needInterpolating {
		r = append(r, k)
	}

	return r
}

func (cont *Container) Cmd() []string {
	return cont.cmd
}

func (cont *Container) Mounts() []*conteng.ContainerFileMount {
	return cont.mounts
}

// Create a file in the mount dir from data and mount it into a container
func (cont *Container) MountString(data, contFile string, mode int, opts Opts) {
	id := lib.NewId()
	path := filepath.Clean(filepath.Join(cont.mountDir, id))
	verifyPath(path, cont.mountDir)

	checkCancelled(cont.ctx)

	if err := ioutil.WriteFile(path, []byte(data), os.FileMode(mode)); err != nil {
		panic(errors.Wrapf(err, "Error copying file %s", path))
	}

	cont.doMount(path, contFile, opts)
}

// Copy a file from data dir to mount dir and mount it into a container
func (cont *Container) MountData(dataFile, contFile string, opts Opts) {
	dataPath := filepath.Clean(filepath.Join(cont.dataDir, dataFile))
	mountPath := filepath.Clean(filepath.Join(cont.mountDir, dataFile))

	verifyPath(dataPath, cont.dataDir)
	verifyPath(mountPath, cont.mountDir)

	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		if _, ok := opts["skip-if-nonexistent"]; ok {
			contLog.Infof(
				"[%s:%s] Cannot mount %s as it doesn't exist. Skipping",
				cont.envId, cont.Hostname(), dataPath)

			return
		} else {
			panic(errors.Errorf("Data file does not exist: %s", dataPath))
		}
	}

	checkCancelled(cont.ctx)

	if err := Copy(dataPath, mountPath); err != nil {
		panic(errors.Wrapf(err, "Error copying data to mount dir"))
	}

	cont.doMount(mountPath, contFile, opts)
}

func (cont *Container) doMount(hostFile, contFile string, opts Opts) {
	contLog.Debugf("[%s] Mounting %s to %s [opts=%+v]",
		cont.envId, hostFile, contFile, opts)

	cont.mounts = append(cont.mounts, &conteng.ContainerFileMount{
		HostFile:      hostFile,
		ContainerFile: contFile,
		Readonly:      opts.GetBool("readonly", true),
	})

	if opts.GetBool("interpolate", false) {
		cont.needInterpolating[hostFile] = true
	}
}

func (cont *Container) Template() (string, int) {
	return cont.tplName, cont.tplIdx
}

func (cont *Container) Hostname() string {
	tplName := strings.Replace(cont.tplName, "/", "-", -1)

	return fmt.Sprintf("%s.%d.%s", cont.name, cont.tplIdx, tplName)
}
