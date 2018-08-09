/*
 * Copyright 2018 The OpenWallet Authors
 * This file is part of the OpenWallet library.
 *
 * The OpenWallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenWallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package walletnode

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"docker.io/go-docker/api/types"
)

// Copy file from container to local filesystem
//
//	src/dst: filename
func CopyFromContainer(symbol, src, dst string) error {
	// func(vals ...interface{}) {}(
	// 	tar.Writer{}, os.File{},
	// ) // Delete before commit

	var buf bytes.Buffer

	if err := loadConfig(symbol); err != nil {
		return err
	}

	// Init docker client
	c, err := _GetClient()
	if err != nil {
		return err
	}

	cname := strings.ToLower(symbol)
	// API Return: CopyFromContainer -> (io.ReadCloser, types.ContainerPathStat, error)
	fp, _, err := c.CopyFromContainer(context.Background(), cname, src)
	if err != nil {
		return err
	}
	defer fp.Close()

	tw := tar.NewReader(fp)
	for {
		// Copy file from container return within archive/tar
		_, err := tw.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			log.Fatal(err)
		}
		// log.Printf("Contents of %s:\n", hdr.Name)

		if _, err := buf.ReadFrom(tw); err != nil {
			log.Fatal(err)
			return err
		}
	}

	if err = ioutil.WriteFile(dst, buf.Bytes(), 0600); err != nil {
		return err
	}

	return nil
}

// Copy file to container from local filesystem
//
//	src: filename
//	dst: path
func CopyToContainer(symbol, src, dst string) error {

	var content io.Reader

	if err := loadConfig(symbol); err != nil {
		return err
	}

	// Init docker client
	c, err := _GetClient()
	if err != nil {
		return err
	}

	cname := strings.ToLower(symbol)
	// Return: ioutil.ReadFile() -> ([]byte, err)
	if dat, err := ioutil.ReadFile(src); err != nil {
		log.Println(err)
		return err
	} else {
		// Copy file into container within archive/tar
		var buf bytes.Buffer
		tw := tar.NewWriter(&buf)
		tw.WriteHeader(&tar.Header{
			Name: filepath.Base(src), //file.Name,
			Mode: 0600,
			Size: int64(len(dat)), //int64(len(file.Body)),
		})
		tw.Write([]byte(dat))
		tw.Close()

		// Transform tar to []byte as Reader for Docker API
		content = bytes.NewReader(buf.Bytes())
	}

	// API Params: (ctx context.Context, container, path string, content io.Reader, options types.CopyToContainerOptions)
	if err := c.CopyToContainer(context.Background(), cname, dst, content, types.CopyToContainerOptions{AllowOverwriteDirWithFile: false}); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
