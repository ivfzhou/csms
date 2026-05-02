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
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"net/http"
	"net/url"
	"os/exec"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/log"
)

// Number 数字类型。
type Number interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 |
		float32 | float64 |
		uintptr
}

type CreateWindowsFirewallRuleParams struct {
	Name       string
	Protocol   string
	Direct     string
	LocalIP    string
	RemoteIP   string
	Action     string
	Program    string
	LocalPort  string
	RemotePort string
}

type SafeWriter struct {
	Writer io.Writer
	lock   sync.Mutex
}

// GetStackCallers 获取本模块中的调用堆栈。
func GetStackCallers() []string {
	callers := make([]uintptr, 32)
	n := runtime.Callers(3, callers)
	callers = callers[:n]
	frames := runtime.CallersFrames(callers)
	elems := make([]string, 0, len(callers))
	for {
		frame, more := frames.Next()
		index := strings.LastIndex(frame.File, consts.SystemName)
		if index != -1 {
			elems = append(elems, fmt.Sprintf("%s:%d",
				strings.Trim(frame.File[index+len(consts.SystemName):], "/"), frame.Line))
		}

		if !more {
			break
		}
	}
	return elems
}

// FastRandomAlphaNumberString 生成随机字符串（数字加字母组合）。
func FastRandomAlphaNumberString(length int) string {
	if length <= 0 {
		return ""
	}
	str := make([]byte, length)
	for i := range length {
		switch rand.Intn(3) {
		case 0:
			str[i] = '0' + byte(rand.Intn('9'-'0'+1))
		case 1:
			str[i] = 'a' + byte(rand.Intn('z'-'a'+1))
		case 2:
			str[i] = 'A' + byte(rand.Intn('Z'-'A'+1))
		}
	}
	return string(str)
}

// RandomPrintableASCIIString 生成随机字符串。
func RandomPrintableASCIIString(length int) string {
	str := make([]byte, length)
	for i := range length {
		bigInt, _ := crand.Int(crand.Reader, big.NewInt(127-32))
		num := bigInt.Int64() + 32
		str[i] = byte(num)
	}
	return string(str)
}

// RandomBytes 生成随机的字节串。
func RandomBytes(length int) []byte {
	arr := make([]byte, length)
	for i := range length {
		bigInt, _ := crand.Int(crand.Reader, big.NewInt(1<<8))
		num := bigInt.Int64()
		arr[i] = byte(num)
	}
	return arr
}

// RandomPrintableASCIINoSpaceString 生成不包含空格的随机字符串。
func RandomPrintableASCIINoSpaceString(length int) string {
	str := make([]byte, length)
	for i := range length {
		bigInt, _ := crand.Int(crand.Reader, big.NewInt(127-33))
		num := bigInt.Int64() + 33
		str[i] = byte(num)
	}
	return string(str)
}

// GetPrintJSON 转成 JSON 字符串。
func GetPrintJSON(v any) string {
	bs, _ := json.Marshal(v)
	return string(bs)
}

