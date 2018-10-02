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

package env

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/syhpoon/xenvman/pkg/conteng"
	"github.com/syhpoon/xenvman/pkg/def"
	"github.com/syhpoon/xenvman/pkg/lib"
	"github.com/syhpoon/xenvman/pkg/logger"
	"github.com/syhpoon/xenvman/pkg/tpl"
)

var envLog = logger.GetLogger("xenvman.pkg.api.env")

// Configured environment
type Env struct {
	id            string
	ports         ports
	ed            *def.Env
	ceng          conteng.ContainerEngine
	netId         string
	containers    []string
	terminating   bool
	keepAliveChan chan bool
	builtImages   map[string]struct{}
	params        Params
	sync.RWMutex
}

type Params struct {
	EnvDef        *def.Env
	ContEng       conteng.ContainerEngine
	PortRange     *lib.PortRange
	BaseTplDir    string
	BaseWsDir     string
	BaseMountDir  string
	ExportAddress string
	Ctx           context.Context
}

func NewEnv(params Params) (*Env, error) {
	id := newEnvId(params.EnvDef.Name)

	tplIdx := map[string]int{}
	imagesToBuild := map[string]*tpl.BuildImage{}
	imagesToFetch := map[string]*tpl.FetchImage{}
	imgPorts := map[string][]uint16{}
	builtTags := map[string]struct{}{}

	var containers []*tpl.Container

	// Execute templates
	for _, template := range params.EnvDef.Templates {
		idx := tplIdx[template.Tpl]

		// TODO: Run in parallel
		t, err := tpl.Execute(id, template.Tpl, idx,
			tpl.ExecuteParams{
				BaseTplDir:   params.BaseTplDir,
				BaseWsDir:    params.BaseWsDir,
				BaseMountDir: params.BaseMountDir,
				TplParams:    template.Parameters,
			})

		if err != nil {
			return nil, errors.WithStack(err)
		}

		tplIdx[template.Tpl]++

		// Collect build images
		for _, bimg := range t.GetBuildImages() {
			imagesToBuild[bimg.Name()] = bimg

			// Collect containers
			for _, c := range bimg.Containers() {
				containers = append(containers, c)
			}
		}

		// Collect fetch images
		for _, fimg := range t.GetFetchImages() {
			imagesToFetch[fimg.Name()] = fimg

			// Collect containers
			for _, c := range fimg.Containers() {
				containers = append(containers, c)
			}
		}
	}

	// Build images
	for tag, img := range imagesToBuild {
		//TODO: Run in parallel
		bctx, err := img.BuildContext()

		if err != nil {
			return nil, errors.WithStack(err)
		}

		if err := params.ContEng.BuildImage(tag, bctx); err != nil {
			//TODO: Clean up
			return nil, errors.Wrapf(err, "Error building image %s", tag)
		}

		builtTags[tag] = struct{}{}

		ports, err := params.ContEng.GetImagePorts(tag)

		if err != nil {
			envLog.Warningf("Error getting exposed ports for %s: %s", tag, err)
		} else {
			envLog.Debugf("Exposed ports for %s: %v", tag, ports)

			imgPorts[tag] = ports
		}
	}

	// Fetch images
	for tag := range imagesToFetch {
		if err := params.ContEng.FetchImage(tag); err != nil {
			//TODO: Clean up
			return nil, errors.Wrapf(err, "Error fetching image %s", tag)
		}
	}

	// Create network
	netId, sub, err := params.ContEng.CreateNetwork(id)

	if err != nil {
		return nil, errors.Wrapf(err, "Error creating network")
	}

	ips, hosts, err := assignIps(sub, containers)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	var cids []string

	ports := make(ports)

	// Now create containers
	for i, cont := range containers {
		imgName := cont.Image()

		// Expose ports
		portsToExpose := cont.Ports()

		if len(portsToExpose) == 0 {
			portsToExpose = imgPorts[imgName]
		}

		cports := map[uint16]uint16{}

		for _, contPort := range portsToExpose {
			port, err := params.PortRange.NextPort()

			if err != nil {
				return nil, errors.WithStack(err)
			}

			envLog.Debugf("Exposing internal port %d as %d for %s",
				contPort, port, cont.Name())

			cports[contPort] = port
		}

		ports.add(cont, cports)

		cparams := conteng.RunContainerParams{
			NetworkId:  netId,
			IP:         ips[i],
			Hosts:      hosts,
			Ports:      cports,
			Environ:    cont.Environ(),
			Cmd:        cont.Cmd(),
			FileMounts: cont.Mounts(),
		}

		cid, err := params.ContEng.RunContainer(cont.Hostname(), imgName, cparams)

		if err != nil {
			return nil, errors.Wrapf(err, "Error running container: %s", cont.Name())
		}

		cids = append(cids, cid)
	}

	envLog.Infof("New env created: %s", id)

	env := &Env{
		id:            id,
		ports:         ports,
		ed:            params.EnvDef,
		ceng:          params.ContEng,
		netId:         netId,
		containers:    cids,
		keepAliveChan: make(chan bool, 1),
		builtImages:   builtTags,
		params:        params,
	}

	if params.EnvDef.Options.KeepAlive != 0 {
		envLog.Infof("Keep alive for %s = %s", id,
			params.EnvDef.Options.KeepAlive)

		go env.keepAliveWatchdog(params.Ctx)
	}

	return env, nil
}

