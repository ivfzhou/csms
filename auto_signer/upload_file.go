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
	"sync/atomic"
	"time"

	bp "gitee.com/ivfzhou/csms/backend/protocol"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	gu "gitee.com/ivfzhou/goroutine-util"
)

// UploadFile 上传文件。
func UploadFile(cfg *Configuration, token string, fileSize int64, step *StepRunner) (string, string, []string, error) {
	info := make([]string, 0, 5)
	// 获取 MD5 值。
	fileMD5, err := calculateFileMD5(cfg.Base.InFile, fileSize)
	if err != nil {
		return token, "", nil, err
	}
	info = append(info, fmt.Sprintf("文件 MD5: %s", fileMD5), fmt.Sprintf("文件大小: %s", FormatSize(fileSize)))

	// 初始化上传。
	fileID, token, exist, err := initializeUploading(cfg, token, fileSize, fileMD5, filepath.Base(cfg.Base.InFile))
	if err != nil {
		return token, "", nil, err
	}
	if exist {
		info = append(info, fmt.Sprintf("文件已存在: %s", fileID))
		return token, fileID, info, nil
	}

	// 上传文件分片。
	startTime := time.Now()
	token, fileChunksInfo, err := uploadFileParts(cfg, token, fileID, fileSize, step)
	if err != nil {
		return token, "", nil, err
	}
	info = append(info, fileChunksInfo)
	cost := time.Since(startTime)

	// 合并分片。
	token, err = mergeFileParts(cfg, token, fileID)
	if err != nil {
		return token, "", nil, err
	}
	info = append(info, fmt.Sprintf("文件 ID: %s", fileID))
	info = append(info, fmt.Sprintf("上传文件平均速度: %s/s", FormatSize(int64(float64(fileSize)/cost.Seconds()))))

	return token, fileID, info, nil
}

func calculateFileMD5(filePath string, fileSize int64) (string, error) {
	hash := md5.New()
	fileStream, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("请检查输入文件：%v", err)
	}
	defer func() {
		if err = fileStream.Close(); err != nil {
			log.Printf("failed to close file %s: %v\n", filePath, err)
		}
	}()

	written, err := io.Copy(hash, fileStream)
	if err != nil {
		return "", fmt.Errorf("计算输入文件 MD5 失败：%v", err)
	}
	if written != fileSize {
		return "", fmt.Errorf("计算输入文件MD5 失败，未能完整读取文件")
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func initializeUploading(cfg *Configuration, token string, fileSize int64, fileMD5 string, fileName string) (
	string, string, bool, error) {

	retryTimes := HTTPRetryTimes

Do:
	// 构建请求参数。
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
		return "", token, false, fmt.Errorf("请检查任务类型：%s", cfg.Base.JobType)
	}
	reqBodyBytes, _ := json.Marshal(&bp.FileAPIInitialReq{
		Name: fileName,
		Size: fileSize,
		MD5:  fileMD5,
		Type: typ,
	})
	reqURL := fmt.Sprintf("%s/%s", strings.TrimRight(ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathFileAPIInitializeUploadFile))
	request, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return "", token, false, fmt.Errorf("创建 HTTP 请求失败：%v", err)
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", token)
	request.ContentLength = int64(len(reqBodyBytes))

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", token, false, fmt.Errorf("发送 HTTP 请求失败：%v", err)
	}
	if response == nil {
		retryTimes--
		if retryTimes < 0 {
			return "", token, false, fmt.Errorf("空的 HTTP 响应体")
		}
		time.Sleep(time.Second)
		goto Do
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusBadRequest, http.StatusForbidden, http.StatusInternalServerError:
		result := ReadAndUnmarshal[util.Response[bp.FileAPIInitialRsp]](response.Body)
		return "", token, false, fmt.Errorf("初始化文件上传失败：%d %s %s",
			result.Code, result.Message, response.Header.Get(cc.HTTPHeaderRequestID))
	case http.StatusUnauthorized:
		result := ReadAndUnmarshal[util.Response[bp.FileAPIInitialRsp]](response.Body)
		token, _ = CreateAuthorization(cfg)
		retryTimes--
		if retryTimes < 0 {
			return "", token, false, fmt.Errorf("初始化文件上传失败，请检查凭证是否有权限：%d %s %s",
				result.Code, result.Message, response.Header.Get(cc.HTTPHeaderRequestID))
		}
		time.Sleep(time.Second)
		goto Do
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		time.Sleep(time.Second)
		goto Do
	case http.StatusOK:
		result := ReadAndUnmarshal[util.Response[bp.FileAPIInitialRsp]](response.Body)
		return result.Data.FileID, token, result.Data.Exist, nil
	default:
		retryTimes--
		if retryTimes < 0 {
			return "", token, false, fmt.Errorf("初始化文件上传失败，响应信息：%s %s %s",
				response.Status, response.Header.Get(cc.HTTPHeaderRequestID), string(ReadAndClose(response.Body)))
		}
		CloseIO(response.Body)
		time.Sleep(time.Second)
		token, _ = CreateAuthorization(cfg)
		goto Do
	}
}

