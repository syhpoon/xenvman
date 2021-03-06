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

package server

import (
	"fmt"
	"net/http"

	"github.com/syhpoon/xenvman/pkg/logger"
)

var authLog = logger.GetLogger("xenvman.pkg.server.auth")

type AuthBackend interface {
	Authenticate(req *http.Request) error
	fmt.Stringer
}

func apiAuthMiddleware(h http.HandlerFunc, auth AuthBackend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := auth.Authenticate(r); err != nil {
			authLog.Warningf("Auth error: %+v", err)

			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

			ApiSendMessage(w, http.StatusUnauthorized, err.Error())

			return
		}

		h.ServeHTTP(w, r)
	}
}
