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
	"context"
	"io"

	"github.com/stretchr/testify/mock"
)

type MockedEngine struct {
	mock.Mock
}

func (me *MockedEngine) CreateNetwork(ctx context.Context, name string) (NetworkId, string, error) {
	args := me.Called(ctx, name)

	return args.String(0), args.String(1), args.Error(2)
}

func (me *MockedEngine) BuildImage(ctx context.Context, imgName string,
	buildContext io.Reader) error {
	args := me.Called(ctx, imgName, buildContext)

	return args.Error(0)
}

func (me *MockedEngine) GetImagePorts(ctx context.Context,
	imgName string) ([]uint16, error) {
	args := me.Called(ctx, imgName)

	return args.Get(0).([]uint16), args.Error(1)
}

func (me *MockedEngine) RemoveImage(ctx context.Context, imgName string) error {
	args := me.Called(ctx, imgName)

	return args.Error(0)
}

func (me *MockedEngine) RunContainer(ctx context.Context, name, tag string,
	params RunContainerParams) (string, error) {
	args := me.Called(ctx, name, tag, params)

	return args.String(0), args.Error(1)
}

func (me *MockedEngine) RemoveContainer(ctx context.Context, id string) error {
	args := me.Called(ctx, id)

	return args.Error(0)
}

func (me *MockedEngine) StopContainer(ctx context.Context, id string) error {
	args := me.Called(ctx, id)

	return args.Error(0)
}

func (me *MockedEngine) RestartContainer(ctx context.Context, id string) error {
	args := me.Called(ctx, id)

	return args.Error(0)
}

func (me *MockedEngine) RemoveNetwork(ctx context.Context, id string) error {
	args := me.Called(ctx, id)

	return args.Error(0)
}

func (me *MockedEngine) FetchImage(ctx context.Context, imgName string) error {
	args := me.Called(ctx, imgName)

	return args.Error(0)
}

func (me *MockedEngine) Terminate() {
	me.Called()
}