func uploadFileParts(cfg *Configuration, token string, fileID string, fileSize int64, step *StepRunner) (string, string, error) {
	// 处理请求数据。
	fileStream, err := os.Open(cfg.Base.InFile)
	if err != nil {
		return token, "", fmt.Errorf("读取输入文件失败，请检查文件路径：%v", err)
	}
	defer func() {
		if err = fileStream.Close(); err != nil {
			log.Printf("failed to close file %s: %v\n", cfg.Base.InFile, err)
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reqURL := fmt.Sprintf("%s/%s", strings.TrimRight(ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathFileAPIUploadFilePart))
	token, _ = CreateAuthorization(cfg)

	// 按分片读取文件，同时计数总块数。
	type Data struct {
		Bytes  []byte
		Number int
		Token  string
	}
	var allParts []Data
	number := 1
	finishReading := true
	totalReadBytes := int64(0)
	beginTime := time.Now()
	for finishReading {
		buf := make([]byte, UploadFilePartSize)
		n, err2 := io.ReadFull(fileStream, buf)
		if err2 != nil {
			if errors.Is(err2, io.ErrUnexpectedEOF) {
				buf = buf[:n]
				finishReading = false
			} else if errors.Is(err2, io.EOF) {
				break
			} else {
				return token, "", fmt.Errorf("读取输入文件失败，请检查文件路径：%v", err)
			}
		}
		totalReadBytes += int64(len(buf))
		if time.Since(beginTime) > AccessTokenExpiredDuration-time.Minute {
			beginTime = time.Now()
			token, _ = CreateAuthorization(cfg)
		}
		allParts = append(allParts, Data{buf, number, token})
		number++
	}

	totalParts := int64(len(allParts))

	// 并发上传计数器。
	var completedParts int64

	// 启动进度报告协程。
	progressDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				done := atomic.LoadInt64(&completedParts)
				step.UpdateRunning(fmt.Sprintf("上传中... %d/%d 分片", done, totalParts))
			case <-progressDone:
				return
			}
		}
	}()

	add, wait := gu.NewRunner(ctx, runtime.NumCPU(), func(ctx context.Context, t *Data) error {
		retryTimes := HTTPRetryTimes

	Do:
		// 构建请求体。
		buf := &bytes.Buffer{}
		writer := multipart.NewWriter(buf)
		err2 := writer.WriteField("fileId", fileID)
		if err2 != nil {
			return err2
		}
		if err2 = writer.WriteField("chunkNumber", strconv.Itoa(t.Number)); err2 != nil {
			return err2
		}
		w, err2 := writer.CreateFormFile("chunk", "chunk")
		if err2 != nil {
			return err2
		}
		n, err2 := w.Write(t.Bytes)
		if err2 != nil {
			return err2
		}
		if n != len(t.Bytes) {
			return fmt.Errorf("failed to create multipart: %d != %d", n, len(t.Bytes))
		}
		if err2 = writer.Close(); err2 != nil {
			return err2
		}

		// 发送请求。
		request, err2 := http.NewRequest(http.MethodPost, reqURL, buf)
		if err2 != nil {
			return err2
		}
		request.Header.Set("Content-Type", writer.FormDataContentType())
		request.Header.Set("Authorization", t.Token)
		request.ContentLength = int64(buf.Len())
		response, err2 := http.DefaultClient.Do(request)
		if err2 != nil {
			return err2
		}
		if response == nil {
			retryTimes--
			if retryTimes < 0 {
				return errors.New("nil http response")
			}
			time.Sleep(time.Second)
			goto Do
		}

		// 处理结果。
		switch response.StatusCode {
		case http.StatusTooManyRequests:
			CloseIO(response.Body)
			time.Sleep(time.Second)
			goto Do
		case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden, http.StatusRequestTimeout:
			result := ReadAndUnmarshal[util.Response[any]](response.Body)
			return fmt.Errorf("%d %s %s", result.Code, response.Header.Get(cc.HTTPHeaderRequestID), result.Message)
		case http.StatusUnauthorized:
			result := ReadAndUnmarshal[util.Response[any]](response.Body)
			token, _ = CreateAuthorization(cfg)
			retryTimes--
			if retryTimes < 0 {
				return fmt.Errorf("access token is invalid %v %v", result.Code, result.Message)
			}
			time.Sleep(time.Second)
			goto Do
		case http.StatusNoContent:
			atomic.AddInt64(&completedParts, 1)
			return nil
		default:
			retryTimes--
			if retryTimes < 0 {
				return fmt.Errorf("http status is invalid %s %s", response.Status, string(ReadAndClose(response.Body)))
			}
			time.Sleep(time.Second)
			goto Do
		}
	})

	// 提交所有任务。
	for i := range allParts {
		if err2 := add(&allParts[i], true); err2 != nil {
			close(progressDone)
			return token, "", fmt.Errorf("上传文件失败：%v", err2)
		}
	}

	// 等待上传完毕。
	if err = wait(false); err != nil {
		close(progressDone)
		return token, "", fmt.Errorf("上传文件失败：%v", err)
	}
	close(progressDone)

	// 最终进度刷新。
	step.UpdateRunning(fmt.Sprintf("上传中... %d/%d 分片", totalParts, totalParts))

	// 校验字节数。
	if totalReadBytes != fileSize {
		return token, "", fmt.Errorf("上传文件失败，上传字节数不相等：%v!=%v", fileSize, totalReadBytes)
	}

	token, _ = CreateAuthorization(cfg)
	return token, fmt.Sprintf("上传分片数量 %d，分片大小 %s", completedParts, FormatSize(UploadFilePartSize)), nil
}

