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

package api

import (
	"fmt"
	"time"

	"github.com/syhpoon/xenvman/pkg/conteng"
	"github.com/syhpoon/xenvman/pkg/lib"
	"github.com/syhpoon/xenvman/pkg/logger"
	"github.com/syhpoon/xenvman/pkg/repo"
)

var envLog = logger.GetLogger("xenvman.pkg.api.env")

type imageDef struct {
	Name       string                 `json:"name"`
	Repo       string                 `json:"repo"`
	Provider   string                 `json:"provider"`
	Parameters map[string]interface{} `json:"parameters"`
}

type containerDef struct {
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Env     map[string]string `json:"env"`
	Volumes map[string]string `json:"volumes"`
}

type envDef struct {
	// Environment name
	Name string `json:"name"`
	// Environment description
	Description string `json:"description"`
	// Images to provision
	Images []*imageDef `json:"images"`
	// Containers to create
	Containers []*containerDef `json:"containers"`
}

func (ed *envDef) Validate() error {
	if ed.Name == "" {
		return fmt.Errorf("Name is empty")
	}

	return nil
}

func newEnvId(name string) string {
	id := lib.NewId()
	t := time.Now().Format("20060102150405")

	return fmt.Sprintf("%s-%s-%s", name, t, id[:5])
}

// Configured environment
type Env struct {
	Id string

	ceng  conteng.ContainerEngine
	netId string
}

func NewEnv(ed *envDef, ceng conteng.ContainerEngine,
	repos map[string]repo.Repo) (*Env, error) {

	id := newEnvId(ed.Name)

	images := map[string]repo.ProvisionedImage{}
	imagesToBuild := map[string]*repo.BuildImage{}

	// First provision images
	for _, imDef := range ed.Images {
		rep, ok := repos[imDef.Repo]

		if !ok {
			envLog.Errorf("Unknown repo: %s", imDef.Repo)

			return nil, fmt.Errorf("Unknown repo: %s", imDef.Repo)
		}

		img, err := rep.Provision(imDef.Provider, imDef.Parameters)

		if err != nil {
			envLog.Errorf("Error provisioning image %+v: %s", imDef, err)

			return nil, fmt.Errorf("Error provisioning image %s: %s", imDef.Name, err)
		} else {
			name := imDef.Name

			if name == "" {
				name = imDef.Provider
			}

			tag := fmt.Sprintf("xenv-%s-%s:%s-%s", imDef.Repo, imDef.Provider,
				name, id)

			images[tag] = img

			switch i := img.(type) {
			case *repo.BuildImage:
				imagesToBuild[tag] = i
			case *repo.FetchImage:
				//TODO
			default:
				envLog.Errorf("Unknown provisioned image type: %T", i)

				return nil, fmt.Errorf("Uknown provisioned image type: %T", i)
			}
		}
	}

	// Build images
	for tag, img := range imagesToBuild {
		//TODO: Run in parallel
		if err := ceng.BuildImage(tag, img.BuildContext); err != nil {
			envLog.Errorf("Error building image %s: %s", tag, err)

			//TODO: Clean up
			return nil, fmt.Errorf("Error building image %s: %s", tag, err)
		}
	}

	// Create network
	netId, err := ceng.CreateNetwork(id)

	if err != nil {
		return nil, err
	}

	// TODO: Assign ports

	/*
		// Now create containers
		for _, cont := range ed.Containers {

		}
	*/

	envLog.Infof("New env created: %s", id)

	env := &Env{
		Id:    id,
		ceng:  ceng,
		netId: netId,
	}

	return env, nil
}
