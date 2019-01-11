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

	"github.com/syhpoon/xenvman/pkg/def"

	"github.com/mitchellh/mapstructure"
	"github.com/robertkrimen/otto"

	"github.com/pkg/errors"
)

const infoFunctionName = "info"

func LoadTemplatesInfo(baseDir string) (map[string]*def.TplInfo, error) {
	var tpls []string

	f := func(path string, info os.FileInfo, err error) error {
		name := info.Name()

		if info.Mode().IsRegular() && strings.HasSuffix(name, "tpl.js") {
			tpls = append(tpls, path)
		}

		return nil
	}

	if err := filepath.Walk(baseDir, f); err != nil {
		return nil, errors.Wrapf(err, "Error scanning tpl directory")
	}

	vm := otto.New()
	res := map[string]*def.TplInfo{}

	for _, tpl := range tpls {
		bytes, err := ioutil.ReadFile(tpl)

		if err != nil {
			return nil, errors.Wrapf(err, "Error reading tpl %s", tpl)
		}

		_, err = vm.Run(bytes)

		if err != nil {
			return nil, errors.Wrap(err, "Error executing tpl")
		}

		rawInfo, err := vm.Call(infoFunctionName, nil)

		if err != nil {
			return nil, errors.Wrap(err, "Error calling info function")
		}

		rawInfo2, err := rawInfo.Export()

		if err != nil {
			return nil, errors.Wrap(err, "Error exporting info map")
		}

		infoMap, ok := rawInfo2.(map[string]interface{})

		if !ok {
			return nil, errors.Errorf("Expected info map to be object but got: %T", rawInfo2)
		}

		info := def.TplInfo{}

		if err := mapstructure.Decode(infoMap, &info); err != nil {
			return nil, errors.Wrap(err, "Error decoding info map")
		}

		res[tpl] = &info
	}

	return res, nil
}
