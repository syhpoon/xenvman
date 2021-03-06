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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/docker/distribution/reference"
	hclient "github.com/docker/docker-credential-helpers/client"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/registry"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/syhpoon/xenvman/pkg/lib"
	"github.com/syhpoon/xenvman/pkg/logger"
)

var dockerLog = logger.GetLogger("xenvman.pkg.conteng.conteng_docker")

type DockerEngineParams struct {
}

type DockerEngine struct {
	cl         *client.Client
	params     DockerEngineParams
	subNetOct1 int
	subNetOct2 int
	subNetMu   sync.Mutex
}

func NewDockerEngine(params DockerEngineParams) (*DockerEngine, error) {
	var opts []func(client2 *client.Client) error

	cli, err := client.NewClientWithOpts(opts...)

	if err != nil {
		return nil, errors.Wrapf(err, "Error creating docker client")
	}

	cli.NegotiateAPIVersion(context.Background())

	dockerLog.Debugf("Docker engine client created")

	return &DockerEngine{
		cl:         cli,
		params:     params,
		subNetOct1: 0,
		subNetOct2: 0,
	}, nil
}

func (de *DockerEngine) CreateNetwork(ctx context.Context,
	name string) (NetworkId, string, error) {
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

	r, err := de.cl.NetworkCreate(ctx, name, netParams)

	if err != nil {
		return "", "", errors.Wrapf(err, "Error creating docker network: %s", sub)
	}

	dockerLog.Debugf("Network created: %s - %s :: %s", name, r.ID, sub)

	return r.ID, sub, nil
}

// Run Docker container
func (de *DockerEngine) RunContainer(ctx context.Context, name, tag string,
	params RunContainerParams) (string, error) {

	// Hosts
	var hosts []string

	for host, ip := range params.Hosts {
		hosts = append(hosts, fmt.Sprintf("%s:%s", host, ip))
	}

	// Ports
	var rawPorts []string

	for contPort, hostPort := range params.Ports {
		rawPorts = append(rawPorts, fmt.Sprintf("%d:%d", hostPort, contPort))
	}

	ports, bindings, err := nat.ParsePortSpecs(rawPorts)

	if err != nil {
		return "", errors.Wrapf(err, "Error parsing ports for %s", name)
	}

	// Environ
	var environ []string

	for k, v := range params.Environ {
		environ = append(environ, fmt.Sprintf("%s=%s", k, v))
	}

	// Mounts
	var mounts []mount.Mount

	for _, fileMount := range params.FileMounts {
		mounts = append(mounts, mount.Mount{
			Type:     "bind",
			Source:   fileMount.HostFile,
			Target:   fileMount.ContainerFile,
			ReadOnly: fileMount.Readonly,
		})
	}

	var dns []string

	if params.DiscoverDNS != "" {
		dns = append(dns, params.DiscoverDNS)
	}

	hostCont := &container.HostConfig{
		NetworkMode:   container.NetworkMode(params.NetworkId),
		ExtraHosts:    hosts,
		AutoRemove:    false,
		DNS:           dns,
		DNSSearch:     []string{"xenv"},
		RestartPolicy: container.RestartPolicy{Name: "on-failure"},
		PortBindings:  bindings,
		Mounts:        mounts,
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

	r, err := de.cl.ContainerCreate(ctx, &container.Config{
		Hostname:     name,
		AttachStdout: true,
		AttachStderr: true,
		Image:        tag,
		ExposedPorts: ports,
		Env:          environ,
		Cmd:          params.Cmd,
		Entrypoint:   params.Entrypoint,
	}, hostCont, netConf, lib.NewIdShort())

	if err != nil {
		return "", errors.Wrapf(err, "Error creating container %s", tag)
	}

	err = de.cl.ContainerStart(ctx, r.ID, types.ContainerStartOptions{})

	if err != nil {
		return "", errors.Wrapf(err, "Error starting container: %s", tag)
	}

	dockerLog.Debugf("Container started: %s, network=%s", tag, params.NetworkId)

	return r.ID, nil
}

func (de *DockerEngine) RemoveContainer(ctx context.Context, id string) error {
	return de.cl.ContainerRemove(ctx, id,
		types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		})
}

func (de *DockerEngine) StopContainer(ctx context.Context, id string) error {
	return de.cl.ContainerKill(ctx, id, "INT")
}

func (de *DockerEngine) RestartContainer(ctx context.Context, id string) error {
	return de.cl.ContainerStart(ctx, id, types.ContainerStartOptions{})
}

func (de *DockerEngine) RemoveNetwork(ctx context.Context, id string) error {
	return de.cl.NetworkRemove(ctx, id)
}