// Atoi 字符串转整数。
func Atoi(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

// CertificateDERToPEM 证书格式转换。
func CertificateDERToPEM(src []byte) ([]byte, error) {
	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: src,
	}
	buf := &bytes.Buffer{}
	err := pem.Encode(buf, block)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// CertificatePEMToDER 证书格式转换。
func CertificatePEMToDER(src string) ([]byte, error) {
	block, _ := pem.Decode([]byte(src))
	if block == nil {
		return nil, errors.New("failed to decode pem block")
	}
	return block.Bytes, nil
}

// CloseIO 关闭流。发生错误时，打印日志。
func CloseIO(ctx context.Context, c io.Closer) {
	if c != nil {
		log.ErrorIf(ctx, c.Close(), "failed to close stream")
	}
}

// CloseHTTPBody 关闭 HTTP 响应体。发生错误时，打印日志。
func CloseHTTPBody(ctx context.Context, r *http.Response) {
	if r != nil && r.Body != nil {
		CloseIO(ctx, r.Body)
	}
}

// FormatDuration 将时间段格式化成中文表示形式。
func FormatDuration(d time.Duration) string {
	signed := ""
	if d < 0 {
		d = -d
		signed = "负"
	}
	result := ""
	days := d / (24 * time.Hour)
	if days > 0 {
		d = d - days*24*time.Hour
		result += fmt.Sprintf("%d天", days)
	}
	hours := d / time.Hour
	if hours > 0 {
		d = d - hours*time.Hour
		result += fmt.Sprintf("%d小时", hours)
	}
	minutes := d / time.Minute
	if minutes > 0 {
		d = d - minutes*time.Minute
		result += fmt.Sprintf("%d分钟", minutes)
	}
	seconds := d / time.Second
	if seconds > 0 {
		d = d - seconds*time.Second
		result += fmt.Sprintf("%d秒", seconds)
	}
	milliseconds := d / time.Millisecond
	if milliseconds > 0 {
		d = d - milliseconds*time.Millisecond
		result += fmt.Sprintf("%d毫秒", milliseconds)
	}
	if len(result) <= 0 {
		result = "0毫秒"
	}
	return signed + result
}

// Indirect 安全解指针。
func Indirect[T any](ptr *T) T {
	if ptr == nil {
		var t T
		return t
	}
	return *ptr
}

// TrimBlank 去除字符串两头的空白字符。
func TrimBlank(str string) string {
	return strings.TrimFunc(str,
		func(r rune) bool { return r == '\n' || r == ' ' || r == '\r' || r == '\t' || r == '\v' })
}

// TrimBlank2 去除字符串两头的空白字符。
func TrimBlank2(str []byte) []byte {
	return bytes.TrimFunc(str,
		func(r rune) bool { return r == '\n' || r == ' ' || r == '\r' || r == '\t' || r == '\v' })
}

// EncodeStructToURLQuery 将结构体 v 转成 URL query 参数。
func EncodeStructToURLQuery(v any) string {
	value := reflect.Indirect(reflect.ValueOf(v))
	if value.Kind() != reflect.Struct {
		return ""
	}
	q := url.Values{}
	typ := value.Type()
	numField := typ.NumField()
	for i := range numField {
		tag := typ.Field(i).Tag
		name := tag.Get("form")
		if len(name) <= 0 {
			continue
		}
		field := value.Field(i).Interface()
		switch val := field.(type) {
		case time.Time:
			f := tag.Get("time_format")
			if len(f) <= 0 {
				f = consts.TimeFormat
			}
			q.Add(name, val.Format(f))
		case []int:
			for _, elem := range val {
				q.Add(name, strconv.Itoa(elem))
			}
		case []string:
			for _, elem := range val {
				q.Add(name, elem)
			}
		default:
			q.Add(name, fmt.Sprint(field))
		}
	}
	return q.Encode()
}

// RunPowerShellCommands 运行 PowerShell 命令。
func RunPowerShellCommands(ctx context.Context, commands ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "PowerShell.exe", "-NoLogo", "-NoExit")
	writeCloser, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	readCloser, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	readCloser2, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	output := &bytes.Buffer{}
	writer := &SafeWriter{Writer: output}
	notifyRead := make(chan struct{}, 1)
	notifyWrite := make(chan struct{}, 1)
	ctx, cancel := context.WithCancelCause(ctx)
	wg := sync.WaitGroup{}
	defer cancel(nil)

	// 读取输出。
	wg.Go(func() {
		defer CloseIO(ctx, readCloser)
		buffer := make([]byte, 32*1024)
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-notifyRead:
				if !ok {
					return
				}
				// 读取标准输出。
				time.Sleep(100 * time.Millisecond)
				n, err2 := readCloser.Read(buffer)
				_, _ = writer.Write(buffer[:n])
				if errors.Is(err2, io.EOF) {
					return
				}
				if err2 != nil {
					close(notifyRead)
					close(notifyWrite)
					cancel(err2)
					return
				}
				notifyWrite <- struct{}{}
			}
		}
	})
	wg.Go(func() {
		defer CloseIO(ctx, readCloser2)
		buf := make([]byte, 32*1024)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			n, err2 := readCloser2.Read(buf)
			_, _ = writer.Write(buf[:n])
			if errors.Is(err2, io.EOF) {
				return
			}
			if err2 != nil {
				cancel(err2)
				return
			}
		}
	})

	// 写入数据。
	wg.Go(func() {
		defer CloseIO(ctx, writeCloser)
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-notifyWrite:
				if !ok {
					return
				}
				if len(commands) > 0 {
					command := commands[0]
					commands = commands[1:]
					_, err2 := io.Copy(writeCloser, strings.NewReader(command+"\r\n"))
					if err2 != nil {
						cancel(err2)
						close(notifyRead)
						close(notifyWrite)
						return
					}
					notifyRead <- struct{}{}
				} else {
					close(notifyWrite)
					close(notifyRead)
					return
				}
			}
		}
	})

	if err = cmd.Start(); err != nil {
		return nil, err
	}
	notifyRead <- struct{}{}
	wg.Wait()
	err = cmd.Wait()
	if err != nil {
		return output.Bytes(), err
	}
	return output.Bytes(), context.Cause(ctx)
}

