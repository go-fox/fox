package generator

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

const (
	OsLinux   = "linux"
	OsMac     = "darwin"
	OsWindows = "windows"
	NewLine   = "\n"
)

// Exec 使用命令行终端执行命令
func Exec(arg string, dir string, in ...*bytes.Buffer) (string, error) {
	goos := runtime.GOOS
	var cmd *exec.Cmd
	switch goos {
	case OsMac, OsLinux:
		cmd = exec.Command("sh", "-c", arg)
	case OsWindows:
		cmd = exec.Command("cmd.exe", "/c", arg)
	default:
		return "", fmt.Errorf("unexpected os: %v", goos)
	}
	if len(dir) > 0 {
		cmd.Dir = dir
	}
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	if len(in) > 0 {
		cmd.Stdin = in[0]
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return "", errors.New(strings.TrimSuffix(stderr.String(), NewLine))
		}
		return "", err
	}
	return strings.TrimSuffix(stdout.String(), NewLine), nil
}
