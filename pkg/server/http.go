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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/syhpoon/xenvman/pkg/def"
)

func ApiSendMessage(w http.ResponseWriter, code int,
	msg string, args ...interface{}) {

	e := def.ApiResponse{}

	if msg != "" {
		e.Message = fmt.Sprintf(msg, args...)
	}

	ApiSendReply(w, code, &e)
}

func ApiSendData(w http.ResponseWriter, code int, data interface{}) {
	ApiSendReply(w, code, &def.ApiResponse{
		Data: data,
	})
}

func ApiSendReply(w http.ResponseWriter, code int, resp *def.ApiResponse) {
	body, _ := json.MarshalIndent(resp, "", "   ")

	_ = SendHttpResponse(w, code, nil, body)
}

func SendHttpResponse(w http.ResponseWriter, code int,
	headers interface{}, body interface{}) error {

	foundContentType := false

	if headers != nil {
		switch h := headers.(type) {
		case map[string]string:
			for k, v := range h {
				w.Header().Set(k, v)
			}

		case http.Header:
			for k, v := range h {
				for _, value := range v {
					w.Header().Add(k, value)
				}

				if strings.ToLower(k) == "content-type" {
					foundContentType = true
				}
			}
		default:
			err := fmt.Errorf(
				"Invalid headers type %T, map[string]string or http.Header expected", body)

			serverLog.Errorf(err.Error())

			return err
		}
	}

	if !foundContentType {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	w.WriteHeader(code)

	if body != nil {
		switch b := body.(type) {
		case []byte:
			_, err := w.Write(b)

			return err
		case string:
			_, err := io.WriteString(w, b)

			return err
		case io.ReadCloser:
			_body, err := ioutil.ReadAll(b)

			if err != nil {
				return err
			}

			_ = b.Close()

			_, err = w.Write(_body)

			return err
		default:
			err := fmt.Errorf("Invalid body type %T, []byte or string expected", body)

			serverLog.Errorf(err.Error())

			return err
		}
	}

	return nil
}
