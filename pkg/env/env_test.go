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
	"os"
	"path/filepath"
	"strconv"
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

	ceng.On("CreateNetwork", mock.Anything, mock.Anything).
		Return("net-id", "10.0.0.0/24", nil)
	ceng.On("RemoveNetwork", mock.Anything, mock.AnythingOfType("string")).
		Return(nil)

	ceng.On("GetImagePorts", mock.Anything, bimgName).Return([]uint16(nil), nil)
	ceng.On("FetchImage", mock.Anything, fimgName).Return(nil)
	ceng.On("BuildImage", mock.Anything, bimgName, mock.Anything).Return(nil)

	ceng.On("RunContainer", mock.Anything,
		fmt.Sprintf("%s.0.ok", fcontName), fimgName,
		mock.Anything).Return("fcont-id", nil)

	ceng.On("RunContainer", mock.Anything,
		fmt.Sprintf("%s.0.ok", bcontName), bimgName,
		mock.Anything).Return("bcont-id", nil)

	ceng.On("RemoveContainer", mock.Anything, mock.Anything).Return(nil)
	ceng.On("RemoveImage", mock.Anything, mock.Anything).Return(nil)

	cwd, err := os.Getwd()
	require.Nil(t, err)

	tmpDir := filepath.Join(os.TempDir(), "xenvman-test-"+lib.NewId())

	binaryData := "123456789"
	binary := base64.StdEncoding.EncodeToString([]byte(binaryData))

	envName := "test"
	tplName := "ok"

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
					},
				},
			},
			Options: &def.EnvOptions{},
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
	require.Equal(t, string(bytes),
		fmt.Sprintf("%s.0.%s", bcontName, tplName))

	env.KeepAlive()

	// Make sure ports are properly exposed
	exported := env.Export()

	eb := exported.Templates["ok"][0].Containers[bcontName].Ports[bport]
	ef := exported.Templates["ok"][0].Containers[fcontName].Ports[fport]

	bsplit := strings.Split(eb, ":")
	fsplit := strings.Split(ef, ":")

	ebp, err := strconv.ParseInt(bsplit[1], 0, 16)
	require.Nil(t, err)
	fbp, err := strconv.ParseInt(bsplit[1], 0, 16)
	require.Nil(t, err)

	require.Equal(t, bsplit[0], "localhost")
	require.True(t, ebp >= 20000 && ebp <= 30000)

	require.Equal(t, fsplit[0], "localhost")
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

	// Mock assertion
	ceng.AssertCalled(t, "FetchImage", mock.Anything, fimgName)
	ceng.AssertCalled(t, "BuildImage", mock.Anything, bimgName, mock.Anything)

	ceng.AssertCalled(t, "RunContainer", mock.Anything,
		fmt.Sprintf("%s.0.ok", fcontName), fimgName,
		mock.Anything)

	ceng.AssertCalled(t, "RunContainer", mock.Anything,
		fmt.Sprintf("%s.0.ok", bcontName), bimgName,
		mock.Anything)

	ceng.AssertCalled(t, "RemoveContainer", mock.Anything, "fcont-id")
	ceng.AssertCalled(t, "RemoveContainer", mock.Anything, "bcont-id")

	ceng.AssertCalled(t, "RemoveImage", mock.Anything, bimgName)

	//TODO: Readiness checks
}