// QuoteArguments 把命令参数用双引号括起来。
func QuoteArguments(arr ...string) []string {
	newArr := make([]string, len(arr))
	for i, v := range arr {
		if strings.Contains(v, `"`) {
			v = strings.ReplaceAll(v, `"`, `\"`)
		}
		newArr[i] = fmt.Sprintf(`"%s"`, v)
	}
	return newArr
}

// IsLocalEnvironment 是否是本地环境。
func IsLocalEnvironment() bool {
	return cfg.Get().Environment() == cfg.EnvironmentLocal
}

// IsProductionEnvironment 是否生产环境。
func IsProductionEnvironment() bool {
	return cfg.Get().Environment() == cfg.EnvironmentProduction
}

// CreateWindowsFirewallRule 创建防火墙规则。
func CreateWindowsFirewallRule(params *CreateWindowsFirewallRuleParams) error {
	args := []string{
		"advfirewall",
		"firewall",
		"add",
		"rule",
		"name=" + params.Name,
	}
	if len(params.Direct) > 0 {
		args = append(args, "dir="+params.Direct)
	}
	if len(params.Protocol) > 0 {
		args = append(args, "protocol="+params.Protocol)
	}
	if len(params.LocalIP) > 0 {
		args = append(args, "localip="+params.LocalIP)
	}
	if len(params.RemotePort) > 0 {
		args = append(args, "remoteip="+params.RemotePort)
	}
	if len(params.Program) > 0 {
		args = append(args, "program="+params.Program)
	}
	if len(params.Action) > 0 {
		args = append(args, "action="+params.Action)
	}
	if len(params.LocalPort) > 0 {
		args = append(args, "localport="+params.LocalPort)
	}
	if len(params.RemotePort) > 0 {
		args = append(args, "remoteport="+params.RemotePort)
	}
	output, err := exec.Command("netsh", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w, %s", err, output)
	}
	return nil
}

// DeleteWindowsFirewallRule 删除防火墙规则。
func DeleteWindowsFirewallRule(name string) error {
	output, err := exec.Command("netsh", "advfirewall", "firewall", "delete", "rule", "name="+name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w, %s", err, output)
	}
	return nil
}

// GbkToUtf8 转编码，失败就原样返回。
func GbkToUtf8(bs []byte) []byte {
	reader := transform.NewReader(bytes.NewReader(bs), simplifiedchinese.GBK.NewDecoder())
	res, err := io.ReadAll(reader)
	if err != nil {
		return bs
	}
	return res
}

func (w *SafeWriter) Write(p []byte) (n int, err error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.Writer.Write(p)
}
