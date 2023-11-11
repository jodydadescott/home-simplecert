package client

import (
	"os"
	"runtime"
)

func getOS() OSType {

	osType := OSTypeFromString(runtime.GOOS)

	if osType == OSTypeLinux {
		if fileExist(synologyProcFile) {
			osType = OSTypeSynology
		}
	}

	return osType
}

func fileExist(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}
	return false
}
