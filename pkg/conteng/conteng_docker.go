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
	"io"

	"encoding/json"
	"io/ioutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

type DockerEngineParams struct {
	Ctx context.Context
}

type DockerEngine struct {
	cl     *client.Client
	params DockerEngineParams
}

func NewDockerEngine(params DockerEngineParams) (*DockerEngine, error) {
	cli, err := client.NewClientWithOpts(client.WithVersion("1.37"))

	if err != nil {
		return nil, errors.Wrapf(err, "Error creating docker client")
	}

	return &DockerEngine{
		cl:     cli,
		params: params,
	}, nil
}

func (de *DockerEngine) CreateNetwork(name string) (NetworkId, error) {
	netParams := types.NetworkCreate{
		CheckDuplicate: true,
		Driver:         "bridge",
	}

	r, err := de.cl.NetworkCreate(de.params.Ctx, name, netParams)

	if err != nil {
		return "", errors.Wrapf(err, "Error creating docker network")
	}

	return r.ID, nil
}

func (de *DockerEngine) RunContainer(tag, netId string) error {
	hostCont := &container.HostConfig{
		NetworkMode: container.NetworkMode(netId),
	}

	r, err := de.cl.ContainerCreate(de.params.Ctx, &container.Config{
		AttachStdout: true,
		AttachStderr: true,
		Image:        tag,
		Cmd:          []string{"/bin/sleep", "infinity"},
	}, hostCont, nil, "")

	if err != nil {
		return errors.Wrapf(err, "Error creating container %s", tag)
	}

	err = de.cl.ContainerStart(de.params.Ctx, r.ID,
		types.ContainerStartOptions{})

	if err != nil {
		return errors.Wrapf(err, "Error starting container: %s", tag)
	}

	return nil
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

		return rerr
	}

	return err
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
