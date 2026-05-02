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
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/backend/route"
	"gitee.com/ivfzhou/csms/backend/service"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
)

const (
	RequestIP      = "127.0.0.1"
	RequestSession = "01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567"
	UserPassword   = "123456"
)

const (
	windowsCertificate         = `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`
	windowsCertificatePassword = "123456"
)

const (
	androidCertificate  = `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`
	androidRSAPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAlfLDSjvghQC6/wb1fALd
cXktRMF8AOnb+s+Un8wG7mdCPg81zDO168NNV5vUTgLD1mwVHZ0DRSfm6FkvUgtk
GRUX0lWM+d3IzxRqEx4TI3RXVzjTIzZrOuYn8Z3pqHd1iTuJpqvC/EMBC5xTn7aF
HtW7OmdIw9I6lAjvph9B5sqV4tyqYC7NhVPBpRHNpLzAyIxrzaX/OZ/m0p3E+5Zw
tn/DxOjUmkTcl3eW9uVJ0g1zhg5iMWyhRgZ9TYrITvf/vz7lvOxfrBRrmxEJFh1Q
NtN2Ao6JpZmdEABsgSHlzfFzMq0E02BrG2C7X9Uhq57Mro6cgjEnQizmu/2niW0R
QQIDAQAB
-----END PUBLIC KEY-----`
	androidKeystoreStorepass = "123456"
	androidKeystoreKeypass   = "123456"
	androidKeystoreAlias     = "android_uploading_certificate"
)

var (
	Session   string
	LoginUser = &model.User{
		ID:           1,
		NameEn:       "zhangsan",
		NameZh:       "张三",
		AvatarFileID: util.FastRandomAlphaNumberString(38),
		Department:   "/技术部",
		PasswordSalt: util.FastRandomAlphaNumberString(128),
		CreatedTime:  time.Now(),
		UpdatedTime:  time.Now(),
	}
	AppInfo = &model.App{
		ID:          1,
		AppID:       util.FastRandomAlphaNumberString(32),
		UserID:      LoginUser.ID,
		Name:        "应用名",
		LogoFileID:  util.FastRandomAlphaNumberString(38),
		Platform:    model.AppPlatformWindows,
		Status:      model.AppStatusValid,
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
	}
)

func init() {
	gin.SetMode(gin.TestMode)
	protocol.Initialize(context.Background())
	// cfg.Get().Environment() = cfg.EnvironmentProduction
	// log.CurrentLevel = log.LevelError
	sessionBytes, _ := json.Marshal(service.SessionInfo{
		UserID:  LoginUser.ID,
		User:    LoginUser.NameEn,
		Session: RequestSession,
		IP:      RequestIP,
	})
	Session = string(sessionBytes)
	md5Sum := md5.Sum([]byte(LoginUser.PasswordSalt + UserPassword))
	LoginUser.PasswordDigest = hex.EncodeToString(md5Sum[:])
}

func GenerateBytes(length int) []byte {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return b
}

func GenerateECDSAKeyPEM(curve string) (string, string, error) {
	var c elliptic.Curve

	switch curve {
	case "P224":
		c = elliptic.P224()
	case "P256":
		c = elliptic.P256()
	case "P384":
		c = elliptic.P384()
	case "P521":
		c = elliptic.P521()
	default:
		return "", "", fmt.Errorf("unsupported curve: %s", curve)
	}

	// 生成 ECDSA 私钥。
	privateKey, err := ecdsa.GenerateKey(c, rand.Reader)
	if err != nil {
		return "", "", err
	}

	// 将私钥编码为 PKCS#8 格式。
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", "", err
	}

	// 创建私钥 PEM 块。
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// 提取公钥
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}

	// 创建公钥 PEM 块。
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(privateKeyPEM), string(publicKeyPEM), nil
}

func GenerateJPEG(t *testing.T, width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	bgColor := color.RGBA{R: 200, G: 220, B: 255, A: 255}
	for y := range height {
		for x := range width {
			img.Set(x, y, bgColor)
		}
	}
	rectColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	rectX1, rectY1 := 100, 100
	rectX2, rectY2 := 400, 300
	for y := rectY1; y < rectY2; y++ {
		for x := rectX1; x < rectX2; x++ {
			img.Set(x, y, rectColor)
		}
	}
	textColor := color.RGBA{R: 0, G: 150, B: 0, A: 255}
	for i := range 50 {
		x := 500 + i%20*10
		y := 200 + i/5*10
		for dy := range 8 {
			for dx := range 8 {
				img.Set(x+dx, y+dy, textColor)
			}
		}
	}
	buf := &bytes.Buffer{}
	options := &jpeg.Options{Quality: 90}
	err := jpeg.Encode(buf, img, options)
	if err != nil {
		t.Errorf("jpeg encode error %v", err)
		return nil
	}
	return buf.Bytes()
}

func GeneratePNG(t *testing.T, width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	bgColor := color.RGBA{R: 200, G: 220, B: 255, A: 255}
	for y := range height {
		for x := range width {
			img.Set(x, y, bgColor)
		}
	}
	rectColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	rectX1, rectY1 := 100, 100
	rectX2, rectY2 := 400, 300
	for y := rectY1; y < rectY2; y++ {
		for x := rectX1; x < rectX2; x++ {
			img.Set(x, y, rectColor)
		}
	}
	textColor := color.RGBA{R: 0, G: 150, B: 0, A: 255}
	for i := range 50 {
		x := 500 + i%20*10
		y := 200 + i/5*10
		for dy := range 8 {
			for dx := range 8 {
				img.Set(x+dx, y+dy, textColor)
			}
		}
	}
	buf := &bytes.Buffer{}
	err := png.Encode(buf, img)
	if err != nil {
		t.Errorf("png encode error %v", err)
		return nil
	}
	return buf.Bytes()
}

func GenerateGIF(t *testing.T, width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	bgColor := color.RGBA{R: 200, G: 220, B: 255, A: 255}
	for y := range height {
		for x := range width {
			img.Set(x, y, bgColor)
		}
	}
	rectColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	rectX1, rectY1 := 100, 100
	rectX2, rectY2 := 400, 300
	for y := rectY1; y < rectY2; y++ {
		for x := rectX1; x < rectX2; x++ {
			img.Set(x, y, rectColor)
		}
	}
	textColor := color.RGBA{R: 0, G: 150, B: 0, A: 255}
	for i := range 50 {
		x := 500 + i%20*10
		y := 200 + i/5*10
		for dy := range 8 {
			for dx := range 8 {
				img.Set(x+dx, y+dy, textColor)
			}
		}
	}
	buf := &bytes.Buffer{}
	options := &gif.Options{}
	err := gif.Encode(buf, img, options)
	if err != nil {
		t.Errorf("gif encode error %v", err)
		return nil
	}
	return buf.Bytes()
}

func CheckAndReadBody(t *testing.T, rsp *httptest.ResponseRecorder) ([]byte, string) {
	if rsp.Code != http.StatusOK {
		t.Errorf("expect http code %d, but got %d", http.StatusOK, rsp.Code)
	}
	return rsp.Body.Bytes(), rsp.Header().Get("Content-Disposition")
}

func CheckAndUnmarshalBody[T any](t *testing.T, rsp *httptest.ResponseRecorder, code errs.Code) *util.Response[T] {
	if rsp.Code != http.StatusOK {
		t.Errorf("expect http code %d, but got %d", http.StatusOK, rsp.Code)
	}
	var rspBodyObj util.Response[T]
	err := json.Unmarshal(rsp.Body.Bytes(), &rspBodyObj)
	if err != nil {
		t.Errorf("expect response body to be unmarshallable, but got %v", err)
	}
	if rspBodyObj.Code != code {
		t.Errorf("expect %v, but got %v", code, rspBodyObj.Code)
	}
	return &rspBodyObj
}

func CreatePostJSONRequest[T any](ctx context.Context, uri string, req *T) *http.Request {
	bs, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, uri, bytes.NewReader(bs))
	request.Header.Set("Date", time.Now().Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Real-IP", RequestIP)
	request.Header.Set("Cookie",
		fmt.Sprintf("%s=%s; %s=%s", consts.HTTPHeaderSessionUser, LoginUser.NameEn, consts.HTTPHeaderSessionKey, RequestSession))
	return request.WithContext(ctx)
}

func CreatePostJSONRequestWithApp[T any](ctx context.Context, uri, appID string, req *T) *http.Request {
	bs, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, uri+"/"+appID, bytes.NewReader(bs))
	request.Header.Set("Date", time.Now().Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Real-IP", RequestIP)
	request.Header.Set("Cookie",
		fmt.Sprintf("%s=%s; %s=%s", consts.HTTPHeaderSessionUser, LoginUser.NameEn, consts.HTTPHeaderSessionKey, RequestSession))
	return request.WithContext(ctx)
}

func CreatePostMultiFormRequest(ctx context.Context, uri string, req io.Reader, contentType string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, uri, req)
	request.Header.Set("Date", time.Now().Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	request.Header.Set("Content-Type", contentType)
	request.Header.Set("X-Real-IP", RequestIP)
	request.Header.Set("Cookie", fmt.Sprintf("%s=%s; %s=%s",
		consts.HTTPHeaderSessionUser, LoginUser.NameEn, consts.HTTPHeaderSessionKey, RequestSession))
	return request.WithContext(ctx)
}

func CreateGetRequest(ctx context.Context, uri string, queryStruct any) *http.Request {
	request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s?%s", uri, util.EncodeStructToURLQuery(queryStruct)), nil)
	request.Header.Set("Date", time.Now().Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	request.Header.Set("X-Real-IP", RequestIP)
	request.Header.Set("Cookie", fmt.Sprintf("%s=%s; %s=%s",
		consts.HTTPHeaderSessionUser, LoginUser.NameEn, consts.HTTPHeaderSessionKey, RequestSession))
	return request.WithContext(ctx)
}

func CreateGetRequestWithApp(ctx context.Context, uri, appID string, queryStruct any) *http.Request {
	request := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/%s?%s", uri, appID, util.EncodeStructToURLQuery(queryStruct)), nil)
	request.Header.Set("Date", time.Now().Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	request.Header.Set("X-Real-IP", RequestIP)
	request.Header.Set("Cookie", fmt.Sprintf("%s=%s; %s=%s",
		consts.HTTPHeaderSessionUser, LoginUser.NameEn, consts.HTTPHeaderSessionKey, RequestSession))
	return request.WithContext(ctx)
}

func CreateDeleteRequest(ctx context.Context, uri string, queryStruct any) *http.Request {
	request := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("%s?%s", uri, util.EncodeStructToURLQuery(queryStruct)), nil)
	request.Header.Set("Date", time.Now().Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	request.Header.Set("X-Real-IP", RequestIP)
	request.Header.Set("Cookie", fmt.Sprintf("%s=%s; %s=%s",
		consts.HTTPHeaderSessionUser, LoginUser.NameEn, consts.HTTPHeaderSessionKey, RequestSession))
	return request.WithContext(ctx)
}

func CreateDeleteRequestWithApp(ctx context.Context, uri, appID string, queryStruct any) *http.Request {
	request := httptest.NewRequest(http.MethodDelete,
		fmt.Sprintf("%s/%s?%s", uri, appID, util.EncodeStructToURLQuery(queryStruct)), nil)
	request.Header.Set("Date", time.Now().Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	request.Header.Set("X-Real-IP", RequestIP)
	request.Header.Set("Cookie", fmt.Sprintf("%s=%s; %s=%s",
		consts.HTTPHeaderSessionUser, LoginUser.NameEn, consts.HTTPHeaderSessionKey, RequestSession))
	return request.WithContext(ctx)
}

func ServeHTTP(ctx context.Context, req *http.Request) *httptest.ResponseRecorder {
	rsp := httptest.NewRecorder()
	route.Initialize(ctx).ServeHTTP(rsp, req)
	return rsp
}

func TakeIntPtr(i int) *int {
	return &i
}

func TakeStringPtr(s string) *string {
	return &s
}
