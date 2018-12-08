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
	"os"
	"path/filepath"
	"testing"

	"github.com/syhpoon/xenvman/pkg/lib"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/syhpoon/xenvman/pkg/conteng"
	"github.com/syhpoon/xenvman/pkg/def"
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

	imgName := "test-image"
	contName := "test-cont"
	label := "wut"

	ceng.On("RemoveNetwork", mock.Anything, mock.AnythingOfType("string")).
		Return(nil)

	ceng.On("FetchImage", mock.Anything, imgName).Return(nil)

	ceng.On("CreateNetwork", mock.Anything, mock.Anything).
		Return("net-id", "10.0.0.0/24", nil)

	ceng.On("RunContainer", mock.Anything, "test-cont.0.ok", imgName,
		mock.Anything).Return("cont-id", nil)

	ceng.On("RemoveContainer", mock.Anything, "cont-id").Return(nil)

	cwd, err := os.Getwd()
	require.Nil(t, err)

	tmpDir := filepath.Join(os.TempDir(), "xenvman-test-"+lib.NewId())

	env, err := NewEnv(Params{
		EnvDef: &def.InputEnv{
			Name: "test",
			Templates: []*def.Tpl{
				{
					Tpl: "ok",
					Parameters: map[string]interface{}{
						"image":     imgName,
						"container": contName,
						"label":     label,
					},
				},
			},
			Options: &def.EnvOptions{},
		},
		ContEng:      ceng,
		BaseTplDir:   filepath.Join(cwd, "./testdata"),
		BaseWsDir:    filepath.Join(tmpDir, "ws"),
		BaseMountDir: filepath.Join(tmpDir, "mount"),
		Ctx:          ctx,
	})

	_ = os.RemoveAll(tmpDir)

	require.Nil(t, err)
	require.Len(t, env.containers, 1)
}
