package client

import "time"

const (
	DetectSynologyFile = "/proc/syno_platform"
	DetectUnifiFile    = "/sys/fs/cgroup/system.slice/unifi.service/cgroup.type"

	DefaultRefreshInterval = time.Hour * 24

	ConfigNotes = "RefreshInterval is optional. It is only used if daemon is set to true. If the system type is Synology only the domain Name is required (not CertFile, KeyFile, KeyStore or Hook)"

	UnifiCertFile = "/data/unifi-core/config/unifi-core.crt"
	UnifiKeyFile  = "/data/unifi-core/config/unifi-core.key"

	SynologyDefaultFile = "/usr/syno/etc/certificate/_archive/DEFAULT"
	SynologyCertFile    = "cert.pem"
	SynologyKeyFile     = "privkey.pem"
	SynologyChainFile   = "fullchain.pem"
	Synology            = "Synology"
	Unifi               = "Unifi"
)

type OSType string

const (
	OSTypeEmpty    OSType = ""
	OSTypeSynology OSType = "synology"
	OSTypeUnifi    OSType = "unifi"
	OSTypeLinux    OSType = "linux"
	OSTypeDarwin   OSType = "darwin"
	OSTypeWindows  OSType = "windows"
	OSTypeUnknown  OSType = "unknown"
)

func OSTypeFromString(s string) OSType {

	switch s {

	case string(OSTypeEmpty):
		return OSTypeEmpty

	case string(OSTypeSynology):
		return OSTypeSynology

	case string(OSTypeUnifi):
		return OSTypeUnifi

	case string(OSTypeLinux):
		return OSTypeLinux

	case string(OSTypeDarwin):
		return OSTypeDarwin

	case string(OSTypeWindows):
		return OSTypeWindows

	}

	return OSTypeUnknown
}
