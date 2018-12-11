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
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/mholt/archiver"
	"github.com/pkg/errors"
)

type BuildImage struct {
	*Image
}

// Copy a file/dir from data dir to workspace dir
func (img *Image) CopyDataToWorkspace(objs ...string) {
	checkCancelled(img.ctx)

	for _, obj := range objs {
		if obj == "*" {
			imgLog.Debugf("Copying everything from %s to %s for %s",
				img.dataDir, img.wsDir, img.envId)

			if err := Copy(img.dataDir, img.wsDir); err != nil {
				panic(errors.Wrapf(err, "Error copying data to workspace"))
			}

			return
		} else {
			dataPath := filepath.Clean(filepath.Join(img.dataDir, obj))
			wsPath := filepath.Clean(filepath.Join(img.wsDir, obj))

			verifyPath(dataPath, img.dataDir)
			verifyPath(wsPath, img.wsDir)

			imgLog.Debugf("Copying %s to %s for %s", dataPath, wsPath, img.envId)

			if err := Copy(dataPath, wsPath); err != nil {
				panic(errors.Wrapf(err, "Error copying data to workspace"))
			}
		}
	}
}

func (img *BuildImage) AddFileToWorkspace(file string, data interface{}, mode int) {
	path := filepath.Clean(filepath.Join(img.wsDir, file))
	verifyPath(path, img.wsDir)

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

func (img *BuildImage) InterpolateWorkspaceFile(file string, data interface{}) {
	path := filepath.Clean(filepath.Join(img.wsDir, file))
	verifyPath(path, img.wsDir)

	info, err := os.Stat(path)

	if err != nil {
		panic(errors.Wrapf(err, "Error getting file info %s", file))
	}

	raw, err := ioutil.ReadFile(path)

	if err != nil {
		panic(errors.Wrapf(err, "Error reading file %s", file))
	}

	t, err := template.New("").Parse(string(raw))

	if err != nil {
		panic(errors.Wrapf(err, "Error parsing template from %s", file))
	}

	var b bytes.Buffer
	out := bufio.NewWriter(&b)

	if err = t.Execute(out, data); err != nil {
		panic(errors.Wrapf(err, "Error executing template from %s", file))
	}

	_ = out.Flush()

	if err := ioutil.WriteFile(path, b.Bytes(), info.Mode()); err != nil {
		panic(errors.Wrapf(err, "Error saving template %s", path))
	}
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

	_ = out.Flush()

	return bytes.NewReader(b.Bytes()), nil
}
