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

package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	bp "gitee.com/ivfzhou/csms/backend/protocol"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	cl "gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	gu "gitee.com/ivfzhou/goroutine-util"
)

// UploadFile 上传文件。
func UploadFile(cfg *Configuration, token string) (string, string, bool) {
	// 检查下文件。
	fileInfo, err := os.Stat(cfg.Base.InFile)
	if err != nil {
		log.Println(cl.LevelError, "failed to get file info", err)
		return "", token, false
	}
	if fileInfo.IsDir() {
		log.Println(cl.LevelError, cfg.Base.InFile, "is a directory")
		return "", token, false
	}

	// 获取 MD5 值。
	log.Println(cl.LevelInfo, "calculate file md5")
	fileSize := fileInfo.Size()
	fileMD5, err := calculateFileMD5(cfg.Base.InFile, fileSize)
	if err != nil {
		return "", token, false
	}
	log.Println(cl.LevelInfo, "file md5 is", fileMD5)
	log.Println(cl.LevelInfo, "file size is", fileSize)

	// 初始化上传。
	log.Println(cl.LevelInfo, "initial upload file")
	fileID, token, exist, err := initializeUploading(cfg, token, fileSize, fileMD5, filepath.Base(cfg.Base.InFile))
	if err != nil {
		return "", token, false
	}
	if exist {
		log.Println(cl.LevelInfo, "file exists")
		return fileID, token, true
	}

	// 上传文件分片。
	log.Println(cl.LevelInfo, "upload file parts")
	token, err = uploadFileParts(cfg, token, fileID, fileSize)
	if err != nil {
		return "", token, false
	}

	// 合并分片。
	log.Println(cl.LevelInfo, "merge file parts")
	token, err = mergeFileParts(cfg, token, fileID)
	if err != nil {
		return "", token, false
	}

	return fileID, token, true
}

