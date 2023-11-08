package types

import (
	"os"
)

const (
	certFilePerm = os.FileMode(0644)
	certDirPerm  = os.FileMode(0755)
)
