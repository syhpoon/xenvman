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
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/syhpoon/xenvman/pkg/conteng"
	"github.com/syhpoon/xenvman/pkg/def"
	"github.com/syhpoon/xenvman/pkg/lib"
)

func TestEnvNonesistentTemplate(t *testing.T) {
	ctx := context.Background()
	ceng := new(conteng.MockedEngine)

	ceng.On("RemoveNetwork", mock.Anything, mock.AnythingOfType("string")).
		Return(nil)

	_, err := NewEnv(Params{
		EnvDef: &def.InputEnv{
			Name: "test",
			Templates: []*def.Tpl{
				{
					Tpl: "invalid",
				},
			},
		},
		ContEng:    ceng,
		BaseTplDir: "/tmp/nonexistent",
		Ctx:        ctx,
	})

	require.Contains(t, err.Error(), "no such file or directory")
}

func TestEnvInvalidTemplateName(t *testing.T) {
	ctx := context.Background()
	ceng := new(conteng.MockedEngine)

	ceng.On("RemoveNetwork", mock.Anything, mock.AnythingOfType("string")).
		Return(nil)

	params := Params{
		EnvDef: &def.InputEnv{
			Name:      "test",
			Templates: []*def.Tpl{{Tpl: " /etc/passwd"}},
		},
		ContEng:    ceng,
		BaseTplDir: "/tmp/nonexistent",
		Ctx:        ctx,
	}

	_, err := NewEnv(params)
	require.Contains(t, err.Error(), "no such file or directory")

	params.EnvDef.Templates[0].Tpl = "../out/and/out"
	_, err = NewEnv(params)
	require.Contains(t, err.Error(), "must not contain")
}

