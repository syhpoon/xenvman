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
	"path/filepath"

	"github.com/syhpoon/xenvman/pkg/conteng"
	"github.com/syhpoon/xenvman/pkg/lib"
	"github.com/syhpoon/xenvman/pkg/logger"
)

var imgLog = logger.GetLogger("xenvman.pkg.tpl.image")

type Image struct {
	envId      string
	name       string
	dataDir    string
	wsDir      string
	mountDir   string
	containers map[string]*Container
	mounts     []*conteng.ContainerFileMount
}

func newImage(envId, name, wsDir, mountDir, dataDir string) *Image {
	return &Image{
		envId:      envId,
		name:       name,
		wsDir:      wsDir,
		mountDir:   mountDir,
		dataDir:    dataDir,
		containers: map[string]*Container{},
	}
}

func (img *Image) NewContainer(name string) *Container {

	mountDir := filepath.Join(img.mountDir, name, lib.NewIdShort())
	makeDir(mountDir)

	cont := &Container{
		envId:    img.envId,
		name:     name,
		image:    img.name,
		environ:  map[string]string{},
		mountDir: mountDir,
		dataDir:  img.dataDir,
	}

	imgLog.Debugf("[%s] Creating new container for %s, %s",
		img.envId, name, img.name)

	img.containers[name] = cont

	return cont
}

func (img *Image) Name() string {
	return img.name
}

func (img *Image) Containers() map[string]*Container {
	return img.containers
}
