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
	envId    string
	name     string
	image    string
	environ  map[string]string
	cmd      []string
	ports    []uint16
	dataDir  string
	mountDir string
	mounts   []*conteng.ContainerFileMount
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

func (cont *Container) Mounts() []*conteng.ContainerFileMount {
	return cont.mounts
}

// Create a file in the mount dir from data and mount it inside container
func (cont *Container) MountString(data, contFile string, mode int, readonly bool) {
	id := lib.NewId()
	path := filepath.Clean(filepath.Join(cont.mountDir, id))
	verifyPath(path, cont.mountDir)

	if err := ioutil.WriteFile(path, []byte(data), os.FileMode(mode)); err != nil {
		panic(errors.Wrapf(err, "Error copying file %s", path))
	}

	cont.MountFile(path, contFile, mode, readonly)
}

// Mount a file from the mount dir
// It must be copied there from the data dir (using CopyDataToMountDir) first
func (cont *Container) MountFile(hostFile, contFile string, mode int, readonly bool) {
	if !strings.HasPrefix(hostFile, cont.mountDir) {
		hostFile = filepath.Clean(filepath.Join(cont.mountDir, hostFile))
		verifyPath(hostFile, cont.mountDir)
	}

	contLog.Debugf("[%s] Mounting %s to %s", cont.envId, hostFile, contFile)

	cont.mounts = append(cont.mounts, &conteng.ContainerFileMount{
		HostFile:      hostFile,
		ContainerFile: contFile,
		Readonly:      readonly,
	})
}

func (cont *Container) CopyDataToMountDir(objs ...string) {
	for _, obj := range objs {
		dataPath := filepath.Clean(filepath.Join(cont.dataDir, obj))
		mountPath := filepath.Clean(filepath.Join(cont.mountDir, obj))

		verifyPath(dataPath, cont.dataDir)
		verifyPath(mountPath, cont.mountDir)

		if err := Copy(dataPath, mountPath); err != nil {
			panic(errors.Wrapf(err, "Error copying data to workspace"))
		}
	}
}

func (cont *Container) ExposePorts(ports ...uint16) {
	cont.ports = append(cont.ports, ports...)
}
