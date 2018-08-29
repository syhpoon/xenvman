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

package conteng

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"sync"

	"encoding/json"
	"io/ioutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/syhpoon/xenvman/pkg/lib"
	"github.com/syhpoon/xenvman/pkg/logger"
)

var dockerLog = logger.GetLogger("xenvman.pkg.conteng.conteng_docker")

type DockerEngineParams struct {
	Ctx context.Context
}

type DockerEngine struct {
	cl         *client.Client
	params     DockerEngineParams
	subNetOct1 int
	subNetOct2 int
	subNetMu   sync.Mutex
}

func NewDockerEngine(params DockerEngineParams) (*DockerEngine, error) {
	cli, err := client.NewClientWithOpts(client.WithVersion("1.37"))

	if err != nil {
		return nil, errors.Wrapf(err, "Error creating docker client")
	}

	dockerLog.Debugf("Docker engine client created")

	return &DockerEngine{
		cl:         cli,
		params:     params,
		subNetOct1: 0,
		subNetOct2: 0,
	}, nil
}

func (de *DockerEngine) CreateNetwork(name string) (NetworkId, string, error) {
	sub, err := de.getSubNet()

	if err != nil {
		return "", "", err
	}

	netParams := types.NetworkCreate{
		CheckDuplicate: true,
		Driver:         "bridge",
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{
				{
					Subnet:  sub,
					IPRange: sub,
				},
			},
		},
	}

	r, err := de.cl.NetworkCreate(de.params.Ctx, name, netParams)

	if err != nil {
		return "", "", errors.Wrapf(err, "Error creating docker network")
	}

	dockerLog.Debugf("Network created: %s - %s :: %s", name, r.ID, sub)

	return r.ID, sub, nil
}

func (de *DockerEngine) RunContainer(name, tag string,
	params RunContainerParams) (string, error) {

	var hosts []string

	for host, ip := range params.Hosts {
		hosts = append(hosts, fmt.Sprintf("%s:%s", host, ip))
	}

	var rawPorts []string

	for contPort, hostPort := range params.Ports {
		rawPorts = append(rawPorts, fmt.Sprintf("%d:%d", hostPort, contPort))
	}

	ports, bindings, err := nat.ParsePortSpecs(rawPorts)

	if err != nil {
		return "", errors.Wrapf(err, "Error parsing ports for %s", name)
	}

	hostCont := &container.HostConfig{
		NetworkMode:  container.NetworkMode(params.NetworkId),
		ExtraHosts:   hosts,
		AutoRemove:   true,
		PortBindings: bindings,
	}

	netConf := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			params.NetworkId: {
				IPAMConfig: &network.EndpointIPAMConfig{
					IPv4Address: params.IP,
				},
			},
		},
	}

	r, err := de.cl.ContainerCreate(de.params.Ctx, &container.Config{
		Hostname:     name,
		AttachStdout: true,
		AttachStderr: true,
		Image:        tag,
		ExposedPorts: ports,
		//TODO:
		Cmd: []string{"/bin/sleep", "infinity"},
	}, hostCont, netConf, lib.NewId()[:5])

	if err != nil {
		return "", errors.Wrapf(err, "Error creating container %s", tag)
	}

	err = de.cl.ContainerStart(de.params.Ctx, r.ID,
		types.ContainerStartOptions{})

	if err != nil {
		return "", errors.Wrapf(err, "Error starting container: %s", tag)
	}

	dockerLog.Debugf("Container started: %s, network=%s", tag, params.NetworkId)

	return r.ID, nil
}

func (de *DockerEngine) RemoveContainer(id string) error {
	return de.cl.ContainerRemove(de.params.Ctx, id,
		types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		})
}

func (de *DockerEngine) RemoveNetwork(id string) error {
	return de.cl.NetworkRemove(de.params.Ctx, id)
}

func (de *DockerEngine) BuildImage(tag string, buildContext io.Reader) error {
	opts := types.ImageBuildOptions{
		NetworkMode:    "bridge",
		Tags:           []string{tag},
		Remove:         true,
		ForceRemove:    true,
		SuppressOutput: true,
		NoCache:        true,
		PullParent:     true,
	}

	r, err := de.cl.ImageBuild(de.params.Ctx, buildContext, opts)
	defer r.Body.Close()

	// Check server response
	if rerr := de.isErrorResponse(r.Body); rerr != nil {
		return errors.Errorf("Error from Docker server: %s", rerr)
	}

	dockerLog.Debugf("Image built: %s", tag)

	return err
}

func (de *DockerEngine) RemoveImage(tag string) error {
	opts := types.ImageRemoveOptions{
		Force:         true,
		PruneChildren: true,
	}

	_, err := de.cl.ImageRemove(de.params.Ctx, tag, opts)

	if err == nil {
		dockerLog.Debugf("Image removed: %s", tag)
	}

	return err
}

func (de *DockerEngine) GetImagePorts(tag string) ([]uint16, error) {
	r, _, err := de.cl.ImageInspectWithRaw(de.params.Ctx, tag)

	if err != nil {
		return nil, errors.Wrapf(err, "Error inspecting image %s", tag)
	}

	var ports []uint16

	for p := range r.Config.ExposedPorts {
		ports = append(ports, uint16(p.Int()))
	}

	return ports, nil
}

func (de *DockerEngine) Terminate() {
	de.cl.Close()
}

func (de *DockerEngine) isErrorResponse(r io.Reader) error {
	data, err := ioutil.ReadAll(r)

	if err != nil {
		return err
	}

	split := bytes.Split(data, []byte("\n"))

	type errResp struct {
		Error string
	}

	for i := range split {
		e := errResp{}

		if err := json.Unmarshal(split[i], &e); err == nil && e.Error != "" {
			return errors.New(e.Error)
		}
	}

	return nil
}

// TODO: This should probably be made more robust at some point
func (de *DockerEngine) getSubNet() (string, error) {
	de.subNetMu.Lock()
	defer de.subNetMu.Unlock()

	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return "", errors.Wrap(err, "Error getting network addresses")
	}

	var nets []*net.IPNet

	for _, addr := range addrs {
		_, n, err := net.ParseCIDR(addr.String())

		if err != nil {
			dockerLog.Warningf("Error parsing address: %s", addr.String())

			continue
		}

		nets = append(nets, n)
	}

	netaddr := func() string {
		tpl := "10.%d.%d.0/24"

		return fmt.Sprintf(tpl, de.subNetOct1, de.subNetOct2)
	}

	_, pnet, _ := net.ParseCIDR(netaddr())

	for {
		// Find non-overlapping network
		overlap := false

		for _, n := range nets {
			if lib.NetsOverlap(pnet, n) {
				overlap = true
				break
			}
		}

		if overlap {
			de.subNetOct2 += 1

			if de.subNetOct2 > 255 {
				de.subNetOct1 += 1
				de.subNetOct2 = 0
			}

			_, pnet, _ = net.ParseCIDR(netaddr())
		} else {
			break
		}
	}

	return netaddr(), nil
}
