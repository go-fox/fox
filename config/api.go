// Package config
// MIT License
//
// # Copyright (c) 2024 go-fox
// Author https://github.com/go-fox/fox
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package config

var defaultConfig = New()

// SetDefault Set default config
//
//	@param config Config
//	@player
func SetDefault(config Config) {
	defaultConfig = config
}

// Load load resources
//
//	@param sources ...Source
//	@return error
//	@player
func Load(sources ...Source) error {
	return defaultConfig.Load(sources...)
}

// MustLoad resources must be loaded. if there is an error, simply panic
//
//	@param sources ...Source
//	@player
func MustLoad(sources ...Source) {
	if err := Load(sources...); err != nil {
		panic(err)
	}
}

// Get get a value for defaultConfig
//
//	@param path string
//	@return Value
//	@player
func Get(path string) Value {
	return defaultConfig.Get(path)
}

// Scan scans the value to the specified structure for defaultConfig
//
//	@param v any
//	@return error
//	@player
func Scan(v any) error {
	return defaultConfig.Scan(v)
}

// Watch Configure Observer
//
//	@param key string
//	@param o Observer
//	@return error
//	@player
func Watch(key string, o Observer) error {
	return defaultConfig.Watch(key, o)
}

// Close closer all Observer
//
//	@return error
//	@player
func Close() error {
	return defaultConfig.Close()
}
