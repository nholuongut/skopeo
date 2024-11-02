// +build windows

package graphdriver

import (
	"os"
	"syscall"

	"github.com/nholuongut/storage/pkg/idtools"
)

func platformLChown(path string, info os.FileInfo, toHost, toContainer *idtools.IDMappings) error {
	return &os.PathError{"lchown", path, syscall.EWINDOWS}
}
