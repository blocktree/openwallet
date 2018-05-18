package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CopyEnv describes the environment in which a copy is invoked.
type CopyEnv struct {
	FollowSymbolicLinks bool
	Overwrite           bool
	Verbose             bool
}

func (c *CopyEnv) Copy(src, dst string) error {
	return nil
}

var std = &CopyEnv{
	FollowSymbolicLinks: false,
	Overwrite:           false,
	Verbose:             false,
}

// Copy copies the file or directory from source path to destination path with
// the standart copy environment. For more control create a custom copy environment.
func Copy(src, dst string) error {
	if dst == "." {
		dst = filepath.Base(src)
	}

	if src == dst {
		return fmt.Errorf("%s and %s are identical (not copied).", src, dst)
	}

	if !Exists(src) {
		return fmt.Errorf("%s: no such file or directory.", src)
	}

	if Exists(dst) && IsFile(dst) {
		return fmt.Errorf("%s is a directory (not copied).", src)
	}

	srcBase, _ := filepath.Split(src)
	walks := 0

	// dstPath returns the rewritten destination path for the given source path
	dstPath := func(srcPath string) string {
		// some/random/long/path/example/hello.txt -> example/hello.txt
		srcPath = strings.TrimPrefix(srcPath, srcBase)

		// example/hello.txt -> destination/example/hello.txt
		if walks != 0 {
			return filepath.Join(dst, srcPath)
		}

		// hello.txt -> example/hello.txt
		if Exists(dst) && !IsFile(dst) {
			return filepath.Join(dst, filepath.Base(srcPath))
		}

		// hello.txt -> test.txt
		return dst
	}

	filepath.Walk(src, func(srcPath string, file os.FileInfo, err error) error {
		defer func() { walks++ }()

		if file.IsDir() {
			//fmt.Printf("copy dir from '%s' to '%s'\n", srcPath, dstPath(srcPath))
			os.MkdirAll(dstPath(srcPath), 0755)
		} else {
			//fmt.Printf("copy file from '%s' to '%s'\n", srcPath, dstPath(srcPath))
			err = copyFile(srcPath, dstPath(srcPath))
			if err != nil {
				fmt.Println(err)
			}
		}

		return nil
	})

	return nil
}
