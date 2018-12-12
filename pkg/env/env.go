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
	"sort"
	"sync"
	"time"

	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/syhpoon/xenvman/pkg/conteng"
	"github.com/syhpoon/xenvman/pkg/def"
	"github.com/syhpoon/xenvman/pkg/lib"
	"github.com/syhpoon/xenvman/pkg/logger"
	"github.com/syhpoon/xenvman/pkg/metrics"
	"github.com/syhpoon/xenvman/pkg/tpl"
)

var envLog = logger.GetLogger("xenvman.pkg.env.env")

// Configured environment
type Env struct {
	id            string
	wsDir         string
	mountDir      string
	ports         ports
	ed            *def.InputEnv
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
	EnvDef           *def.InputEnv
	ContEng          conteng.ContainerEngine
	PortRange        *lib.PortRange
	BaseTplDir       string
	BaseWsDir        string
	BaseMountDir     string
	ExportAddress    string
	DefaultKeepAlive def.Duration
	Ctx              context.Context
}

func NewEnv(params Params) (env *Env, err error) {
	metrics.NumberOfEnvironments.WithLabelValues().Add(1)

	id := newEnvId(params.EnvDef.Name)
	env = &Env{
		id:            id,
		wsDir:         filepath.Join(params.BaseWsDir, id),
		mountDir:      filepath.Join(params.BaseMountDir, id),
		ed:            params.EnvDef,
		ceng:          params.ContEng,
		keepAliveChan: make(chan bool, 1),
		params:        params,
		builtImages:   map[string]struct{}{},
	}

	defer func() {
		metrics.NumberOfEnvironments.WithLabelValues().Add(-1)

		if r := recover(); r != nil {
			err = errors.Errorf("Error creating env %s: %s", env.id, r)

			_ = env.Terminate()
		}
	}()

	imagesToBuild := map[string]*tpl.BuildImage{}
	imagesToFetch := map[string]*tpl.FetchImage{}

	var containers []*tpl.Container

	tpls, err := env.executeTemplates()

	if err != nil {
		_ = env.Terminate()
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

	imgPorts, err := env.buildAndFetch(imagesToBuild, imagesToFetch,
		params.ContEng, params.Ctx)

	if err != nil {
		_ = env.Terminate()

		return nil, errors.WithStack(err)
	}

	// Create network
	netId, sub, err := params.ContEng.CreateNetwork(params.Ctx, id)

	if err != nil {
		_ = env.Terminate()

		return nil, errors.Wrapf(err, "Error creating network")
	}

	env.netId = netId

	ips, hosts, err := assignIps(sub, containers)

	if err != nil {
		_ = env.Terminate()

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
				_ = env.Terminate()

				return nil, errors.WithStack(err)
			}

			envLog.Debugf("Exposing internal port %d as %d for %s",
				contPort, port, cont.Hostname())

			cports[contPort] = port
		}

		// Interpolate container files
		if err = env.interpolate(cont, cports, containers); err != nil {
			_ = env.Terminate()

			return nil, errors.WithStack(err)
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
			_ = env.Terminate()

			return nil, errors.Wrapf(err, "Error running container: %s", cont.Hostname())
		}

		env.containers = append(env.containers, cid)
	}

	env.ports = ports

	// Perform readiness checks
	if err := env.waitUntilReady(containers); err != nil {
		_ = env.Terminate()

		return nil, errors.Wrapf(err, "Error running readiness checks")
	}

	envLog.Infof("New env created: %s", id)
	keepalive := params.DefaultKeepAlive

	if params.EnvDef.Options != nil && params.EnvDef.Options.KeepAlive != 0 {
		keepalive = params.EnvDef.Options.KeepAlive
	}

	if keepalive != 0 {
		envLog.Infof("Keep alive for %s = %s", id, keepalive)
		go env.keepAliveWatchdog(keepalive, params.Ctx)
	}

	return env, nil
}

func (env *Env) waitUntilReady(containers []*tpl.Container) error {
	var checks []tpl.ReadinessCheck

	for _, cont := range containers {
		for _, check := range cont.GetReadinessChecks() {
			tplName, tplIdx := cont.Template()

			intrp := &interpolator{
				externalAddress: env.params.ExportAddress,
				self:            cont,
				selfPorts:       env.ports[tplName][tplIdx][cont.Name()],
				containers:      containers,
			}

			if err := check.InterpolateParameters(intrp); err != nil {
				return errors.Wrapf(err,
					"Error interpolating readiness check parameters: %s",
					check.String())
			}

			checks = append(checks, check)
		}
	}

	if len(checks) == 0 {
		envLog.Infof("No readiness checks for %s", env.id)

		return nil
	}

	ctx, cancel := context.WithCancel(env.params.Ctx)
	defer cancel()

	total := len(checks)
	errCh := make(chan error, total)
	rch := make(chan struct{}, total)

	for _, check := range checks {
		go func(check tpl.ReadinessCheck) {
			if !check.WaitUntilReady(ctx) {
				errCh <- errors.WithStack(
					errors.Errorf("Readiness check %s failed", check.String()))
			} else {
				envLog.Infof("Readiness check passed %s for %s", check, env.id)

				rch <- struct{}{}
			}
		}(check)
	}

	done := 0

	for done < total {
		select {
		case err := <-errCh:
			return errors.WithStack(err)

		case <-rch:
			done++
		}
	}

	envLog.Infof("Env %s is ready", env.id)

	return nil
}

