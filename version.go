// Package fox
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
package fox

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/uuid"
)

// version fox version
const version = "v0.0.0"

var (
	appId     string // 当前应用唯一标识
	startTime string // 项目开始时间
	goVersion string // go版本号
)

// 构建信息
var (
	appName     string // 应用名称
	hostName    string // 主机名
	appVersion  string // 应用版本
	buildTime   string // 构建时间
	buildUser   string // 构建的用户
	buildStatus string // 应用构建状态
	buildHost   string // 构建的主机
)

// 环境变量设置
var (
	appRegion string
	appZone   string
)

// init 初始化
func init() {
	appId = uuid.New().String()
	if appName == "" {
		appName = os.Getenv("APP_NAME")
		if appName == "" {
			appName = filepath.Base(os.Args[0])
		}
	}
	name, err := os.Hostname()
	if err != nil {
		name = "unknown"
	}
	hostName = name
	startTime = time.Now().Format("2006-01-02 15:04:05")
	buildTime = strings.Replace(buildTime, "--", " ", 1)
	goVersion = runtime.Version()
}

// AppId 当前应用唯一标识
func AppId() string {
	return appId
}

// AppName 应用名称
func AppName() string {
	return appName
}

// AppVersion 应用版本
func AppVersion() string {
	return appVersion
}

// BuildTime 构建时间
func BuildTime() string {
	return buildTime
}

// HostName 主机名
func HostName() string {
	return hostName
}

// GoVersion go的运行版本
func GoVersion() string {
	return goVersion
}

// StartTime 项目启动时间
func StartTime() string {
	return startTime
}

// SetAppRegion 设置部属区域
func SetAppRegion(region string) {
	appRegion = region
}

// AppRegion 部属地域
func AppRegion() string {
	return appRegion
}

// SetAppZone set app zone
func SetAppZone(z string) {
	appZone = z
}

// AppZone 应用部属的分区
func AppZone() string {
	return appZone
}

// PrintVersion print version
func PrintVersion() {
	fmt.Printf("[%-3s]> %-30s => %s\n", "fox", color.RedString("name"), color.BlueString(appName))
	fmt.Printf("[%-3s]> %-30s => %s\n", "fox", color.RedString("version"), color.BlueString(appVersion))
	fmt.Printf("[%-3s]> %-30s => %s\n", "fox", color.RedString("hostname"), color.BlueString(hostName))
	fmt.Printf("[%-3s]> %-30s => %s\n", "fox", color.RedString("foxVersion"), color.BlueString(version))
	fmt.Printf("[%-3s]> %-30s => %s\n", "fox", color.RedString("goVersion"), color.BlueString(goVersion))
	fmt.Printf("[%-3s]> %-30s => %s\n", "fox", color.RedString("buildUser"), color.BlueString(buildUser))
	fmt.Printf("[%-3s]> %-30s => %s\n", "fox", color.RedString("buildHost"), color.BlueString(buildHost))
	fmt.Printf("[%-3s]> %-30s => %s\n", "fox", color.RedString("buildTime"), color.BlueString(buildTime))
	fmt.Printf("[%-3s]> %-30s => %s\n", "fox", color.RedString("buildStatus"), color.BlueString(buildStatus))
}

// VersionFox fox version
func VersionFox() string {
	return version
}