func (de *DockerEngine) BuildImage(ctx context.Context, imgName string,
	buildContext io.Reader) error {

	opts := types.ImageBuildOptions{
		NetworkMode:    "bridge",
		Tags:           []string{imgName},
		Remove:         true,
		ForceRemove:    true,
		SuppressOutput: true,
		NoCache:        true,
		PullParent:     true,
	}

	r, err := de.cl.ImageBuild(ctx, buildContext, opts)

	if r.Body != nil {
		defer r.Body.Close()

		// Check server response
		if rerr := de.isErrorResponse(r.Body); rerr != nil {
			return errors.Errorf("Error from Docker server: %s", rerr)
		}
	}

	if err == nil {
		dockerLog.Debugf("Image built: %s", imgName)
	}

	return err
}

func (de *DockerEngine) RemoveImage(ctx context.Context, imgName string) error {
	opts := types.ImageRemoveOptions{
		Force:         true,
		PruneChildren: true,
	}

	_, err := de.cl.ImageRemove(ctx, imgName, opts)

	if err == nil {
		dockerLog.Debugf("Image removed: %s", imgName)
	}

	return err
}

func (de *DockerEngine) FetchImage(ctx context.Context, imgName string) error {
	out, err := de.cl.ImagePull(ctx, imgName, types.ImagePullOptions{})
	var auth string

	if err != nil {
		// Retry with auth
		auth, err = de.getAuthForImage(imgName)

		if err != nil {
			return err
		}

		out, err = de.cl.ImagePull(ctx, imgName, types.ImagePullOptions{
			RegistryAuth: auth,
		})
	}

	if err == nil {
		dockerLog.Debugf("Image fetched: %s", imgName)
	}

	if out != nil {
		_, _ = io.Copy(ioutil.Discard, out)
	}

	return err
}

func (de *DockerEngine) GetImagePorts(ctx context.Context,
	tag string) ([]uint16, error) {
	r, _, err := de.cl.ImageInspectWithRaw(ctx, tag)

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
		dockerLog.Debugf("Inspecting interface %s", addr.String())

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

// TODO: Pretty naive implementation and will likely not work in all the cases
func (de *DockerEngine) getAuthForImage(imageName string) (string, error) {
	type ConfigFile struct {
		AuthConfigs map[string]types.AuthConfig `json:"auths"`
		CredHelpers map[string]string           `json:"credHelpers"`
	}

	ref, err := reference.ParseNormalizedNamed(imageName)

	if err != nil {
		return "", errors.Wrapf(err, "Error parsing image name %s", imageName)
	}

	repoInfo, err := registry.ParseRepositoryInfo(ref)
	if err != nil {
		return "", errors.Wrapf(err, "Error parsing repository %s", imageName)
	}

	file := filepath.Join(os.Getenv("HOME"), ".docker", "config.json")

	b, err := ioutil.ReadFile(file)

	if err != nil {
		return "", errors.Wrapf(err, "Error reading docker config %s", file)
	}

	conf := &ConfigFile{}

	err = json.Unmarshal(b, conf)

	if err != nil {
		return "", errors.Wrapf(err, "Error parsing docker config %s", file)
	}

	srv := repoInfo.Index.Name
	ac := &types.AuthConfig{}

	var variants []string

	if srv == "docker.io" || srv == "index.docker.io" {
		variants = []string{"https://docker.io", "https://index.docker.io/v1/"}
	} else {
		variants = []string{srv}
	}

	var finalErr error

	for _, host := range variants {
		if credHelper, ok := conf.CredHelpers[host]; ok {
			dockerLog.Infof("Using '%s' credential helper for %s", credHelper, host)

			prog := fmt.Sprintf("docker-credential-%s", credHelper)
			p := hclient.NewShellProgramFunc(prog)

			creds, err := hclient.Get(p, host)

			if err != nil {
				finalErr = errors.Wrapf(err, "Error running %s", prog)

				continue
			}

			ac.Username = creds.Username
			ac.Password = creds.Secret
		} else if authConf, ok := conf.AuthConfigs[host]; ok {
			dockerLog.Infof("Using auth section for %s", host)

			if authConf.Username != "" {
				ac.Username = authConf.Username
			}

			if authConf.Password != "" {
				ac.Password = authConf.Password
			}

			if ac.Username == "" {
				auth, err := base64.StdEncoding.DecodeString(authConf.Auth)

				if err != nil {
					finalErr = errors.Wrap(err, "Error decoding auth entry")

					continue
				}

				split := strings.SplitN(string(auth), ":", 2)

				if len(split) < 2 {
					finalErr = errors.Errorf("Invalid auth entry format: %s", auth)
					continue
				}

				ac.Username = split[0]
				ac.Password = split[1]
			}
		} else {
			continue
		}

		ac.ServerAddress = host
		b, _ = json.Marshal(ac)

		return base64.StdEncoding.EncodeToString(b), nil
	}

	return "", finalErr
}
