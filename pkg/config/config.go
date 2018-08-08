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

package config

import (
	"bytes"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Init configuration from a given file.
func InitConfig(path string) error {
	var reader io.Reader

	// First read default values
	reader = bytes.NewBuffer(defaultConfig)
	viper.SetConfigType("toml")

	err := viper.ReadConfig(reader)

	if err != nil {
		log.Fatal(err)
	}

	if path != "" {
		file, err := os.Open(path)

		if err != nil {
			return err
		}

		defer file.Close()

		reader = file

		viper.SetConfigType(filepath.Ext(path)[1:])

		err = viper.MergeConfig(reader)

		if err != nil {
			return err
		}
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("XENV")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return nil
}

func GetString(key string) string {
	return viper.GetString(key)
}

func GetStrings(key string) []string {
	return viper.GetStringSlice(key)
}

func GetDuration(key string) time.Duration {
	return viper.GetDuration(key)
}

func GetInt(key string) int {
	return viper.GetInt(key)
}

func GetUint64(key string) uint64 {
	return uint64(viper.GetInt64(key))
}

func GetUint(key string) uint {
	if !viper.IsSet(key) {
		return uint(0)
	}

	tmp := viper.GetString(key)

	ui64, err := strconv.ParseUint(tmp, 10, 32)

	if err != nil {
		return uint(0)
	}

	return uint(ui64)
}

func Get(key string) interface{} {
	return viper.Get(key)
}

func IsSet(key string) bool {
	return viper.IsSet(key)
}

func BindPFlag(key string, flag *pflag.Flag) error {
	return viper.BindPFlag(key, flag)
}

func AllSettings() map[string]interface{} {
	return viper.AllSettings()
}