func TestEnvOk(t *testing.T) {
	ctx := context.Background()
	ceng := new(conteng.MockedEngine)

	fimgName := "test-image-fetch"
	fcontName := "test-cont-fetch"
	flabel := "wut-fetch"
	fport := 5000

	bimgName := "test-image-build"
	bcontName := "test-cont-build"
	blabel := "wut-build"
	bport := 6000

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.Nil(t, err)
	defer listener.Close()

	readinessPort := listener.Addr().(*net.TCPAddr).Port

	bimgMatcher := mock.MatchedBy(func(imgName string) bool {
		return strings.Contains(imgName, bimgName)
	})

	ceng.On("CreateNetwork", mock.Anything, mock.Anything).
		Return("net-id", "10.0.0.0/24", nil)
	ceng.On("RemoveNetwork", mock.Anything, mock.AnythingOfType("string")).
		Return(nil)

	ceng.On("GetImagePorts", mock.Anything,
		bimgMatcher).Return([]uint16(nil), nil)
	ceng.On("FetchImage", mock.Anything, fimgName).Return(nil)
	ceng.On("BuildImage", mock.Anything, bimgMatcher,
		mock.Anything).Return(nil)

	ceng.On("RunContainer", mock.Anything,
		fmt.Sprintf("%s.0.ok.xenv", fcontName), fimgName,
		mock.Anything).Return("fcont-id", nil)

	var bcontRunParams *conteng.RunContainerParams

	ceng.On("RunContainer", mock.Anything,
		fmt.Sprintf("%s.0.ok.xenv", bcontName), bimgMatcher,
		mock.Anything).Return("bcont-id", nil).Run(
		func(args mock.Arguments) {
			arg := args.Get(3).(conteng.RunContainerParams)
			bcontRunParams = &arg
		})

	ceng.On("RemoveContainer", mock.Anything, mock.Anything).Return(nil)
	ceng.On("RemoveImage", mock.Anything, mock.Anything).Return(nil)

	cwd, err := os.Getwd()
	require.Nil(t, err)

	tmpDir := filepath.Join(os.TempDir(), "xenvman-test-"+lib.NewId())

	binaryData := "123456789"
	binary := base64.StdEncoding.EncodeToString([]byte(binaryData))

	envName := "test"
	tplName := "ok"
	extraKey := "WUT-WAT"

	env, err := NewEnv(Params{
		EnvDef: &def.InputEnv{
			Name: envName,
			Templates: []*def.Tpl{
				{
					Tpl: tplName,
					Parameters: map[string]interface{}{
						"fimage":     fimgName,
						"fcontainer": fcontName,
						"flabel":     flabel,
						"fport":      fport,
						"bimage":     bimgName,
						"bcontainer": bcontName,
						"bport":      bport,
						"blabel":     blabel,
						"binary":     binary,
						"rport":      readinessPort,
						"extra_key":  extraKey,
					},
				},
			},
			Options: &def.EnvOptions{
				DisableDiscovery: true,
			},
		},
		ContEng:       ceng,
		BaseTplDir:    filepath.Join(cwd, "./testdata"),
		BaseWsDir:     filepath.Join(tmpDir, "ws"),
		BaseMountDir:  filepath.Join(tmpDir, "mount"),
		PortRange:     lib.NewPortRange(20000, 30000),
		ExportAddress: "localhost",
		Ctx:           ctx,
	})

	defer os.RemoveAll(tmpDir)

	require.Nil(t, err)
	require.Len(t, env.containers, 2)
	require.Contains(t, env.containers, "bcont-id")
	require.Contains(t, env.containers, "fcont-id")

	// Check workspace files
	wsGlob := filepath.Join(tmpDir, "ws", envName+"*", tplName, "0",
		"xenv-ok-"+bimgName+":*")

	wsBinary, err := filepath.Glob(filepath.Join(wsGlob, "binary"))
	require.Nil(t, err)
	require.Len(t, wsBinary, 1)

	binaryF, err := os.Open(wsBinary[0])
	require.Nil(t, err)

	stat, err := binaryF.Stat()
	require.Nil(t, err)

	require.Equal(t, os.FileMode(0755), stat.Mode())

	bytes, err := ioutil.ReadAll(binaryF)
	require.Nil(t, err)
	require.Equal(t, binaryData, string(bytes))

	wsWs, err := filepath.Glob(filepath.Join(wsGlob, "ws"))
	require.Nil(t, err)
	require.Len(t, wsWs, 1)

	bytes, err = ioutil.ReadFile(wsWs[0])
	require.Nil(t, err)
	require.Equal(t, string(bytes),
		fmt.Sprintf(">>> %s : %s <<<", bimgName, bcontName))

	// Check mount file
	mGlob := filepath.Join(tmpDir, "mount", envName+"*", "mounts",
		bcontName, "*", "mount")

	mountF, err := filepath.Glob(mGlob)
	require.Nil(t, err)
	require.Len(t, mountF, 1)

	bytes, err = ioutil.ReadFile(mountF[0])
	require.Nil(t, err)
	require.Equal(t, fmt.Sprintf("%s.0.%s.xenv\n>>>WUT-WAT<<<\n", bcontName, tplName),
		string(bytes))

	env.KeepAlive()

	// Make sure ports are properly exposed
	exported := env.Export()

	ebp := exported.Templates["ok"][0].Containers[bcontName].Ports[fmt.Sprintf("%d", bport)]
	fbp := exported.Templates["ok"][0].Containers[fcontName].Ports[fmt.Sprintf("%d", fport)]

	require.Equal(t, exported.ExternalAddress, "localhost")
	require.True(t, ebp >= 20000 && ebp <= 30000)
	require.True(t, fbp >= 20000 && fbp <= 30000)

	// Make sure temporary files have been cleaned after Terminate
	require.Nil(t, env.Terminate())

	wsglob := filepath.Join(tmpDir, "ws", "*")
	wsFiles, err := filepath.Glob(wsglob)
	require.Nil(t, err)
	require.Empty(t, wsFiles)

	mountglob := filepath.Join(tmpDir, "mounts", "*")
	mountFiles, err := filepath.Glob(mountglob)
	require.Nil(t, err)
	require.Empty(t, mountFiles)

	require.NotNil(t, bcontRunParams)
	require.Equal(t, bcontRunParams.Environ["INTERPOLATE-ME"],
		fmt.Sprintf("%s:%d", exported.ExternalAddress, ebp))
	require.Equal(t, bcontRunParams.Environ["DONT-INTERPOLATE-ME"], "WUT")

	// Mock assertion
	ceng.AssertCalled(t, "FetchImage", mock.Anything, fimgName)
	ceng.AssertCalled(t, "BuildImage", mock.Anything, bimgMatcher,
		mock.Anything)

	ceng.AssertCalled(t, "RunContainer", mock.Anything,
		fmt.Sprintf("%s.0.ok.xenv", fcontName), fimgName,
		mock.Anything)

	ceng.AssertCalled(t, "RunContainer", mock.Anything,
		fmt.Sprintf("%s.0.ok.xenv", bcontName), bimgMatcher,
		mock.Anything)

	ceng.AssertCalled(t, "RemoveContainer", mock.Anything, "fcont-id")
	ceng.AssertCalled(t, "RemoveContainer", mock.Anything, "bcont-id")

	ceng.AssertCalled(t, "RemoveImage", mock.Anything, bimgMatcher)
	ceng.AssertCalled(t, "RemoveNetwork", mock.Anything, mock.Anything)
	ceng.AssertNumberOfCalls(t, "RemoveNetwork", 1)
}

