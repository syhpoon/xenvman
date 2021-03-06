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

package tpl

import (
	"github.com/pkg/errors"
)

type Opts map[string]interface{}

func (o Opts) GetBool(key string, def bool) bool {
	if v, ok := o[key]; ok {
		if boolv, ok := v.(bool); !ok {
			panic(errors.Errorf("Invalid type for opt %s, expected bool got %T",
				key, v))
		} else {
			return boolv
		}
	} else {
		return def
	}
}

func (o Opts) GetInt(key string, def int) int {
	if v, ok := o[key]; ok {
		if intv, ok := v.(int); !ok {
			panic(errors.Errorf("Invalid type for opt %s, expected int got %T",
				key, v))
		} else {
			return intv
		}
	} else {
		return def
	}
}

func (o Opts) GetObject(key string, def map[string]interface{}) map[string]interface{} {
	if v, ok := o[key]; ok {
		if objv, ok := v.(map[string]interface{}); !ok {
			panic(errors.Errorf("Invalid type for opt %s, expected map[string]interface{} got %T",
				key, v))
		} else {
			return objv
		}
	} else {
		return def
	}
}
