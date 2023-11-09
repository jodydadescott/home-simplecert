package client

import (
	"os"
	"time"
)

const (
	defaultRefreshInterval = time.Hour * 500
	filePerm               = os.FileMode(0644)
	dirPerm                = os.FileMode(0755)
	synologyProcFile       = "/proc/syno_platform"
	synologyDefaultFile    = "/usr/syno/etc/certificate/_archive/DEFAULT"
	synologyCertFile       = "cert.pem"
	synologyKeyFile        = "privkey.pem"
	synologyChainFile      = "fullchain.pem"
)

type OSType string

const (
	OSTypeEmpty    OSType = ""
	OSTypeSynology        = "synology"
	OSTypeLinux           = "linux"
	OSTypeDarwin          = "darwin"
	OSTypeWindows         = "windows"
	OSTypeUnknown         = "unknown"
)

func OSTypeFromString(s string) OSType {

	switch s {

	case string(OSTypeEmpty):
		return OSTypeEmpty

	case string(OSTypeSynology):
		return OSTypeSynology

	case string(OSTypeLinux):
		return OSTypeLinux

	case string(OSTypeDarwin):
		return OSTypeDarwin

	case string(OSTypeWindows):
		return OSTypeWindows

	}

	return OSTypeUnknown
}