func mergeFileParts(cfg *Configuration, token string, fileID string) (string, error) {
	retryTimes := HTTPRetryTimes

Do:
	// 构建请求体。
	query := util.EncodeStructToURLQuery(&bp.FileAPIMergePartsReq{FileID: fileID})
	reqURL := fmt.Sprintf("%s/%s?%s", strings.TrimRight(ServerAddress, "/"),
		path.Join(cc.ServiceNameBackend, bp.HTTPPathFileAPIMergeFilePart), query)
	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return token, fmt.Errorf("创建 HTTP 请求失败：%v", err)
	}
	request.Header.Set("Authorization", token)

	// 发送请求。
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return token, fmt.Errorf("发送 HTTP 请求失败：%v", err)
	}
	if response == nil {
		retryTimes--
		if retryTimes < 0 {
			return token, fmt.Errorf("空的 HTTP 响应体")
		}
		time.Sleep(time.Second)
		goto Do
	}

	// 处理结果。
	switch response.StatusCode {
	case http.StatusTooManyRequests:
		CloseIO(response.Body)
		time.Sleep(time.Second)
		goto Do
	case http.StatusBadRequest, http.StatusInternalServerError, http.StatusForbidden, http.StatusRequestTimeout:
		result := ReadAndUnmarshal[util.Response[any]](response.Body)
		return token, fmt.Errorf("合并分片失败：%d %s %s", result.Code, result.Message, response.Header.Get(cc.HTTPHeaderRequestID))
	case http.StatusUnauthorized:
		result := ReadAndUnmarshal[util.Response[any]](response.Body)
		token, _ = CreateAuthorization(cfg)
		retryTimes--
		if retryTimes < 0 {
			return token, fmt.Errorf("合并分片失败，请检查凭证权限：%v %v %v", result.Code, result.Message, response.Header.Get(cc.HTTPHeaderRequestID))
		}
		time.Sleep(time.Second)
		goto Do
	case http.StatusNoContent:
		CloseIO(response.Body)
		return token, nil
	default:
		retryTimes--
		if retryTimes < 0 {
			return token, fmt.Errorf("合并分片失败，响应信息：%s %s %s",
				response.Status, response.Header.Get(cc.HTTPHeaderRequestID), string(ReadAndClose(response.Body)))
		}
		CloseIO(response.Body)
		time.Sleep(time.Second)
		goto Do
	}
}
