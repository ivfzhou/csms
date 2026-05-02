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
	"crypto/aes"
	"crypto/cipher"
)

// AESCBCEncrypt 加密。
func AESCBCEncrypt(key, plain []byte) ([]byte, error) {
	if len(plain) <= 0 {
		return nil, nil
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	plain = pkcs5Padding(plain, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(plain))
	blockMode.CryptBlocks(crypted, plain)
	return crypted, nil
}

// AESCBCDecrypt 解密。
func AESCBCDecrypt(crypted, key []byte) ([]byte, error) {
	if len(crypted) == 0 {
		return nil, nil
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	plain := make([]byte, len(crypted))
	blockMode.CryptBlocks(plain, crypted)
	return pkcs5Unpadding(plain), nil
}

func pkcs5Padding(data []byte, blockSize int) []byte {
	paddingCount := blockSize - len(data)%blockSize
	paddingText := bytes.Repeat([]byte{byte(paddingCount)}, paddingCount)
	return append(data, paddingText...)
}

func pkcs5Unpadding(data []byte) []byte {
	return data[:(len(data) - int(data[len(data)-1]))]
}
