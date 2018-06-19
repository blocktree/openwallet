// Package file provides simple utility functions to be used with files
package file

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"log"
)

// FileInfo describes a file. It's a wrapper around os.FileInfo.
type FileInfo struct {
	Exists bool
	os.FileInfo
}

// Stat returns a FileInfo describing the named file. It's a wrapper around os.Stat.
func Stat(file string) (*FileInfo, error) {
	fi, err := os.Stat(file)
	if err == nil {
		f := &FileInfo{Exists: true}
		f.FileInfo = fi
		return f, nil
	}

	if os.IsNotExist(err) {
		f := &FileInfo{Exists: false}
		f.FileInfo = nil
		return f, nil
	}

	return nil, err
}

// IsFile checks wether the given file is a directory or not. It panics if an
// error is occured. Use IsFileOk to use the returned error.
func IsFile(file string) bool {
	ok, err := IsFileOk(file)
	if err != nil {
		panic(err)
	}

	return ok
}

// IsFileOk checks whether the given file is a directory or not.
func IsFileOk(file string) (bool, error) {
	sf, err := os.Open(file)
	if err != nil {
		return false, err
	}
	defer sf.Close()

	fi, err := sf.Stat()
	if err != nil {
		return false, err
	}

	if fi.IsDir() {
		return false, nil
	}

	return true, nil
}

// Exists checks whether the given file exists or not. It panics if an error
// is occured. Use ExistsOk to use the returned error.
func Exists(file string) bool {
	ok, err := ExistsOk(file)
	if err != nil {
		panic(err)
	}

	return ok
}

// ExistsOk checks whether the given file exists or not.
func ExistsOk(file string) (bool, error) {
	_, err := os.Stat(file)
	if err == nil {
		return true, nil // file exist
	}

	if os.IsNotExist(err) {
		return false, nil // file does not exist
	}

	return false, err
}

func copyFile(src, dst string) error {
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()

	fi, err := sf.Stat()
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return errors.New("src is a directory, please provide a file")
	}

	df, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fi.Mode())
	if err != nil {
		return err
	}
	defer df.Close()

	if _, err := io.Copy(df, sf); err != nil {
		return err
	}

	return nil
}

// GetCurrentPath 获取当前执行文件的路径
func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return "", errors.New(`error: Can't find "/" or "\".`)
	}
	return string(path[0 : i+1]), nil
}

// GetAbsolutePath 获取当前执行文件的绝对路径目录
// subfolder 子目录，如果没有，自动创建
func GetAbsolutePath(subfolder ...string) (string, error) {
	var (
		dir string
		err error
	)
	dir, err = os.Getwd()
	if len(subfolder) > 0 {
		dir = dir + subfolder[0]
		err = os.MkdirAll(dir, os.ModePerm) //生成多级目录
	}
	if err != nil {
		log.Println(err)
	}
	return dir, err
}

//MkdirAll 创建文件夹
//@param file 文件夹路径
func MkdirAll(file string) bool {

	if err := os.MkdirAll(file, os.ModePerm); err != nil {
		log.Printf("%v\n", err)
		return false
	}
	return true
}

//WriteFile 写入内容到文件
//@param name 文件名
//@param content 内容
//@param append 是否追加到末尾
func WriteFile(name string, content []byte, append bool) bool {

	var (
		flag int
	)

	if append {
		flag = os.O_RDWR | os.O_CREATE | os.O_APPEND
	} else {
		flag = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	}

	fileObj, err := os.OpenFile(name, flag, 0644)
	if err != nil {
		log.Printf("Failed to open the file\n", err.Error())
		//os.Exit(2)
		return false
	}
	defer fileObj.Close()
	if _, err := fileObj.Write(content); err != nil {
		log.Printf("Failed to write content to the file\n", err.Error())
		return false
	}
	return true
}

//Delete 删除文件
func Delete(file string) bool {
	err := os.Remove(file)
	if err != nil {
		return false
	} else {
		return true
	}
}

// TODO: implement those functions
func isReadable(mode os.FileMode) bool { return mode&0400 != 0 }

func isWritable(mode os.FileMode) bool { return mode&0200 != 0 }

func isExecutable(mode os.FileMode) bool { return mode&0100 != 0 }