func TestStartStopContainers(t *testing.T) {
	ctx := context.Background()
	ceng := new(conteng.MockedEngine)

	imgName := "image"
	contName := "cont"

	imgMatcher := mock.MatchedBy(func(img string) bool {
		return strings.Contains(img, imgName)
	})

	ceng.On("CreateNetwork", mock.Anything, mock.Anything).
		Return("net-id", "10.0.0.0/24", nil)
	ceng.On("RemoveNetwork", mock.Anything, mock.AnythingOfType("string")).
		Return(nil)

	ceng.On("GetImagePorts", mock.Anything,
		imgMatcher).Return([]uint16(nil), nil)
	ceng.On("FetchImage", mock.Anything, imgName).Return(nil)

	ceng.On("RunContainer", mock.Anything,
		fmt.Sprintf("%s.0.simple.xenv", contName), imgName,
		mock.Anything).Return("cont-0", nil)

	ceng.On("RunContainer", mock.Anything,
		fmt.Sprintf("%s.1.simple.xenv", contName), imgName,
		mock.Anything).Return("cont-1", nil)

	ceng.On("StopContainer", mock.Anything, "cont-0").Return(nil)
	ceng.On("RestartContainer", mock.Anything, "cont-0").Return(nil)

	ceng.On("RemoveContainer", mock.Anything, mock.Anything).Return(nil)
	ceng.On("RemoveImage", mock.Anything, mock.Anything).Return(nil)

	cwd, err := os.Getwd()
	require.Nil(t, err)

	tmpDir := filepath.Join(os.TempDir(), "xenvman-test-"+lib.NewId())

	envName := "test"
	tplName := "simple"

	env, err := NewEnv(Params{
		EnvDef: &def.InputEnv{
			Name: envName,
			Templates: []*def.Tpl{
				{
					Tpl: tplName,
					Parameters: map[string]interface{}{
						"image":     imgName,
						"container": contName,
					},
				},
				{
					Tpl: tplName,
					Parameters: map[string]interface{}{
						"image":     imgName,
						"container": contName,
					},
				},
			},
			Options: &def.EnvOptions{
				DisableDiscovery: true,
			},
		},
		ContEng:       ceng,
		BaseTplDir:    filepath.Join(cwd, "./testdata"),
		BaseWsDir:     filepath.Join(tmpDir, "ws"),
		BaseMountDir:  filepath.Join(tmpDir, "mount"),
		PortRange:     lib.NewPortRange(20000, 30000),
		ExportAddress: "localhost",
		Ctx:           ctx,
	})

	defer os.RemoveAll(tmpDir)

	require.Nil(t, err)
	exported := env.Export()
	cid0 := exported.Templates["simple"][0].Containers[contName].Id
	cid1 := exported.Templates["simple"][1].Containers[contName].Id

	require.Nil(t, err)
	require.Len(t, env.containers, 2)
	require.Contains(t, env.containers, cid0)
	require.Contains(t, env.containers, cid1)

	err = env.StopContainers([]string{cid0})
	require.Nil(t, err)

	err = env.RestartContainers([]string{cid0})
	require.Nil(t, err)

	require.Nil(t, env.Terminate())

	// Mock assertion
	ceng.AssertNumberOfCalls(t, "FetchImage", 1)

	ceng.AssertCalled(t, "RunContainer", mock.Anything,
		fmt.Sprintf("%s.0.simple.xenv", contName), imgName,
		mock.Anything)

	ceng.AssertCalled(t, "RunContainer", mock.Anything,
		fmt.Sprintf("%s.1.simple.xenv", contName), imgMatcher,
		mock.Anything)

	ceng.AssertCalled(t, "StopContainer", mock.Anything, cid0)
	ceng.AssertCalled(t, "RestartContainer", mock.Anything, cid0)
}

