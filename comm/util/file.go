/*
 * Copyright (c) 2024 ivfzhou
 * csms is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package util

import (
	"archive/zip"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/log"
	gu "gitee.com/ivfzhou/goroutine-util"
)

// ErrNoFilePathAvailable 创建临时文件路径时，无可用文件，
var ErrNoFilePathAvailable = errors.New("no available temporary file")

// CAB 文件头结构
type cabHeader struct {
	Signature     [4]byte // "MSCF"
	_             uint32
	CabinetSize   uint32
	_             uint32
	FilesOffset   uint32
	_             uint32
	VersionMinor  uint8
	VersionMajor  uint8
	FolderCount   uint16
	FileCount     uint16
	Flags         uint16
	SetID         uint16
	CabinetNumber uint16
}

// CreateTemporaryFile 创建临时文件。
func CreateTemporaryFile(ctx context.Context, directory, fileNamePattern string, fileData []byte) (string, error) {
	directory = filepath.Join(os.TempDir(), consts.SystemName, directory)
	err := os.MkdirAll(directory, consts.DirectoryMode)
	if err != nil {
		return "", err
	}
	fileObj, err := os.CreateTemp(directory, fileNamePattern)
	if err != nil {
		return "", err
	}
	defer func() { log.ErrorIf(ctx, fileObj.Close(), "failed to close file") }()
	if n, err2 := fileObj.Write(fileData); err2 != nil {
		return "", err2
	} else if n != len(fileData) {
		return "", io.ErrShortWrite
	}
	return fileObj.Name(), nil
}

// GenerateTemporaryFile 生成一个临时文件路径。pattern 中的 * 会替换成随机字符串。
func GenerateTemporaryFile(directory, pattern string) (string, error) {
	directory = filepath.Join(os.TempDir(), consts.SystemName, directory)
	err := os.MkdirAll(directory, consts.DirectoryMode)
	if err != nil {
		return "", err
	}
	for range 3 {
		for strings.Contains(pattern, "*") {
			pattern = strings.ReplaceAll(pattern, "*", FastRandomAlphaNumberString(32))
		}
		if len(pattern) <= 0 {
			pattern = FastRandomAlphaNumberString(32)
		}
		filePath := filepath.Join(directory, pattern)
		if _, err = os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
			return filePath, nil
		}
	}
	return "", ErrNoFilePathAvailable
}

// GenerateTemporaryDirectory 生成一个临时文件夹。
func GenerateTemporaryDirectory(directory string) (string, error) {
	directory = filepath.Join(os.TempDir(), consts.SystemName, directory)
	err := os.MkdirAll(directory, consts.DirectoryMode)
	if err != nil {
		return "", err
	}
	return directory, nil
}

// RemoveFile 删除文件。
func RemoveFile(ctx context.Context, filePath string) {
	log.ErrorIf(ctx, os.Remove(filePath), "remove file failed")
}

// RemoveDirectory 删除文件夹。
func RemoveDirectory(ctx context.Context, filePath string) {
	log.ErrorIf(ctx, os.RemoveAll(filePath), "remove directory failed")
}

// IsFileInCabFormat 文件是否是 CAB 文件类型。
func IsFileInCabFormat(ctx context.Context, cabFilePath string) (bool, error) {
	fileObj, err := os.Open(cabFilePath)
	if err != nil {
		return false, err
	}
	defer CloseIO(ctx, fileObj)

	// 读取 CAB 文件头。
	var header cabHeader
	if err = binary.Read(fileObj, binary.LittleEndian, &header); err != nil {
		return false, err
	}

	// 检查 CAB 文件签名。
	if string(header.Signature[:]) != "MSCF" {
		return false, nil
	}

	return true, nil
}

// Unzip 解压文件到目录下。
func Unzip(ctx context.Context, zipFilePath string, destinationDirectoryPath string) (err error) {
	if err = os.MkdirAll(destinationDirectoryPath, consts.DirectoryMode); err != nil {
		return
	}

	reader, err := zip.OpenReader(zipFilePath)
	defer CloseIO(ctx, reader)
	if err != nil {
		return
	}

	add, wait := gu.NewRunner(ctx, 0, func(ctx context.Context, v *zip.File) (err error) {
		r, err := v.Open()
		if err != nil {
			return
		}
		defer CloseIO(ctx, r)

		name := filepath.Join(destinationDirectoryPath, v.FileHeader.Name)
		if err = os.MkdirAll(filepath.Dir(name), consts.DirectoryMode); err != nil {
			return
		}
		fileStream, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, v.Mode().Perm())
		if err != nil {
			return
		}
		defer CloseIO(ctx, fileStream)

		written, err := io.Copy(fileStream, r)
		if v.FileInfo().Size() != written {
			return fmt.Errorf("the number of bytes written to the file is inconsistent, %d != %d",
				v.FileInfo().Size(), written)
		}
		return
	})
	for _, v := range reader.File {
		if err = add(v, false); err != nil {
			return
		}
	}

	return wait(false)
}
