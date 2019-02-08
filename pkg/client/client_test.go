/*
 MIT License

 Copyright (c) 2019 Max Kuznetsov <syhpoon@syhpoon.ca>

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

package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/syhpoon/xenvman/pkg/def"
)

func TestNewEnv(t *testing.T) {
	out := &def.OutputEnv{
		Id:              "id",
		Name:            "name",
		Description:     "description",
		WsDir:           "ws-dir",
		MountDir:        "mount-dir",
		NetId:           "net-id",
		Created:         time.Now().Format(time.RFC3339),
		Keepalive:       "2m",
		ExternalAddress: "127.0.0.1",
		Templates: map[string][]*def.TplData{
			"tpl": {
				{
					Containers: map[string]*def.ContainerData{
						"cont": {
							Id:       "id",
							Hostname: "cont.xenv",
						},
					},
				},
			},
		},
	}

	srv := testSrv(out, t)
	defer srv.Close()

	cl := New(Params{
		ServerAddress: srv.URL,
	})

	env, err := cl.NewEnv(&def.InputEnv{})
	require.Nil(t, err)

	require.Equal(t, out, env.OutputEnv)
}

func TestListEnvs(t *testing.T) {
	ienvs := []*def.OutputEnv{
		{
			Id:   "id1",
			Name: "name1",
		},
		{
			Id:   "id2",
			Name: "name2",
		},
		{
			Id:   "id3",
			Name: "name3",
		},
	}

	srv := testSrv(ienvs, t)
	defer srv.Close()

	cl := New(Params{
		ServerAddress: srv.URL,
	})

	oenvs, err := cl.ListEnvs()
	require.Nil(t, err)

	require.Equal(t, len(ienvs), len(oenvs))

	for i := range ienvs {
		require.Equal(t, ienvs[i], oenvs[i].OutputEnv)
	}
}

func TestGetEnvInfo(t *testing.T) {
	ienv := &def.OutputEnv{
		Id:   "id",
		Name: "name",
	}

	srv := testSrv(ienv, t)
	defer srv.Close()

	cl := New(Params{
		ServerAddress: srv.URL,
	})

	oenv, err := cl.GetEnvInfo("id")
	require.Nil(t, err)

	require.Equal(t, ienv, oenv.OutputEnv)
}

func TestListTemplates(t *testing.T) {
	itpls := map[string]*def.TplInfo{
		"tpl1": {
			Description: "description",
			Parameters: map[string]*def.TplInfoParam{
				"param1": {
					Description: "Param #1",
					Type:        "string",
					Mandatory:   true,
				},
				"param2": {
					Description: "Param #2",
					Type:        "number",
					Mandatory:   false,
					Default:     float64(42),
				},
			},
			DataDir: []string{"a", "b", "c"},
		},
	}

	srv := testSrv(itpls, t)
	defer srv.Close()

	cl := New(Params{
		ServerAddress: srv.URL,
	})

	otpls, err := cl.ListTemplates()
	require.Nil(t, err)

	require.Equal(t, itpls, otpls)
}

func TestEnvTerminate(t *testing.T) {
	srv := testSrv("Env deleted", t)
	defer srv.Close()

	env := &Env{
		OutputEnv: &def.OutputEnv{
			Id: "id",
		},
		serverAddress: srv.URL,
		httpClient: http.Client{
			Timeout: 3 * time.Second,
		},
	}

	require.Nil(t, env.Terminate())
}

func TestEnvPatch(t *testing.T) {
	ienv := &def.OutputEnv{
		Id:   "id",
		Name: "name",
	}

	srv := testSrv(ienv, t)
	defer srv.Close()

	env := &Env{
		OutputEnv: &def.OutputEnv{
			Id: "id",
		},
		serverAddress: srv.URL,
		httpClient: http.Client{
			Timeout: 3 * time.Second,
		},
	}

	_, err := env.Patch(&def.PatchEnv{})
	require.Nil(t, err)

	require.Equal(t, ienv, env.OutputEnv)
}

func TestEnvKeepalive(t *testing.T) {
	srv := testSrv("", t)
	defer srv.Close()

	env := &Env{
		OutputEnv: &def.OutputEnv{
			Id: "id",
		},
		serverAddress: srv.URL,
		httpClient: http.Client{
			Timeout: 3 * time.Second,
		},
	}

	require.Nil(t, env.Keepalive())
}

func testSrv(out interface{}, t *testing.T) *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)

			var res []byte

			switch v := out.(type) {
			case string:
				res = []byte(v)
			default:
				b, err := json.Marshal(def.ApiResponse{
					Data: out,
				})
				require.Nil(t, err)

				res = b
			}

			_, err := w.Write(res)
			require.Nil(t, err)
		}))
}