// Interpolate container:
// * mount files
// * environment variables
func (env *Env) interpolate(cont *tpl.Container, ports map[uint16]uint16,
	containers []*tpl.Container) error {

	cont.SetCtx(nil)

	i := &interpolator{
		externalAddress: env.params.ExportAddress,
		self:            cont,
		selfPorts:       ports,
		containers:      containers,
	}

	// Environ
	for k, val := range cont.Environ() {
		newVal, err := lib.Interpolate(string(val), i)

		if err != nil {
			envLog.Warningf("Error interpolating environment variable %s for %s: %s",
				k, cont.Hostname(), err)
		} else {
			cont.SetEnv(k, newVal)
		}
	}

	// Files
	for _, file := range cont.ToInterpolate() {
		// Get file mode
		info, err := os.Stat(file)

		if err != nil {
			return errors.Wrapf(err, "Error getting file info %s", file)
		}

		data, err := ioutil.ReadFile(file)

		if err != nil {
			return errors.Wrapf(err, "Error reading file %s", file)
		}

		res, err := lib.Interpolate(string(data), i)

		if err != nil {
			return errors.WithStack(err)
		}

		err = ioutil.WriteFile(file, []byte(res), info.Mode())

		if err != nil {
			return errors.Wrapf(err, "Error saving file %s", file)
		}

		envLog.Debugf("Interpolated: %s", file)
	}

	return nil
}

func (env *Env) buildAndFetch(toBuild map[string]*tpl.BuildImage,
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
func (env *Env) executeTemplates() ([]*tpl.Tpl, error) {
	tplIdx := map[string]int{}

	ctx, cancel := context.WithCancel(env.params.Ctx)
	defer cancel()

	tpls := env.params.EnvDef.Templates
	rch := make(chan *executeResult, len(tpls))
	errch := make(chan error, len(tpls))

	for _, template := range tpls {
		idx := tplIdx[template.Tpl]
		tplIdx[template.Tpl]++

		go func(tplObj *def.Tpl, idx int) {
			t, err := tpl.Execute(env.id, tplObj.Tpl, idx,
				tpl.ExecuteParams{
					TplDir:    env.params.BaseTplDir,
					WsDir:     env.wsDir,
					MountDir:  env.mountDir,
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

	var execErr error
	count := 0

	for count < len(tpls) {
		select {
		case err := <-errch:
			if execErr == nil {
				cancel()
				execErr = errors.Wrapf(err, "Error in tpl execution")
			}

		case r := <-rch:
			results = append(results, r)
		}

		count++
	}

	if execErr != nil {
		return nil, execErr
	}

	sort.Sort(results)

	templates := make([]*tpl.Tpl, len(results))

	for i, r := range results {
		templates[i] = r.t
	}

	return templates, nil
}

func (env *Env) IsAlive() bool {
	env.RLock()
	alive := !env.terminating
	env.RUnlock()

	return alive
}

func (env *Env) Terminate() error {
	env.Lock()

	if env.terminating {
		env.Unlock()
		return nil
	}

	env.terminating = true
	env.Unlock()

	envLog.Infof("Terminating env %s", env.id)

	var err error

	// Stop containers
	for _, cid := range env.containers {
		if err = env.ceng.RemoveContainer(env.params.Ctx, cid); err != nil {
			envLog.Errorf("Error terminating env %s: %s", env.id, err)
		}

		envLog.Debugf("Container %s removed", cid)
	}

	if err != nil {
		return err
	}

	// Clean up workspace and mount dirs
	envLog.Debugf("[%s] Removing workspace dir %s", env.id, env.wsDir)

	if err := os.RemoveAll(env.wsDir); err != nil {
		envLog.Errorf("[%s] Error removing workspace dir %s: %s",
			env.id, env.wsDir, err)
	}

	envLog.Debugf("[%s] Removing mount dir %s", env.id, env.mountDir)

	if err := os.RemoveAll(env.mountDir); err != nil {
		envLog.Errorf("[%s] Error removing mount dir %s: %s",
			env.id, env.mountDir, err)
	}

	// Remove images
	for tag := range env.builtImages {
		if err := env.ceng.RemoveImage(env.params.Ctx, tag); err != nil {
			envLog.Warningf("Error removing image: %+v", err)
		}
	}

	// Remove network
	err = env.ceng.RemoveNetwork(env.params.Ctx, env.netId)

	if err == nil {
		envLog.Debugf("Network %s removed", env.netId)
	}

	return err
}

func (env *Env) KeepAlive() {
	env.keepAliveChan <- true
}

func (env *Env) Id() string {
	return env.id
}

func (env *Env) keepAliveWatchdog(dur def.Duration, ctx context.Context) {
	d := time.Duration(dur)

	keepAliveTimer := time.NewTimer(d)

	for {
		select {
		case <-ctx.Done():
			return
		case <-env.keepAliveChan:
			if !keepAliveTimer.Stop() {
				<-keepAliveTimer.C
			}

			keepAliveTimer.Reset(d)
		case <-keepAliveTimer.C:
			envLog.Infof("Keep alive timeout triggered for %s, terminating", env.id)

			_ = env.Terminate()

			return
		}
	}
}

func (env *Env) Export() *def.OutputEnv {
	templates := map[string][]*def.TplData{}

	for tplName, tpls := range env.ports {
		for _, t := range tpls {
			tpld := &def.TplData{
				Containers: map[string]*def.ContainerData{},
			}

			for cont, ps := range t {
				if _, ok := tpld.Containers[cont]; !ok {
					tpld.Containers[cont] = def.NewContainerData()
				}

				for ip, ep := range ps {
					tpld.Containers[cont].Ports[int(ip)] = fmt.Sprintf("%s:%d",
						env.params.ExportAddress, ep)
				}
			}

			templates[tplName] = append(templates[tplName], tpld)
		}
	}

	return &def.OutputEnv{
		Id:        env.id,
		Templates: templates,
	}
}
