/*
 MIT License

 Copyright (c) 2017 Max Kuznetsov <syhpoon@syhpoon.ca>

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

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/syhpoon/xenvman/pkg/config"
	"github.com/syhpoon/xenvman/pkg/logger"
)

var flagConfig string

var RootCmd = &cobra.Command{
	Use:   "xenvman",
	Short: "Environment manager",
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initLogger)

	RootCmd.PersistentFlags().StringVarP(&flagConfig,
		"config", "c", "", "Path to configuration file")
}

func initConfig() {
	err := config.InitConfig(flagConfig)

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error in configuration: %s\n", err)
		os.Exit(1)
	}
}

func initLogger() {
	if config.IsSet("log.config") {
		err := logger.ConfigureLoggers(config.GetString("log.config"))

		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Unable to initialize logger: %s\n",
				err.Error())

			os.Exit(1)
		}
	}
}

func init() {
	RootCmd.AddCommand(discCmd)
	RootCmd.AddCommand(runCmd)
	RootCmd.AddCommand(versionCmd)
}