func TestAddTemplates(t *testing.T) {
	ctx := context.Background()
	ceng := new(conteng.MockedEngine)

	imgName := "image"
	contName := "cont"

	imgMatcher := mock.MatchedBy(func(img string) bool {
		return strings.Contains(img, imgName)
	})

	ceng.On("CreateNetwork", mock.Anything, mock.Anything).
		Return("net-id", "10.0.0.0/24", nil)
	ceng.On("RemoveNetwork", mock.Anything, mock.AnythingOfType("string")).
		Return(nil)

	ceng.On("GetImagePorts", mock.Anything,
		imgMatcher).Return([]uint16(nil), nil)
	ceng.On("FetchImage", mock.Anything, imgName).Return(nil)

	ceng.On("RunContainer", mock.Anything,
		fmt.Sprintf("%s.0.simple.xenv", contName), imgName,
		mock.Anything).Return("cont-0", nil)

	ceng.On("RunContainer", mock.Anything,
		"appended.1.simple.xenv", imgName,
		mock.Anything).Return("cont-1", nil)

	ceng.On("RemoveContainer", mock.Anything, mock.Anything).Return(nil)
	ceng.On("RemoveImage", mock.Anything, mock.Anything).Return(nil)

	cwd, err := os.Getwd()
	require.Nil(t, err)

	tmpDir := filepath.Join(os.TempDir(), "xenvman-test-"+lib.NewId())

	envName := "test"
	tplName := "simple"

	env, err := NewEnv(Params{
		EnvDef: &def.InputEnv{
			Name: envName,
			Templates: []*def.Tpl{
				{
					Tpl: tplName,
					Parameters: map[string]interface{}{
						"image":     imgName,
						"container": contName,
					},
				},
			},
			Options: &def.EnvOptions{
				DisableDiscovery: true,
			},
		},
		ContEng:       ceng,
		BaseTplDir:    filepath.Join(cwd, "./testdata"),
		BaseWsDir:     filepath.Join(tmpDir, "ws"),
		BaseMountDir:  filepath.Join(tmpDir, "mount"),
		PortRange:     lib.NewPortRange(20000, 30000),
		ExportAddress: "localhost",
		Ctx:           ctx,
	})

	defer os.RemoveAll(tmpDir)
	defer env.Terminate()

	require.Nil(t, err)
	exported := env.Export()
	cid0 := exported.Templates["simple"][0].Containers[contName].Id

	require.Nil(t, err)
	require.Len(t, env.containers, 1)
	require.Contains(t, env.containers, cid0)

	err = env.ApplyTemplates([]*def.Tpl{
		{
			Tpl: tplName,
			Parameters: map[string]interface{}{
				"image":     imgName,
				"container": "appended",
				"mount":     true,
			},
		},
	}, false, false)
	require.Nil(t, err)

	exported2 := env.Export()
	cid1 := exported2.Templates["simple"][1].Containers["appended"].Id

	require.Nil(t, err)
	require.Len(t, env.containers, 2)
	require.Contains(t, env.containers, cid0)
	require.Contains(t, env.containers, cid1)

	// Check mounted
	mGlob := filepath.Join(tmpDir, "mount", envName+"*", "mounts",
		"appended", "*", "mount")

	mountF, err := filepath.Glob(mGlob)
	require.Nil(t, err)
	require.Len(t, mountF, 1)

	bytes, err := ioutil.ReadFile(mountF[0])
	require.Nil(t, err)
	require.Equal(t, "cont.0.simple.xenv\n", string(bytes))

	// Mock assertion
	ceng.AssertNumberOfCalls(t, "FetchImage", 2)

	ceng.AssertCalled(t, "RunContainer", mock.Anything,
		fmt.Sprintf("%s.0.simple.xenv", contName), imgName,
		mock.Anything)

	ceng.AssertCalled(t, "RunContainer", mock.Anything,
		"appended.1.simple.xenv", imgMatcher, mock.Anything)
}
