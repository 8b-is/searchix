//go:build !embed

package frontend

import (
	"os"
)

var Files = os.DirFS("frontend/")
