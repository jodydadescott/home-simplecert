package cmd

import (
	"fmt"
	"path/filepath"
)

const (
	DefaultConfigFile  = "/etc/home-simplecert.yaml"
	SystemdServiceFile = "/etc/systemd/system/home-simplecert.service"
	ConfigEnvVar       = "CONFIG"
	DebugEnvVar        = "DEBUG"
	ConfigNotes        = "Config should have client config or server config. It is possible to have both."
	BinaryInstallPath  = "/usr/sbin"
	BinaryName         = "home-simplecert"
)

func systemD() string {

	s := "[Unit]\n"
	s += "Description=home-simplecert fetches TLS certs from remote and install them locally\n"
	s += "Requires=network.target\n"
	s += "After=network.target\n"
	s += "\n"
	s += "[Service]\n"
	s += "Type=simple\n"
	s += "Restart=always\n"
	s += "RestartSec=10\n"
	s += fmt.Sprintf("ExecStart=%s run\n", filepath.Join(BinaryInstallPath, BinaryName))
	s += "\n"
	s += "[Install]\n"
	s += "WantedBy=multi-user.target\n"
	s += "\n"

	return s
}