func (e *Env) IsAlive() bool {
	e.RLock()
	alive := !e.terminating
	e.RUnlock()

	return alive
}

func (e *Env) Terminate() error {
	e.Lock()

	if e.terminating {
		e.Unlock()
		return nil
	}

	e.terminating = true
	e.Unlock()

	envLog.Infof("Terminating env %s", e.id)

	var err error

	// Stop containers
	for _, cid := range e.containers {
		if err = e.ceng.RemoveContainer(cid); err != nil {
			envLog.Errorf("Error terminating env %s: %s", e.id, err)
		}

		envLog.Debugf("Container %s removed", cid)
	}

	if err != nil {
		return err
	}

	// Remove images
	for tag := range e.builtImages {
		if err := e.ceng.RemoveImage(tag); err != nil {
			envLog.Warningf("Error removing image: %+v", err)
		}
	}

	// Remove network
	err = e.ceng.RemoveNetwork(e.netId)

	if err == nil {
		envLog.Debugf("Network %s removed", e.netId)
	}

	return err
}

func (e *Env) KeepAlive() {
	e.keepAliveChan <- true
}

func (e *Env) Id() string {
	return e.id
}

func (e *Env) keepAliveWatchdog(ctx context.Context) {
	d := time.Duration(e.ed.Options.KeepAlive)

	keepAliveTimer := time.NewTimer(d)

	for {
		select {
		case <-ctx.Done():
			return
		case <-e.keepAliveChan:
			if !keepAliveTimer.Stop() {
				<-keepAliveTimer.C
			}

			keepAliveTimer.Reset(d)
		case <-keepAliveTimer.C:
			envLog.Infof("Keep alive timeout triggered for %s, terminating", e.id)

			_ = e.Terminate()

			return
		}
	}
}

func (e *Env) Export() *Exported {
	templates := map[string][]*TplData{}

	for tplName, tpls := range e.ports {
		for _, t := range tpls {
			tpld := &TplData{
				Containers: map[string]*ContainerData{},
			}

			for cont, ps := range t {
				if _, ok := tpld.Containers[cont]; !ok {
					tpld.Containers[cont] = newContainerData()
				}

				for ip, ep := range ps {
					tpld.Containers[cont].Ports[int(ip)] = fmt.Sprintf("%s:%d",
						e.params.ExportAddress, ep)
				}
			}

			templates[tplName] = append(templates[tplName], tpld)
		}
	}

	return &Exported{
		Id:        e.id,
		Templates: templates,
	}
}