func calculateFileMD5(filePath string, fileSize int64) (string, error) {
	hash := md5.New()
	fileStream, err := os.Open(filePath)
	if err != nil {
		log.Println(cl.LevelError, "failed to open file", err)
		return "", err
	}
	defer func() {
		if err = fileStream.Close(); err != nil {
			log.Println(cl.LevelError, "failed to close file", err)
		}
	}()

	written, err := io.Copy(hash, fileStream)
	if err != nil {
		log.Println(cl.LevelError, "failed to read file", err)
		return "", err
	}
	if written != fileSize {
		log.Println(cl.LevelError, "failed to read file", written, fileSize)
		return "", errors.New("failed to read file fully")
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func initializeUploading(cfg *Configuration, token string, fileSize int64, fileMD5 string, fileName string) (
	string, string, bool, error) {

	retryTimes := HTTPRetryTimes

Do:
	// 构建请求参数，
	typ := model.FileTypeWindowsSigning
	switch cfg.Base.JobType {
	case JobTypeWHQL:
		typ = model.FileTypeWindowsSigning
	case JobTypeAndroid:
		typ = model.FileTypeAndroidSigning
	case JobTypeWindows:
		typ = model.FileTypeWindowsSigning
	case JobTypeApple:
		typ = model.FileTypeAppleSigning
	default:
		log.Println(cl.LevelError, "invalid job type", cfg.Base.JobType)
		return "", token, false, fmt.Errorf("invalid job type: %s", cfg.Base.JobType)
	}
	reqBodyBytes, _ := json.Marshal(&bp.FileAPIInitialReq{
		Name: fileName,
		Size: fileSize,
		MD5:  fileMD5,
		Type: typ,
	})
	reqURL := fmt.Sprintf("%s/%s", strings.TrimRight(cfg.Base.ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathFileAPIInitializeUploadFile))
	request, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(reqBodyBytes))
	if err != nil {
		log.Println(cl.LevelError, "failed to create http request", err)
		return "", token, false, err
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", token)
	request.ContentLength = int64(len(reqBodyBytes))

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println(cl.LevelError, "failed to do http request", err)
		return "", token, false, err
	}
	if response == nil {
		retryTimes--
		if retryTimes < 0 {
			return "", token, false, errors.New("nil http response")
		}
		log.Println(cl.LevelWarn, "try initializing file uploading again")
		time.Sleep(time.Second)
		goto Do
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusForbidden, http.StatusInternalServerError:
		result := ReadAndUnmarshal[util.Response[bp.FileAPIInitialRsp]](response.Body)
		log.Println(cl.LevelError, "failed to initialize file uploading", result.Code, result.Message,
			response.Header.Get(cc.HTTPHeaderRequestID))
		return "", token, false, fmt.Errorf("%d %s", result.Code, result.Message)
	case http.StatusUnauthorized:
		result := ReadAndUnmarshal[util.Response[bp.FileAPIInitialRsp]](response.Body)
		log.Println(cl.LevelWarn, "access token is invalid", result.Code, result.Message)
		token, _ = CreateAuthorization(cfg)
		retryTimes--
		if retryTimes < 0 {
			return "", token, false, fmt.Errorf("%d %s", result.Code, result.Message)
		}
		log.Println(cl.LevelWarn, "try initializing file uploading again")
		time.Sleep(time.Second)
		goto Do
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		log.Println(cl.LevelWarn, "rate limit reached, try initializing file uploading again")
		time.Sleep(time.Second)
		goto Do
	case http.StatusOK:
		result := ReadAndUnmarshal[util.Response[bp.FileAPIInitialRsp]](response.Body)
		return result.Data.FileID, token, result.Data.Exist, nil
	default:
		log.Println(cl.LevelError, "invalid response status", response.Status, string(ReadAndClose(response.Body)))
		retryTimes--
		if retryTimes < 0 {
			return "", token, false, fmt.Errorf("invalid response status %s", response.Status)
		}
		log.Println(cl.LevelWarn, "try initializing file uploading again")
		time.Sleep(time.Second)
		token, _ = CreateAuthorization(cfg)
		goto Do
	}
}

func uploadFileParts(cfg *Configuration, token string, fileID string, fileSize int64) (string, error) {
	// 处理请求数据。
	fileStream, err := os.Open(cfg.Base.InFile)
	if err != nil {
		log.Println(cl.LevelError, "failed to open file", err)
		return token, err
	}
	defer func() {
		if err = fileStream.Close(); err != nil {
			log.Println(cl.LevelError, "failed to close file", err)
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reqURL := fmt.Sprintf("%s/%s", strings.TrimRight(cfg.Base.ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathFileAPIUploadFilePart))
	token, _ = CreateAuthorization(cfg)

	// 上传文件分片。
	type Data struct {
		Bytes  []byte
		Number int
		Token  string
	}
	add, wait := gu.NewRunner(ctx, runtime.NumCPU(), func(ctx context.Context, t *Data) error {
		log.Println(cl.LevelInfo, "upload file chunk", t.Number, len(t.Bytes))

		retryTimes := HTTPRetryTimes

	Do:
		// 构建请求体。
		buf := &bytes.Buffer{}
		writer := multipart.NewWriter(buf)
		err2 := writer.WriteField("fileId", fileID)
		if err2 != nil {
			log.Println(cl.LevelError, "failed to create multipart", err2)
			return err2
		}
		if err2 = writer.WriteField("chunkNumber", strconv.Itoa(t.Number)); err2 != nil {
			log.Println(cl.LevelError, "failed to create multipart", err2)
			return err2
		}
		w, err2 := writer.CreateFormFile("chunk", "chunk")
		if err2 != nil {
			log.Println(cl.LevelError, "failed to create multipart", err2)
			return err2
		}
		n, err2 := w.Write(t.Bytes)
		if err2 != nil {
			log.Println(cl.LevelError, "failed to create multipart", err2)
			return err2
		}
		if n != len(t.Bytes) {
			log.Println(cl.LevelError, "failed to create multipart", n, len(t.Bytes))
			return err2
		}
		if err2 = writer.Close(); err2 != nil {
			log.Println(cl.LevelError, "failed to close multipart writer", err2)
			return err2
		}

		// 发送请求。
		request, err2 := http.NewRequest(http.MethodPost, reqURL, buf)
		if err2 != nil {
			log.Println(cl.LevelError, "failed to create http request", err2)
			return err2
		}
		request.Header.Set("Content-Type", writer.FormDataContentType())
		request.Header.Set("Authorization", t.Token)
		request.ContentLength = int64(buf.Len())
		response, err2 := http.DefaultClient.Do(request)
		if err2 != nil {
			log.Println(cl.LevelError, "failed to do http request", err2)
			return err2
		}
		if response == nil {
			retryTimes--
			if retryTimes < 0 {
				return errors.New("nil http response")
			}
			log.Println(cl.LevelWarn, "try uploading file part again")
			time.Sleep(time.Second)
			goto Do
		}

		// 处理结果。
		switch response.StatusCode {
		case http.StatusTooManyRequests:
			CloseIO(response.Body)
			log.Println(cl.LevelWarn, "rate limit reached, try uploading file part again")
			time.Sleep(time.Second)
			goto Do
		case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden, http.StatusRequestTimeout:
			result := ReadAndUnmarshal[util.Response[any]](response.Body)
			log.Println(cl.LevelError, "failed to upload file part", result.Code, result.Message,
				response.Header.Get(cc.HTTPHeaderRequestID))
			return fmt.Errorf("%d %s", result.Code, result.Message)
		case http.StatusUnauthorized:
			result := ReadAndUnmarshal[util.Response[any]](response.Body)
			log.Println(cl.LevelWarn, "access token is invalid", result.Code, result.Message)
			token, _ = CreateAuthorization(cfg)
			retryTimes--
			if retryTimes < 0 {
				return fmt.Errorf("access token is invalid %v %v", result.Code, result.Message)
			}
			log.Println(cl.LevelWarn, "try uploading file part again")
			time.Sleep(time.Second)
			goto Do
		case http.StatusNoContent:
			return nil
		default:
			log.Println(cl.LevelError, "invalid response status", response.Status, string(ReadAndClose(response.Body)))
			retryTimes--
			if retryTimes < 0 {
				return fmt.Errorf("http status is invalid %s", response.Status)
			}
			log.Println(cl.LevelWarn, "try uploading file part again")
			time.Sleep(time.Second)
			goto Do
		}
	})

	// 读取文件。
	number := 1
	finishReading := true
	totalReadBytes := int64(0)
	beginTime := time.Now()
	for finishReading {
		buf := make([]byte, UploadFilePartSize)
		n, err2 := io.ReadFull(fileStream, buf)
		if err2 != nil {
			if errors.Is(err2, io.ErrUnexpectedEOF) { // 读完了，还剩一点。
				buf = buf[:n]
				finishReading = false
			} else if errors.Is(err2, io.EOF) { // 读完了。
				break
			} else {
				log.Println(cl.LevelError, "failed to read file", err2)
				cancel()
				return token, err2
			}
		}
		totalReadBytes += int64(len(buf))
		if time.Since(beginTime) > AccessTokenExpiredDuration-time.Minute {
			beginTime = time.Now()
			token, _ = CreateAuthorization(cfg)
		}
		if err2 = add(&Data{buf, number, token}, true); err2 != nil {
			log.Println(cl.LevelError, "failed to upload file", err2)
			return token, err2
		}
		number++
	}

	// 等待上传完毕。
	if err = wait(false); err != nil {
		log.Println(cl.LevelError, "failed to upload file", err)
		return token, err
	}

	// 校验字节数。
	if totalReadBytes != fileSize {
		log.Println(cl.LevelError, "the number of bytes uploaded is different from the number of bytes in the file",
			fileSize, totalReadBytes)
		return token, fmt.Errorf("the number of bytes uploaded is different from the number of bytes in the file %v %v",
			fileSize, totalReadBytes)
	}

	token, _ = CreateAuthorization(cfg)
	return token, nil
}

func mergeFileParts(cfg *Configuration, token string, fileID string) (string, error) {
	retryTimes := HTTPRetryTimes

Do:
	// 构建请求体。
	query := util.EncodeStructToURLQuery(&bp.FileAPIMergePartsReq{FileID: fileID})
	reqURL := fmt.Sprintf("%s/%s?%s", strings.TrimRight(cfg.Base.ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathFileAPIMergeFilePart), query)
	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		log.Println(cl.LevelError, "failed to create http request", err)
		return token, err
	}
	request.Header.Set("Authorization", token)

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println(cl.LevelError, "failed to do http request", err)
		return token, err
	}
	if response == nil {
		retryTimes--
		if retryTimes < 0 {
			return token, errors.New("nil http response")
		}
		log.Println(cl.LevelWarn, "try merging file again")
		time.Sleep(time.Second)
		goto Do
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		log.Println(cl.LevelWarn, "rate limit reached, try merging file again")
		time.Sleep(time.Second)
		goto Do
	case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden, http.StatusRequestTimeout:
		result := ReadAndUnmarshal[util.Response[any]](response.Body)
		log.Println(cl.LevelError, "failed to merge file", result.Code, result.Message,
			response.Header.Get(cc.HTTPHeaderRequestID))
		return token, fmt.Errorf("%d %s", result.Code, result.Message)
	case http.StatusUnauthorized:
		result := ReadAndUnmarshal[util.Response[any]](response.Body)
		log.Println(cl.LevelWarn, "access token is invalid", result.Code, result.Message)
		token, _ = CreateAuthorization(cfg)
		retryTimes--
		if retryTimes < 0 {
			return token, fmt.Errorf("access token is invalid %v %v", result.Code, result.Message)
		}
		log.Println(cl.LevelWarn, "try merging file again")
		time.Sleep(time.Second)
		goto Do
	case http.StatusNoContent:
		CloseIO(response.Body)
		return token, nil
	default:
		log.Println(cl.LevelError, "invalid response status", response.Status, string(ReadAndClose(response.Body)))
		retryTimes--
		if retryTimes < 0 {
			return token, fmt.Errorf("http code is invalid %s", response.Status)
		}
		log.Println(cl.LevelWarn, "try merging file again")
		time.Sleep(time.Second)
		goto Do
	}
}
