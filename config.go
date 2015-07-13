// Copyright (c) 2015 Fred Zhou

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
package next

import (
	"io/ioutil"
	"os"
	"strings"
)

// Usage:
//
//	import "github.com/api4me/next"
//
//	conf := next.Config{}
//
type Config struct {
	data *Json
}

func NewConfig() *Config {
	return &Config{
		data: NewJson(),
	}
}

func (cfg *Config) Load(name string) (*Config, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return cfg.Read(content)
}

func (cfg *Config) Read(content []byte) (*Config, error) {
	_, err := cfg.data.Load([]byte(content))
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// String returns the string value for given key.
// val := conf.Get('redis.key')
func (cfg *Config) String(key string) string {
	k := strings.Split(key, ".")
	return cfg.data.GetPath(k...).MustString()
}

// Bool returns the string value for given key.
// val := conf.Get('redis.key')
func (cfg *Config) Bool(key string) bool {
	k := strings.Split(key, ".")
	return cfg.data.GetPath(k...).MustBool()
}

// Bool returns the string value for given key.
// val := conf.Get('redis.key')
func (cfg *Config) Int(key string) int {
	k := strings.Split(key, ".")
	return cfg.data.GetPath(k...).MustInt()
}

// String returns the string value for given key.
// cfg, err := conf.Set('redis.key', 'key123')
func (cfg *Config) Set(key, val string) {
	cfg.data.SetPath(strings.Split(key, "."), val)
}
