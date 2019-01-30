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

//go:generate go-bindata -pkg env -o tpl.bindata.go internal-tpl/...

package env

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"encoding/json"
	"io/ioutil"
	"net/http"
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

const discoveryTplName = "discovery"

// Configured environment
type Env struct {
	id         string
	wsDir      string
	mountDir   string
	ports      ports
	ips        map[string]string // Hostname -> IP
	ed         *def.InputEnv
	ceng       conteng.ContainerEngine
	netId      string
	ipn        *lib.Net
	containers map[string]*tpl.Container // Container ID -> *Container
	// template name -> [container name -> container id]
	contIds                 map[string][]map[string]string
	terminating             bool
	keepAliveChan           chan bool
	builtImages             map[string]struct{}
	discoveryHostname       string
	discoverExternalAddress string
	params                  Params
	tpls                    []*tpl.Tpl
	tplIdx                  map[string]int
	created                 time.Time
	keepalive               time.Duration
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
		ports:         make(ports),
		ed:            params.EnvDef,
		ceng:          params.ContEng,
		keepAliveChan: make(chan bool, 1),
		params:        params,
		builtImages:   map[string]struct{}{},
		ips:           map[string]string{},
		containers:    map[string]*tpl.Container{},
		contIds:       map[string][]map[string]string{},
		tplIdx:        map[string]int{},
		created:       time.Now(),
	}

	defer func() {
		metrics.NumberOfEnvironments.WithLabelValues().Add(-1)

		if r := recover(); r != nil {
			err = errors.Errorf("Error creating env %s: %s", env.id, r)

			_ = env.Terminate()
		}
	}()

	needDiscovery := true
	keepalive := params.DefaultKeepAlive

	if env.params.EnvDef.Options != nil {
		needDiscovery = !env.params.EnvDef.Options.DisableDiscovery

		if params.EnvDef.Options.KeepAlive != 0 {
			keepalive = params.EnvDef.Options.KeepAlive
		}
	}

	env.keepalive = keepalive.ToDuration()

	if err := env.ApplyTemplates(
		env.params.EnvDef.Templates, needDiscovery, false); err != nil {
		_ = env.Terminate()

		return nil, errors.WithStack(err)
	}

	envLog.Infof("New env created: %s", id)

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

			env.RLock()
			selfPorts := env.ports[tplName][tplIdx][cont.Name()]
			ports := env.ports
			ips := env.ips
			env.RUnlock()

			intrp := &interpolator{
				externalAddress: env.params.ExportAddress,
				self: container2interpolate(cont, selfPorts,
					ips[cont.Hostname()]),
				ports:      ports,
				ips:        ips,
				containers: containers,
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
			if !check.Wait(ctx, true) {
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
		self: container2interpolate(cont, ports,
			env.ips[cont.Hostname()]),
		ports:      env.ports,
		ips:        env.ips,
		containers: containers,
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
	intrplFiles, intrplData := cont.ToInterpolate()

	for _, file := range intrplFiles {
		// Get file mode
		info, err := os.Stat(file)

		if err != nil {
			return errors.Wrapf(err, "Error getting file info %s", file)
		}

		data, err := ioutil.ReadFile(file)

		if err != nil {
			return errors.Wrapf(err, "Error reading file %s", file)
		}

		if extra, ok := intrplData[file]; ok {
			i.extra = extra
		} else {
			i.extra = nil
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
func (env *Env) executeTemplates(tpls []*def.Tpl,
	needDiscovery bool) ([]*tpl.Tpl, error) {

	ctx, cancel := context.WithCancel(env.params.Ctx)
	defer cancel()

	tplnum := len(tpls)

	if needDiscovery {
		tplnum += 1
	}

	rch := make(chan *executeResult, tplnum)
	errch := make(chan error, tplnum)

	env.Lock()
	for _, template := range tpls {
		idx := env.tplIdx[template.Tpl]
		env.tplIdx[template.Tpl]++

		go env.execTpl(template, idx, rch, errch, false, ctx)
	}
	env.Unlock()

	if needDiscovery {
		itpl := &def.Tpl{
			Tpl:        discoveryTplName,
			Parameters: def.TplParams{},
		}

		go env.execTpl(itpl, 0, rch, errch, true, ctx)
	}

	var results executeResults

	var execErr error
	count := 0

	for count < tplnum {
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
	for cid := range env.containers {
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
			env.RLock()

			if !env.terminating {
				envLog.Infof("Keep alive timeout triggered for %s, terminating", env.id)
			}

			env.RUnlock()

			_ = env.Terminate()

			return
		}
	}
}

func (env *Env) Export() *def.OutputEnv {
	env.RLock()
	defer env.RUnlock()

	templates := map[string][]*def.TplData{}

	for tplName, tpls := range env.ports {
		for idx, t := range tpls {
			tpld := &def.TplData{
				Containers: map[string]*def.ContainerData{},
			}

			for cont, ps := range t {
				if _, ok := tpld.Containers[cont]; !ok {
					cid := env.contIds[tplName][idx][cont]

					tpld.Containers[cont] = def.NewContainerData(
						cid, env.containers[cid].Hostname())
				}

				for ip, ep := range ps {
					tpld.Containers[cont].Ports[fmt.Sprintf("%d", ip)] = int(ep)
				}
			}

			templates[tplName] = append(templates[tplName], tpld)
		}
	}

	return &def.OutputEnv{
		Id:              env.id,
		Name:            env.params.EnvDef.Name,
		Description:     env.params.EnvDef.Description,
		WsDir:           env.wsDir,
		MountDir:        env.mountDir,
		NetId:           env.netId,
		Created:         env.created.Format(time.RFC3339),
		Keepalive:       env.keepalive.String(),
		ExternalAddress: env.params.ExportAddress,
		Templates:       templates,
	}
}

func (env *Env) execTpl(tplObj *def.Tpl, idx int,
	rch chan *executeResult, errch chan error,
	internal bool, ctx context.Context) {

	params := tpl.ExecuteParams{
		WsDir:     env.wsDir,
		MountDir:  env.mountDir,
		TplParams: tplObj.Parameters,
		Ctx:       ctx,
	}

	if internal {
		params.TplDir = "internal-tpl"
		params.Fs = &tpl.Fs{
			ReadFile: Asset,
			Stat:     AssetInfo,
			Lstat:    AssetInfo,
		}
	} else {
		params.TplDir = env.params.BaseTplDir
	}

	t, err := tpl.Execute(env.id, tplObj.Tpl, idx, params)

	if err != nil {
		errch <- errors.WithStack(err)
	} else {
		rch <- &executeResult{t: t, idx: idx}
	}
}

func (env *Env) StopContainers(strings []string) error {
	env.RLock()
	defer env.RUnlock()

	for _, toRemove := range strings {
		for cid, cont := range env.containers {
			if cid == toRemove {
				envLog.Infof("Stopping container %s", toRemove)

				if err := env.ceng.StopContainer(env.params.Ctx, cid); err != nil {
					return errors.Wrapf(err, "Error stopping container")
				}

				// Wait for readiness checks to fail
				for _, rc := range cont.GetReadinessChecks() {
					if !rc.Wait(env.params.Ctx, false) {
						return errors.Errorf(
							"Error waiting for readiness check: %s", rc.String())
					}
				}

				break
			}
		}
	}

	return nil
}

func (env *Env) RestartContainers(strings []string) error {
	env.RLock()
	defer env.RUnlock()

	for _, toRestart := range strings {
		for cid, cont := range env.containers {
			if cid == toRestart {
				envLog.Infof("Restarting container %s", toRestart)

				if err := env.ceng.RestartContainer(env.params.Ctx, cid); err != nil {
					return errors.Wrapf(err, "Error starting container")
				}

				for _, rc := range cont.GetReadinessChecks() {
					if !rc.Wait(env.params.Ctx, true) {
						return errors.Errorf(
							"Error waiting for readiness check: %s", rc.String())
					}
				}

				break
			}
		}
	}

	return nil
}

func (env *Env) ApplyTemplates(tplDefs []*def.Tpl,
	needDiscovery, updateDiscovery bool) error {

	imagesToBuild := map[string]*tpl.BuildImage{}
	imagesToFetch := map[string]*tpl.FetchImage{}

	var containers []*tpl.Container

	tpls, err := env.executeTemplates(tplDefs, needDiscovery)

	if err != nil {
		return errors.WithStack(err)
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
		env.params.ContEng, env.params.Ctx)

	if err != nil {
		return errors.WithStack(err)
	}

	env.Lock()
	ipn := env.ipn
	netId := env.netId
	var sub string

	if env.ipn == nil {
		// Create network
		netId, sub, err = env.params.ContEng.CreateNetwork(
			env.params.Ctx, env.id)

		if err != nil {
			env.Unlock()

			return errors.Wrapf(err, "Error creating network")
		}

		ipn, err = lib.ParseNet(sub)

		if err != nil {
			env.Unlock()

			return errors.Wrapf(err, "Error parsing subnet")
		}

		// Skip gateway
		_ = ipn.NextIP()

		env.netId = netId
		env.ipn = ipn
	}

	env.Unlock()

	ips, hosts, err := assignIps(ipn, containers)

	if err != nil {
		return errors.WithStack(err)
	}

	env.Lock()
	for k, v := range hosts {
		env.ips[k] = v
	}
	env.Unlock()

	// Expose all the ports
	cports := map[string]map[uint16]uint16{}

	env.RLock()
	discoveryHostname := env.discoveryHostname
	env.RUnlock()

	for _, cont := range containers {
		var dport uint16

		if cont.GetLabel("xenv-discovery") == "true" {
			env.Lock()
			env.discoveryHostname = cont.Hostname()
			discoveryHostname = env.discoveryHostname
			env.Unlock()

			portStr := cont.GetLabel("xenv-discovery-port")
			port, _ := strconv.ParseUint(portStr, 0, 16)
			dport = uint16(port)

		}

		imgName := cont.Image()
		portsToExpose := cont.Ports()

		if len(portsToExpose) == 0 {
			portsToExpose = imgPorts[imgName]
		}

		if _, ok := cports[cont.Hostname()]; !ok {
			cports[cont.Hostname()] = map[uint16]uint16{}
		}

		for _, contPort := range portsToExpose {
			port, err := env.params.PortRange.NextPort()

			if err != nil {
				return errors.WithStack(err)
			}

			envLog.Debugf("Exposing internal port %d as %d for %s",
				contPort, port, cont.Hostname())

			cports[cont.Hostname()][contPort] = port

			if contPort == dport {
				env.Lock()
				env.discoverExternalAddress = fmt.Sprintf(
					"http://%s:%d/api/v1/domains", env.params.ExportAddress, port)
				env.Unlock()
			}
		}

		env.Lock()
		env.ports.add(cont, cports[cont.Hostname()])
		env.Unlock()
	}

	// Collect all the containers for interpolation
	var allContainers []*tpl.Container

	env.RLock()
	for _, cont := range env.containers {
		allContainers = append(allContainers, cont)
	}
	env.RUnlock()

	allContainers = append(allContainers, containers...)

	// Now create containers
	for i, cont := range containers {
		// Interpolate container files
		if err = env.interpolate(cont, cports[cont.Hostname()],
			allContainers); err != nil {

			return errors.WithStack(err)
		}

		cparams := conteng.RunContainerParams{
			NetworkId:  netId,
			IP:         ips[i],
			Ports:      cports[cont.Hostname()],
			Environ:    cont.Environ(),
			Cmd:        cont.Cmd(),
			Entrypoint: cont.Entrypoint(),
			FileMounts: cont.Mounts(),
		}

		if needDiscovery || discoveryHostname != "" {
			cparams.DiscoverDNS = env.ips[discoveryHostname]

			envLog.Infof("[%s] Using discovery DNS: %s",
				env.id, cparams.DiscoverDNS)
		} else {
			cparams.Hosts = hosts

			envLog.Infof("[%s] Using static hosts", env.id)
		}

		cid, err := env.params.ContEng.RunContainer(env.params.Ctx,
			cont.Hostname(), cont.Image(), cparams)

		if err != nil {
			return errors.Wrapf(err, "Error running container: %s",
				cont.Hostname())
		}

		env.Lock()
		env.containers[cid] = cont

		// Allow looking up container id by full name
		tplName, tplIdx := cont.Template()

		if _, ok := env.contIds[tplName]; !ok {
			env.contIds[tplName] = []map[string]string{}
		}

		for len(env.contIds[tplName]) < tplIdx+1 {
			env.contIds[tplName] = append(env.contIds[tplName],
				map[string]string{})
		}

		env.contIds[tplName][tplIdx][cont.Name()] = cid
		env.Unlock()
	}

	// Perform readiness checks
	if err := env.waitUntilReady(containers); err != nil {
		return errors.Wrapf(err, "Error running readiness checks")
	}

	env.Lock()
	env.tpls = append(env.tpls, tpls...)
	env.Unlock()

	if updateDiscovery {
		body := map[string]interface{}{}

		for _, cont := range containers {
			body[fmt.Sprintf("%s.", cont.Hostname())] = hosts[cont.Hostname()]
		}

		bodyBytes, _ := json.Marshal(body)

		cl := http.Client{
			Timeout: 5 * time.Second,
		}

		env.RLock()

		req, _ := http.NewRequest(
			http.MethodPatch, env.discoverExternalAddress,
			bytes.NewReader(bodyBytes))

		env.RUnlock()

		resp, err := cl.Do(req)

		if err != nil {
			return errors.Wrapf(err, "Error updating discovery agent")
		} else {
			_ = resp.Body.Close()
		}

		if resp.StatusCode != http.StatusOK {
			return errors.Errorf(
				"Expected code 200 from discovery agent but got: %d (%s)",
				resp.StatusCode, resp.Status)
		} else {
			envLog.Debugf("[%s]: Discovery agent updated", env.id)
		}
	}

	return nil
}
