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
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/syhpoon/xenvman/pkg/lib"
	"github.com/syhpoon/xenvman/pkg/logger"
)

var tplLog = logger.GetLogger("xenvman.pkg.tpl.tpl")

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

	sync.RWMutex
}

func (tpl *Tpl) BuildImage(name string) *BuildImage {
	// xenv-<tpl name>-<image-name>:<env id>-<idx>
	imgName := fmt.Sprintf("xenv-%s-%s:%s-%d",
		tpl.name, name, tpl.envId, tpl.idx)

	wsDir := filepath.Join(tpl.wsDir, imgName)

	if err := os.MkdirAll(wsDir, 0755); err != nil {
		panic(fmt.Sprintf("%+v", err))
	}

	img := &BuildImage{
		Image: newImage(tpl.envId, imgName, wsDir, tpl.mountDir, tpl.dataDir),
	}

	tplLog.Debugf("[%s] Building image %s", tpl.envId, imgName)

	tpl.Lock()
	tpl.buildImages = append(tpl.buildImages, img)
	tpl.Unlock()

	return img
}

func (tpl *Tpl) FetchImage(imgName string) *FetchImage {
	wsDir := filepath.Join(tpl.wsDir, fmt.Sprintf("%s-%s", imgName, lib.NewIdShort()))

	if err := os.MkdirAll(wsDir, 0755); err != nil {
		panic(fmt.Sprintf("%+v", err))
	}

	img := &FetchImage{
		Image: newImage(tpl.envId, imgName, wsDir, tpl.mountDir, tpl.dataDir),
	}

	tplLog.Debugf("[%s] Fetching image %s", tpl.envId, imgName)

	tpl.Lock()
	tpl.fetchImages = append(tpl.fetchImages, img)
	tpl.Unlock()

	return img
}

func (tpl *Tpl) GetBuildImages() []*BuildImage {
	return tpl.buildImages
}

func (tpl *Tpl) GetFetchImages() []*FetchImage {
	return tpl.fetchImages
}
