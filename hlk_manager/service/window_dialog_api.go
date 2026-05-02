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

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/hlk_manager/consts"
)

type pythonHandlerRsp[E any] struct {
	Result  bool   `json:"result"`
	Message string `json:"message"`
	Content E      `json:"content"`
}

func listAllWindows(_ context.Context) ([]string, error) {
	// 请求 python。
	err := askPython(`{"method": "list_window"}`)
	if err != nil {
		return nil, err
	}
	time.Sleep(10 * time.Second)
	content, err := answerPython()
	if err != nil {
		return nil, err
	}

	// 反序列化结果。
	var result pythonHandlerRsp[[]string]
	if err = json.Unmarshal([]byte(content), &result); err != nil {
		return nil, err
	}
	if !result.Result {
		return nil, errors.New(result.Message)
	}

	return result.Content, nil
}

func closeWindow(_ context.Context, title string, index int) error {
	// 请求 python。
	err := askPython(fmt.Sprintf(`{"method": "close_window", "title": "%s", "index": %d}`, title, index))
	if err != nil {
		return err
	}
	time.Sleep(10 * time.Second)
	content, err := answerPython()
	if err != nil {
		return err
	}

	// 反序列化结果。
	var res pythonHandlerRsp[any]
	if err = json.Unmarshal([]byte(content), &res); err != nil {
		return err
	}
	if !res.Result {
		return errors.New(res.Message)
	}

	return nil
}

func clickButton(_ context.Context, title, button string, index int) error {
	// 请求 python。
	req := fmt.Sprintf(`{"method": "click_button", "title": "%s", "button": "%s", "index": %d}`, title, button, index)
	err := askPython(req)
	if err != nil {
		return err
	}
	time.Sleep(10 * time.Second)
	content, err := answerPython()
	if err != nil {
		return err
	}

	// 反序列化结果。
	var result pythonHandlerRsp[string]
	if err = json.Unmarshal([]byte(content), &result); err != nil {
		return err
	}
	if !result.Result {
		return errors.New(result.Message)
	}

	return nil
}

func getDialogs(_ context.Context, title string) (map[string]string, error) {
	// 请求 python。
	req := fmt.Sprintf(`{"method": "get_dialog", "title": "%s"}`, title)
	err := askPython(req)
	if err != nil {
		return nil, err
	}
	time.Sleep(10 * time.Second)
	content, err := answerPython()
	if err != nil {
		return nil, err
	}

	// 反序列化结果。
	var result pythonHandlerRsp[map[string]string]
	if err = json.Unmarshal([]byte(content), &result); err != nil {
		return nil, err
	}
	if !result.Result {
		return nil, errors.New(result.Message)
	}

	return result.Content, nil
}

func getCmdContent(_ context.Context, index int) (string, error) {
	// 请求 python。
	req := fmt.Sprintf(`{"method": "get_cmd_content", "index": %d}`, index)
	err := askPython(req)
	if err != nil {
		return "", err
	}
	time.Sleep(10 * time.Second)
	content, err := answerPython()
	if err != nil {
		return "", err
	}

	// 反序列化结果。
	var result pythonHandlerRsp[string]
	if err = json.Unmarshal([]byte(content), &result); err != nil {
		return "", err
	}
	if !result.Result {
		return "", errors.New(result.Message)
	}

	return result.Content, nil
}

func sendCmdContent(_ context.Context, cmd string, index int) error {
	// 请求 python。
	req := fmt.Sprintf(`{"method": "send_cmd_content", "content": "%s", "index": %d}`, cmd, index)
	err := askPython(req)
	if err != nil {
		return err
	}
	time.Sleep(10 * time.Second)
	content, err := answerPython()
	if err != nil {
		return err
	}

	// 反序列化结果。
	var result pythonHandlerRsp[string]
	if err = json.Unmarshal([]byte(content), &result); err != nil {
		return err
	}
	if !result.Result {
		return errors.New(result.Message)
	}

	return nil
}

// 向文件写请求
func askPython(content string) error {
	err := os.MkdirAll(filepath.Dir(consts.DialogConversationFilePath), cc.FileMode)
	if err != nil {
		return err
	}
	err = os.WriteFile(consts.DialogConversationFilePath, []byte("ask: "+content), cc.FileMode)
	if err != nil {
		time.Sleep(3 * time.Second)
		err = os.WriteFile(consts.DialogConversationFilePath, []byte(content), cc.FileMode)
	}
	return err
}

// 向文件读取响应
func answerPython() (string, error) {
	file, err := os.ReadFile(consts.DialogConversationFilePath)
	if err != nil {
		time.Sleep(3 * time.Second)
		file, err = os.ReadFile(consts.DialogConversationFilePath)
	}
	if err != nil {
		return "", err
	}
	prefix := []byte("answer: ")
	for v := range bytes.SplitSeq(file, []byte{'\n'}) {
		if bytes.HasPrefix(v, prefix) {
			return string(v[len(prefix):]), nil
		}
	}
	return "", errors.New("answer not found")
}
