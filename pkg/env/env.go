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
	"os"
	"path/filepath"
	"sort"
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
	wsDir         string
	mountDir      string
	ports         ports
	ed            *def.Env
	ceng          conteng.ContainerEngine
	netId         string
	containers    []string
	terminating   bool
	keepAliveChan chan bool
	builtImages   map[string]struct{}
	params        Params
	tpls          []*tpl.Tpl
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
	env := &Env{
		id:            id,
		wsDir:         filepath.Join(params.BaseWsDir, id),
		mountDir:      filepath.Join(params.BaseMountDir, id),
		ed:            params.EnvDef,
		ceng:          params.ContEng,
		keepAliveChan: make(chan bool, 1),
		params:        params,
		builtImages:   map[string]struct{}{},
	}

	imagesToBuild := map[string]*tpl.BuildImage{}
	imagesToFetch := map[string]*tpl.FetchImage{}

	var containers []*tpl.Container

	tpls, err := env.executeTemplates()

	if err != nil {
		return nil, errors.WithStack(err)
	} else {
		env.tpls = tpls
	}

	for _, t := range tpls {
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

	imgPorts, err := buildAndFetch(env, imagesToBuild, imagesToFetch,
		params.ContEng, params.Ctx)

	if err != nil {
		env.Terminate()

		return nil, errors.WithStack(err)
	}

	// Fetch images
	for tag := range imagesToFetch {
		if err := params.ContEng.FetchImage(params.Ctx, tag); err != nil {
			env.Terminate()
			return nil, errors.Wrapf(err, "Error fetching image %s", tag)
		}
	}

	// Create network
	netId, sub, err := params.ContEng.CreateNetwork(params.Ctx, id)

	if err != nil {
		return nil, errors.Wrapf(err, "Error creating network")
	}

	env.netId = netId

	ips, hosts, err := assignIps(sub, containers)

	if err != nil {
		return nil, errors.WithStack(err)
	}

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
				contPort, port, cont.Hostname())

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

		cid, err := params.ContEng.RunContainer(params.Ctx,
			cont.Hostname(), imgName, cparams)

		if err != nil {
			return nil, errors.Wrapf(err, "Error running container: %s", cont.Hostname())
		}

		env.containers = append(env.containers, cid)
	}

	env.ports = ports

	envLog.Infof("New env created: %s", id)

	if params.EnvDef.Options.KeepAlive != 0 {
		envLog.Infof("Keep alive for %s = %s", id,
			params.EnvDef.Options.KeepAlive)

		go env.keepAliveWatchdog(params.Ctx)
	}

	return env, nil
}

func buildAndFetch(env *Env, toBuild map[string]*tpl.BuildImage,
	toFetch map[string]*tpl.FetchImage, ceng conteng.ContainerEngine,
	pctx context.Context) (map[string][]uint16, error) {

	imgPorts := map[string][]uint16{}

	lock := sync.Mutex{}
	ctx, cancel := context.WithCancel(pctx)
	defer cancel()

	total := len(toBuild) + len(toFetch)
	rch := make(chan struct{}, total)
	errch := make(chan error, total)

	// Build images
	for imgName, img := range toBuild {
		go func(imgName string, img *tpl.BuildImage) {
			bctx, err := img.BuildContext()

			if err != nil {
				errch <- errors.WithStack(err)

				return
			}

			if err := ceng.BuildImage(ctx, imgName, bctx); err != nil {
				errch <- errors.Wrapf(err, "Error building image %s", imgName)

				return
			}

			lock.Lock()
			env.builtImages[imgName] = struct{}{}
			lock.Unlock()

			ports, err := ceng.GetImagePorts(ctx, imgName)

			if err != nil {
				envLog.Warningf("Error getting exposed ports for %s: %s", imgName, err)
			} else {
				envLog.Debugf("Exposed ports for %s: %v", imgName, ports)

				lock.Lock()
				imgPorts[imgName] = ports
				lock.Unlock()
			}

			rch <- struct{}{}
		}(imgName, img)
	}

	// Fetch images
	for imgName := range toFetch {
		go func(imgName string) {
			if err := ceng.FetchImage(ctx, imgName); err != nil {
				errch <- errors.Wrapf(err, "Error fetching image %s", imgName)
			}

			rch <- struct{}{}
		}(imgName)
	}

	done := 0

	for done < total {
		select {
		case err := <-errch:
			return nil, errors.WithStack(err)

		case <-rch:
			done++
		}
	}

	return imgPorts, nil
}

type executeResult struct {
	t   *tpl.Tpl
	idx int
}

type executeResults []*executeResult

func (er executeResults) Len() int           { return len(er) }
func (er executeResults) Swap(i, j int)      { er[i], er[j] = er[j], er[i] }
func (er executeResults) Less(i, j int) bool { return er[i].idx < er[j].idx }

// Execute templates in parallel
func (e *Env) executeTemplates() ([]*tpl.Tpl, error) {
	tplIdx := map[string]int{}

	ctx, cancel := context.WithCancel(e.params.Ctx)
	defer cancel()

	tpls := e.params.EnvDef.Templates
	rch := make(chan *executeResult, len(tpls))
	errch := make(chan error, len(tpls))

	for _, template := range tpls {
		idx := tplIdx[template.Tpl]
		tplIdx[template.Tpl]++

		go func(tplObj *def.Tpl, idx int) {
			t, err := tpl.Execute(e.id, tplObj.Tpl, idx,
				tpl.ExecuteParams{
					TplDir:    e.params.BaseTplDir,
					WsDir:     e.wsDir,
					MountDir:  e.mountDir,
					TplParams: tplObj.Parameters,
					Ctx:       ctx,
				})

			if err != nil {
				errch <- errors.WithStack(err)
			} else {

				rch <- &executeResult{t: t, idx: idx}
			}
		}(template, idx)
	}

	var results executeResults

	for len(results) < len(tpls) {
		select {
		case err := <-errch:
			return nil, errors.Wrapf(err, "Error in tpl execution")

		case r := <-rch:
			results = append(results, r)
		}
	}

	sort.Sort(results)

	templates := make([]*tpl.Tpl, len(results))

	for i, r := range results {
		templates[i] = r.t
	}

	return templates, nil
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
		if err = e.ceng.RemoveContainer(e.params.Ctx, cid); err != nil {
			envLog.Errorf("Error terminating env %s: %s", e.id, err)
		}

		envLog.Debugf("Container %s removed", cid)
	}

	if err != nil {
		return err
	}

	// Clean up workspace and mount dirs
	envLog.Debugf("[%s] Removing workspace dir %s", e.id, e.wsDir)

	if err := os.RemoveAll(e.wsDir); err != nil {
		envLog.Errorf("[%s] Error removing workspace dir %s: %s",
			e.id, e.wsDir, err)
	}

	envLog.Debugf("[%s] Removing mount dir %s", e.id, e.mountDir)

	if err := os.RemoveAll(e.mountDir); err != nil {
		envLog.Errorf("[%s] Error removing mount dir %s: %s",
			e.id, e.mountDir, err)
	}

	// Remove images
	for tag := range e.builtImages {
		if err := e.ceng.RemoveImage(e.params.Ctx, tag); err != nil {
			envLog.Warningf("Error removing image: %+v", err)
		}
	}

	// Remove network
	err = e.ceng.RemoveNetwork(e.params.Ctx, e.netId)

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
