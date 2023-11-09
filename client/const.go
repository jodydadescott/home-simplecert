package client

import (
	"os"
	"time"
)

const (
	defaultRefreshInterval = time.Hour * 500
	filePerm               = os.FileMode(0644)
	dirPerm                = os.FileMode(0755)

	synologyDefaultFile = "/usr/syno/etc/certificate/_archive/DEFAULT"
	synologyCertFile    = "cert.pem"
	synologyKeyFile     = "privkey.pem"
	synologyChainFile   = "fullchain.pem"
)

type ModeType string

const (
	NormalModeType   ModeType = "normal"
	SynologyModeType          = "synology"
)
