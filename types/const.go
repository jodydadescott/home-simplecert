package types

import (
	"os"
)

const (
	FilePerm       = os.FileMode(0644)
	DirPerm        = os.FileMode(0755)
	ExePerm        = os.FileMode(0755)
	SecureFilePerm = os.FileMode(0400)

	CodeVersion = "1.0.0"
)
