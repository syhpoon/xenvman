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
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/mholt/archiver"

	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
)

type BuildImage struct {
	name    string
	dataDir string
	wsDir   string
	tag     string

	containers map[string]*Container
}

// Copy a file/dir from data dir to workspace dir
func (img *BuildImage) CopyDataToWorkspace(objs ...string) {
	for _, obj := range objs {
		if obj == "*" {
			if err := Copy(img.dataDir, img.wsDir); err != nil {
				panic(errors.Wrapf(err, "Error copying data to workspace"))
			}
		} else {
			dataPath := filepath.Clean(filepath.Join(img.dataDir, obj))
			wsPath := filepath.Clean(filepath.Join(img.wsDir, obj))

			img.verifyPath(dataPath, img.dataDir)
			img.verifyPath(wsPath, img.wsDir)

			if err := Copy(dataPath, wsPath); err != nil {
				panic(errors.Wrapf(err, "Error copying data to workspace"))
			}
		}
	}
}

func (img *BuildImage) AddFileToWorkspace(file string, data interface{}, mode int) {
	path := filepath.Clean(filepath.Join(img.wsDir, file))

	img.verifyPath(path, img.wsDir)

	var bs []byte

	switch d := data.(type) {
	case []byte:
		bs = d
	case string:
		bs = []byte(d)
	default:
		panic(
			errors.Errorf("Invalid file data type %T, expected bytes or string", data))
	}

	if err := ioutil.WriteFile(path, bs, os.FileMode(mode)); err != nil {
		panic(errors.Wrapf(err, "Error copying file %s", file))
	}
}

func (img *BuildImage) NewContainer(name string) *Container {
	cont := &Container{
		name:    name,
		image:   img.tag,
		environ: map[string]string{},
	}

	img.containers[name] = cont

	return cont
}

func (img *BuildImage) Tag() string {
	return img.tag
}

func (img *BuildImage) Containers() map[string]*Container {
	return img.containers
}

func (img *BuildImage) BuildContext() (io.Reader, error) {
	var b bytes.Buffer
	out := bufio.NewWriter(&b)

	var paths []string

	infos, err := ioutil.ReadDir(img.wsDir)

	if err != nil {
		return nil, errors.Wrapf(err, "Error reading workspace: %s", img.wsDir)
	}

	for _, info := range infos {
		paths = append(paths, filepath.Join(img.wsDir, info.Name()))
	}

	err = archiver.TarBz2.Write(out, paths)

	if err != nil {
		return nil, errors.Wrapf(err, "Error creating archive: %s", img.wsDir)
	}

	out.Flush()

	return bytes.NewReader(b.Bytes()), nil
}

func (img *BuildImage) verifyPath(path, base string) {
	if !strings.HasPrefix(path, base) {
		panic(errors.Errorf("Invalid path: %s", path))
	}
}
