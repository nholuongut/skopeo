// +build !linux

package vfs // import "github.com/nholuongut/storage/drivers/vfs"

import "github.com/nholuongut/storage/pkg/chrootarchive"

func dirCopy(srcDir, dstDir string) error {
	return chrootarchive.NewArchiver(nil).CopyWithTar(srcDir, dstDir)
}
