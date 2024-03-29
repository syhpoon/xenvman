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
	"strings"

	"path/filepath"

	"github.com/pkg/errors"
)

func getTplPaths(tpl, baseTplDir string) (string, string, error) {
	tpl = strings.TrimSpace(tpl)
	tplFile := fmt.Sprintf("%s.tpl.js", tpl)

	if strings.Contains(tplFile, "..") {
		return "", "",
			errors.Errorf("Template name must not contain '..': %s", tpl)
	}

	if strings.HasPrefix(tplFile, "/") {
		tpl = tpl[1:]
	}

	jsFile := filepath.Clean(filepath.Join(baseTplDir, tplFile))

	if !strings.HasPrefix(jsFile, baseTplDir) {
		return "", "", errors.Errorf("Invalid template name: %s", tpl)
	}

	dataDir := filepath.Clean(filepath.Join(baseTplDir,
		fmt.Sprintf("%s.tpl.data", tpl)))

	return jsFile, dataDir, nil
}

func makeDir(path string) {
	if err := os.MkdirAll(path, 0755); err != nil {
		panic(errors.Wrapf(err, "Error creating dir %s", path))
	}
}

func verifyPath(path, base string) {
	if !strings.HasPrefix(path, base) {
		panic(errors.Errorf("Invalid path: %s", path))
	}
}

func checkCancelled(ctx context.Context) {
	if ctx == nil {
		return
	}

	select {
	case <-ctx.Done():
		panic(errors.WithStack(errCancelled))
	default:
		return
	}
}
