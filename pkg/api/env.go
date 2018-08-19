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
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/syhpoon/xenvman/pkg/conteng"
	"github.com/syhpoon/xenvman/pkg/lib"
	"github.com/syhpoon/xenvman/pkg/logger"
	"github.com/syhpoon/xenvman/pkg/repo"
)

var envLog = logger.GetLogger("xenvman.pkg.api.env")

func newEnvId(name string) string {
	id := lib.NewId()
	t := time.Now().Format("20060102150405")

	return fmt.Sprintf("%s-%s-%s", name, t, id[:5])
}

// Configured environment
type Env struct {
	Id string

	ceng        conteng.ContainerEngine
	netId       string
	containers  []string
	terminating bool
	sync.Mutex
}

func NewEnv(ed *envDef, ceng conteng.ContainerEngine,
	repos map[string]repo.Repo) (*Env, error) {

	id := newEnvId(ed.Name)

	pimages := map[string]repo.ProvisionedImage{}
	imagesToBuild := map[string]*repo.BuildImage{}
	name2tag := map[string]string{}

	// First provision images
	for _, imDef := range ed.Images {
		rep, ok := repos[imDef.Repo]

		if !ok {
			return nil, errors.Errorf("Unknown repo: %s", imDef.Repo)
		}

		img, err := rep.Provision(imDef.Provider, imDef.Parameters)

		if err != nil {
			return nil, errors.Wrapf(err, "Error provisioning image %s",
				imDef.Name)
		} else {
			name := imDef.Name

			if name == "" {
				name = imDef.Provider
			}

			tag := fmt.Sprintf("xenv-%s-%s:%s-%s", imDef.Repo, imDef.Provider,
				name, id)

			pimages[tag] = img
			name2tag[imDef.Name] = tag

			switch i := img.(type) {
			case *repo.BuildImage:
				imagesToBuild[tag] = i
			case *repo.FetchImage:
				//TODO
			default:
				return nil, errors.Errorf("Uknown provisioned image type: %T", i)
			}
		}
	}

	// Build images
	for tag, img := range imagesToBuild {
		//TODO: Run in parallel
		if err := ceng.BuildImage(tag, img.BuildContext); err != nil {
			//TODO: Clean up
			return nil, errors.Wrapf(err, "Error building image %s", tag)
		}
	}

	// TODO: Fetch images

	// Create network
	netId, sub, err := ceng.CreateNetwork(id)

	if err != nil {
		return nil, errors.Wrapf(err, "Error creating network")
	}

	// TODO: Assign ports

	ips, hosts, err := ed.assignIps(sub, ed.Containers)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	var cids []string

	// Now create containers
	for _, cont := range ed.Containers {
		// Get corresponding image
		tag, ok := name2tag[cont.Image]

		if !ok {
			return nil, errors.Errorf("Unknown image: %s", cont.Image)
		}

		params := conteng.RunContainerParams{
			NetworkId: netId,
			IP:        ips[cont.Name],
			Hosts:     hosts,
		}

		cid, err := ceng.RunContainer(cont.Name, tag, params)

		if err != nil {
			return nil, errors.Wrapf(err, "Error running container: %s", cont.Name)
		} else {
			cids = append(cids, cid)
		}
	}

	envLog.Infof("New env created: %s", id)

	env := &Env{
		Id:         id,
		ceng:       ceng,
		netId:      netId,
		containers: cids,
	}

	return env, nil
}

func (e *Env) Terminate() error {
	e.Lock()

	if e.terminating {
		e.Unlock()
		return nil
	}

	e.terminating = true
	e.Unlock()

	envLog.Infof("Terminating env %s", e.Id)

	var err error

	// Stop containers
	for _, cid := range e.containers {
		if err = e.ceng.RemoveContainer(cid); err != nil {
			envLog.Errorf("Error terminating env %s: %s", e.Id, err)
		}

		envLog.Debugf("Container %s removed", cid)
	}

	if err != nil {
		return err
	}

	//TODO: Remove images

	// Remove network
	err = e.ceng.RemoveNetwork(e.netId)

	if err == nil {
		envLog.Debugf("Network %s removed", e.netId)
	}

	return err
}
