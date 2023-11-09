package client

import (
	"os"
	"runtime"
)

func getOS() OSType {
	return OSTypeFromString(runtime.GOOS)
}

func fileExist(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}
	return false
}
