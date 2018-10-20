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
	"fmt"

	"encoding/base64"

	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"
)

func libBase() map[string]interface{} {
	return map[string]interface{}{
		"IsArray":   jsIsArray,
		"IsDefined": jsIsDefined,
	}
}

func jsIsArray(obj otto.Value) bool {
	return obj.Class() == "Array"
}

func jsIsDefined(obj otto.Value) bool {
	return !obj.IsUndefined() && !obj.IsNull()
}

func libType() map[string]interface{} {
	return map[string]interface{}{
		"EnsureString":        jsEnsureString,
		"EnsureListOfStrings": jsEnsureListOfStrings,
		"EnsureListOfNumbers": jsEnsureListOfNumbers,
		"FromBase64":          jsFromBase64,
	}
}

func jsEnsureString(name string, obj otto.Value) {
	if obj.IsNull() || obj.IsUndefined() {
		return
	}

	v, err := obj.Export()

	if err != nil {
		panic(errors.Errorf("%s: Error exporting js value: %s", name, err))
	}

	_, ok := v.(string)

	if !ok {
		panic(errors.Errorf("%s: Expected string but got %T", name, v))
	}
}

func jsEnsureListOfStrings(name string, obj otto.Value) {
	if obj.IsNull() || obj.IsUndefined() {
		return
	}

	v, err := obj.Export()

	if err != nil {
		panic(errors.Errorf("%s: Error exporting js value: %s", name, err))
	}

	arr, ok := v.([]interface{})

	if !ok {
		panic(errors.Errorf("%s: Expected array of strings but got %T",
			name, v))
	}

	for _, val := range arr {
		if _, ok := val.(string); !ok {
			panic(errors.Errorf("%s: Found non-string array element: %T",
				name, val))
		}
	}
}

func jsEnsureListOfNumbers(name string, obj otto.Value) {
	if obj.IsNull() || obj.IsUndefined() {
		return
	}

	v, err := obj.Export()

	if err != nil {
		panic(errors.Errorf("%s: Error exporting js value: %s", name, err))
	}

	arr, ok := v.([]interface{})

	if !ok {
		panic(errors.Errorf("%s: Expected array of strings but got %T",
			name, v))
	}

	for _, val := range arr {
		if _, ok := val.(float64); !ok {
			panic(errors.Errorf("%s: Found non-int array element: %T",
				name, val))
		}
	}
}

func jsFromBase64(name, obj otto.Value) []byte {
	v, err := obj.Export()

	if err != nil {
		panic(errors.Errorf("%s: Error exporting js value: %s", name, err))
	}

	s, ok := v.(string)

	if !ok {
		panic(errors.Errorf("%s: Expected string but got %T", name, v))
	}

	b, err := base64.StdEncoding.DecodeString(s)

	if err != nil {
		panic(errors.Errorf("%s: Error decoding base64: %s", name, err))
	}

	return b

}

func setupLib(vm *otto.Otto) {
	_ = vm.Set("fmt", fmt.Sprintf)
	_ = vm.Set("lib", libBase())
	_ = vm.Set("type", libType())
}
