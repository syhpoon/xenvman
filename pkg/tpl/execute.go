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
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"
	"github.com/syhpoon/xenvman/pkg/def"
)

const executeFunctionName = "execute"

type ExecuteParams struct {
	BaseTplDir string
	BaseWsDir  string
	TplParams  def.TplParams
}

func Execute(envId, tplName string, tplIndex int, params ExecuteParams) (tpl *Tpl, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("Error running template: %v", r)
		}
	}()

	vm := otto.New()

	// Setup library
	setupLib(vm)

	jsFile, dataDir, err := getTplPaths(tplName, params.BaseTplDir)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	tplLog.Debugf("Executing tpl from %s", tplName)

	bytes, err := ioutil.ReadFile(jsFile)

	if err != nil {
		return nil, errors.Wrapf(err, "Error reading tpl %s", tplName)
	}

	_, err = vm.Run(bytes)

	if err != nil {
		return nil, errors.Wrapf(err, "Error executing tpl %s", tplName)
	}

	// /<base-dir>/<env-id>/<tpl-name>/<tpl-idx>
	wsDir := filepath.Join(params.BaseWsDir, envId, tplName,
		fmt.Sprintf("%d", tplIndex))

	tpl = &Tpl{
		envId:   envId,
		name:    tplName,
		idx:     tplIndex,
		dataDir: dataDir,
		wsDir:   wsDir,
	}

	_, err = vm.Call(executeFunctionName, nil, tpl, params.TplParams)

	if err != nil {
		return nil, errors.Wrapf(err, "Error calling %s function for %s",
			executeFunctionName, tpl.name)
	}

	return tpl, nil
}
