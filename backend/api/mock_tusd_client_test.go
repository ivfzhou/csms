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

package api_test

import (
	"context"
	"io"
	"os"
	"sync"

	"gitee.com/ivfzhou/csms/comm/conn"
	tus "gitee.com/ivfzhou/tus_client/v2"
)

const (
	tusUploadFile tusdMethod = 1 + iota
	tusGet
	tusDeleteFile
	tusUploadPartByIO
	tusMergeParts
	tusDownloadToFile
)

type TusdMocker interface {
	UploadFileOnce(l string, e error) TusdMocker
	GetOnce(result *tus.GetResult, e error) TusdMocker
	DeleteFileOnce(e error) TusdMocker
	UploadPartByIOOnce(id string, e error) TusdMocker
	MergePartsOnce(id string, e error) TusdMocker
	DownloadToFileOnce(data []byte, e error) TusdMocker
	Reset()
}

type tusdMockerImpl struct {
	datasLock sync.Mutex
	datas     []*tusdResultData
	reset     func()
}

type tusdResultData struct {
	fn     tusdMethod
	result []any
}

type tusdMethod int

func MockTusdClient(ctx context.Context) TusdMocker {
	m := &tusdMockerImpl{}
	client := conn.TusdClient(ctx)
	conn.MockTusdClient(m)
	m.reset = func() { conn.MockTusdClient(client) }
	return m
}

func (c *tusdMockerImpl) UploadFileOnce(l string, e error) TusdMocker {
	c.datas = append(c.datas, &tusdResultData{
		fn:     tusUploadFile,
		result: []any{l, e},
	})
	return c
}

func (c *tusdMockerImpl) GetOnce(result *tus.GetResult, e error) TusdMocker {
	c.datas = append(c.datas, &tusdResultData{
		fn:     tusGet,
		result: []any{result, e},
	})
	return c
}

func (c *tusdMockerImpl) DeleteFileOnce(e error) TusdMocker {
	c.datas = append(c.datas, &tusdResultData{
		fn:     tusDeleteFile,
		result: []any{e},
	})
	return c
}

func (c *tusdMockerImpl) UploadPartByIOOnce(id string, e error) TusdMocker {
	c.datas = append(c.datas, &tusdResultData{
		fn:     tusUploadPartByIO,
		result: []any{id, e},
	})
	return c
}

func (c *tusdMockerImpl) MergePartsOnce(id string, e error) TusdMocker {
	c.datas = append(c.datas, &tusdResultData{
		fn:     tusMergeParts,
		result: []any{id, e},
	})
	return c
}

func (c *tusdMockerImpl) DownloadToFileOnce(data []byte, e error) TusdMocker {
	c.datas = append(c.datas, &tusdResultData{
		fn:     tusDownloadToFile,
		result: []any{data, e},
	})
	return c
}

func (c *tusdMockerImpl) Reset() {
	c.reset()
}

func (c *tusdMockerImpl) Options(context.Context) (*tus.OptionsResult, error) {
	panic("not implement yet")
}

func (c *tusdMockerImpl) Post(context.Context, *tus.PostRequest) (*tus.PostResult, error) {
	panic("not implement yet")
}

func (c *tusdMockerImpl) Head(context.Context, *tus.HeadRequest) (*tus.HeadResult, error) {
	panic("not implement yet")
}

func (c *tusdMockerImpl) Patch(context.Context, *tus.PatchRequest) (*tus.PatchResult, error) {
	panic("not implement yet")
}

func (c *tusdMockerImpl) Delete(context.Context, *tus.DeleteRequest) (*tus.DeleteResult, error) {
	panic("not implement yet")
}

func (c *tusdMockerImpl) Get(context.Context, *tus.GetRequest) (*tus.GetResult, error) {
	data := c.getTusdData(tusGet)
	if data != nil {
		err, _ := data[1].(error)
		result, _ := data[0].(*tus.GetResult)
		return result, err
	}
	return nil, nil
}

func (c *tusdMockerImpl) MultipleUploadFromFile(context.Context, string) (string, error) {
	panic("not implement yet")
}

func (c *tusdMockerImpl) MultipleUploadFromReader(context.Context, io.Reader) (string, error) {
	panic("not implement yet")
}

func (c *tusdMockerImpl) DownloadToFile(_ context.Context, _ string, filePath string) error {
	data := c.getTusdData(tusDownloadToFile)
	if data != nil {
		err, _ := data[1].(error)
		if err != nil {
			return err
		}
		fileData, _ := data[0].([]byte)
		return os.WriteFile(filePath, fileData, 0644)
	}
	panic("not implement yet")
}

func (c *tusdMockerImpl) DownloadToWriter(context.Context, string, io.Writer) error {
	panic("not implement yet")
}

func (c *tusdMockerImpl) UploadPart(context.Context, []byte) (string, error) {
	panic("not implement yet")
}

func (c *tusdMockerImpl) MergeParts(context.Context, []string) (string, error) {
	data := c.getTusdData(tusMergeParts)
	if data != nil {
		err, _ := data[1].(error)
		return data[0].(string), err
	}
	return "", nil
}

func (c *tusdMockerImpl) DiscardParts(context.Context, []string) error { panic("not implement yet") }

func (c *tusdMockerImpl) UploadPartByIO(context.Context, io.ReadCloser, int) (string, error) {
	data := c.getTusdData(tusUploadPartByIO)
	if data != nil {
		err, _ := data[1].(error)
		return data[0].(string), err
	}
	return "", nil
}

func (c *tusdMockerImpl) PatchByIO(context.Context, *tus.PatchByIORequest) (*tus.PatchResult, error) {
	panic("not implement yet")
}

func (c *tusdMockerImpl) UploadFile(context.Context, []byte) (string, error) {
	data := c.getTusdData(tusUploadFile)
	if data != nil {
		err, _ := data[1].(error)
		return data[0].(string), err
	}
	return "", nil
}

func (c *tusdMockerImpl) DeleteFile(context.Context, string) error {
	data := c.getTusdData(tusDeleteFile)
	if data != nil {
		err, _ := data[0].(error)
		return err
	}
	return nil
}

func (c *tusdMockerImpl) getTusdData(fn tusdMethod) []any {
	if len(c.datas) <= 0 {
		return nil
	}
	c.datasLock.Lock()
	defer c.datasLock.Unlock()
	for i, data := range c.datas {
		if data.fn == fn {
			header := c.datas[:i]
			tail := c.datas[i+1:]
			c.datas = append(header, tail...)
			return data.result
		}
	}
	return nil
}
