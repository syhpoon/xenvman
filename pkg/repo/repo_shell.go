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

package repo

import (
	"fmt"
	"strings"

	"encoding/base64"
	"io/ioutil"
	"os/exec"
	"path/filepath"

	"encoding/json"

	"github.com/pkg/errors"
	"github.com/syhpoon/xenvman/pkg/logger"
)

var shellRepoLog = logger.GetLogger("xenvman.pkg.repo.repo_shell")

// Shell repo uses shell scripts as image providers.
// Every executable file in configured directories will be treated
// as image provider.
// Every script must conform to the following usage pattern:
// <provider> provision --<param>=<value>
//
// The output of the above must be one the following:
// In case of building image:
//
// BUILD\n
// <base64-encoded tar.gz of build directory (context)>
//
// In case of FETCHING image:
//
// FETCH\n
// IMAGE=<fetch repo and image>\n
// TAG=<image tag>\n
//
// In case of error:
//
// ERROR\n
// <error message>

type ShellRepoParams struct {
	// List of dirs where image scripts would be searched
	Dirs []string

	// File name prefix for provider scripts
	ProviderFilePrefix string `mapstructure:"provider-file-prefix"`
}

type ShellRepo struct {
	params    ShellRepoParams
	providers map[string]string
}

func (sr ShellRepo) String() string {
	b, _ := json.MarshalIndent(sr.params, "", "   ")

	return string(b)
}

func NewShellRepo(params ShellRepoParams) (*ShellRepo, error) {
	providers := map[string]string{}

	// Build a list of all available scripts in given directories
	// TODO: Add an ability to watch given dirs and re-configure
	// TODO: available providers on-the-fly

	for _, dir := range params.Dirs {
		files, err := ioutil.ReadDir(dir)

		if err != nil {
			return nil, errors.Wrapf(err, "Error reading dir: %s", dir)
		}

		for _, file := range files {
			name := file.Name()

			if file.Mode()&0111 != 0 && strings.HasPrefix(
				name, params.ProviderFilePrefix) {

				trimmed := strings.TrimPrefix(name, params.ProviderFilePrefix)

				providers[trimmed] = filepath.Join(dir, name)

				shellRepoLog.Debugf("Configured provider: %s for %s",
					trimmed, providers[trimmed])
			}
		}
	}

	return &ShellRepo{
		params:    params,
		providers: providers,
	}, nil
}

func (sr *ShellRepo) Provision(name string,
	params ImageParams) (ProvisionedImage, error) {

	provider, ok := sr.providers[name]

	if !ok {
		return nil, errors.Errorf("Invalid provider: %s", name)
	}

	var buf []string

	for k, v := range params {
		buf = append(buf, fmt.Sprintf("--%s=%v", k, v))
	}

	cmd := fmt.Sprintf("%s provision %s", provider, strings.Join(buf, " "))

	// Execute `provision` command
	outBytes, err := exec.Command(cmd).Output()

	if err != nil {
		return nil, errors.Wrapf(err, "Error executing provider %s", name)
	}

	out := string(outBytes)

	return parseProvision(name, strings.Split(out, "\n"))
}

func parseProvision(script string, lines []string) (ProvisionedImage, error) {
	if len(lines) == 0 {
		return nil, errors.Errorf("%s: Provision script output is empty", script)
	}

	typ := strings.TrimSpace(lines[0])

	switch typ {
	case "ERROR":
		return nil, errors.Errorf("%s: Error provisioning image: %s",
			script, strings.Join(lines[1:], ","))
	case "BUILD":
		return parseBuild(script, lines[1:])
	case "FETCH":
		return nil, errors.New("FETCH not implemented yet")
	default:
		return nil, errors.Errorf("%s: Unknown output type %s", script, typ)
	}
}

// The BUILD output body must consist of a single line
// representing a base64-encoded tar.gz archive with docker build context
func parseBuild(script string, lines []string) (ProvisionedImage, error) {
	if len(lines) != 1 {
		return nil, errors.Errorf("%s: BUILD output body must contain a single "+
			"line representing a base64-encoded tar.gz archive with "+
			"docker build context", script)
	}

	buf, err := base64.StdEncoding.DecodeString(lines[1])

	if err != nil {
		return nil, errors.Wrapf(err, "%s: Error decoding base64", script)
	}

	return &BuildImage{BuildContext: buf}, nil
}
