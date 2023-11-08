package client

import (
	"os"
	"time"
)

const (
	defaultRefreshInterval = time.Hour * 500
	filePerm               = os.FileMode(0644)
	dirPerm                = os.FileMode(0755)
)
