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
	"os"
	"sync"

	"path/filepath"

	"github.com/pkg/errors"
	"github.com/syhpoon/xenvman/pkg/lib"
	"github.com/syhpoon/xenvman/pkg/logger"
)

var tplLog = logger.GetLogger("xenvman.pkg.tpl.tpl")

var readinessMap = map[string]func(map[string]interface{}) ReadinessCheck{
	"http": readinessCheckHttpInit,
	"net":  readinessCheckNetInit,
}

type Tpl struct {
	envId string
	name  string
	idx   int

	// Images to build
	buildImages []*BuildImage

	// Images to fetch
	fetchImages []*FetchImage

	dataDir  string
	wsDir    string
	mountDir string
	fs       *Fs

	imported []*Tpl

	ctx context.Context
	sync.RWMutex
}

func (tpl *Tpl) BuildImage(name string) *BuildImage {
	// xenv-<tpl name>-<image-name>:<env id>-<idx>
	imgName := fmt.Sprintf("xenv-%s-%s:%s-%d",
		tpl.name, name, tpl.envId, tpl.idx)

	wsDir := filepath.Clean(filepath.Join(tpl.wsDir, imgName))
	verifyPath(wsDir, tpl.wsDir)

	checkCancelled(tpl.ctx)

	if err := os.MkdirAll(wsDir, 0755); err != nil {
		panic(errors.WithStack(err))
	}

	img := &BuildImage{
		Image: &Image{
			envId:      tpl.envId,
			tplName:    tpl.name,
			tplIdx:     tpl.idx,
			name:       imgName,
			wsDir:      wsDir,
			mountDir:   tpl.mountDir,
			dataDir:    tpl.dataDir,
			containers: map[string]*Container{},
			fs:         tpl.fs,
			ctx:        tpl.ctx,
		},
	}

	tplLog.Debugf("[%s] Building image %s", tpl.envId, imgName)

	tpl.Lock()
	tpl.buildImages = append(tpl.buildImages, img)
	tpl.Unlock()

	return img
}

func (tpl *Tpl) FetchImage(imgName string) *FetchImage {
	wsDir := filepath.Join(tpl.wsDir,
		fmt.Sprintf("%s-%s", imgName, lib.NewIdShort()))
	wsDir = filepath.Clean(wsDir)

	verifyPath(wsDir, tpl.wsDir)

	checkCancelled(tpl.ctx)

	if err := os.MkdirAll(wsDir, 0755); err != nil {
		panic(errors.WithStack(err))
	}

	img := &FetchImage{
		Image: &Image{
			envId:      tpl.envId,
			tplName:    tpl.name,
			tplIdx:     tpl.idx,
			name:       imgName,
			wsDir:      wsDir,
			mountDir:   tpl.mountDir,
			dataDir:    tpl.dataDir,
			containers: map[string]*Container{},
			fs:         tpl.fs,
			ctx:        tpl.ctx,
		},
	}

	tplLog.Debugf("[%s] Fetching image %s", tpl.envId, imgName)

	tpl.Lock()
	tpl.fetchImages = append(tpl.fetchImages, img)
	tpl.Unlock()

	return img
}

func (tpl *Tpl) GetName() string {
	return tpl.name
}

func (tpl *Tpl) GetIdx() int {
	return tpl.idx
}

func (tpl *Tpl) GetBuildImages() []*BuildImage {
	return tpl.buildImages
}

func (tpl *Tpl) GetFetchImages() []*FetchImage {
	return tpl.fetchImages
}

func (tpl *Tpl) SetImported(imported []*Tpl) {
	tpl.imported = imported
}

func (tpl *Tpl) GetImported() []*Tpl {
	return tpl.imported
}
