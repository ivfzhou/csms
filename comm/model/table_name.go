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

package model

import "time"

// GetFileTableName 获取文件信息表表名。
func GetFileTableName(date time.Time) string {
	return (&File{}).TableName() + "_" + date.Format("200601")
}

// GetEventTableName 获取应用事件信息表表名。
func GetEventTableName(date time.Time) string {
	return (&Event{}).TableName() + "_" + date.Format("200601")
}

// GetAndroidSigningJobTableName 获取安卓签名任务表表名。
func GetAndroidSigningJobTableName(date time.Time) string {
	return (&AndroidSigningJob{}).TableName() + "_" + date.Format("200601")
}

// GetAppleSigningJobTableName 获取 Apple 签名任务表表名。
func GetAppleSigningJobTableName(date time.Time) string {
	return (&AppleSigningJob{}).TableName() + "_" + date.Format("200601")
}

// GetWindowsSigningJobTableName 获取 Windows 签名任务表表名。
func GetWindowsSigningJobTableName(date time.Time) string {
	return (&WindowsSigningJob{}).TableName() + "_" + date.Format("200601")
}

// GetFileTableNameByID 获取文件信息表表名。
func GetFileTableNameByID(id string) string {
	if len(id) < 6 {
		return (&File{}).TableName()
	}
	yearMonth := id[:6]
	_, err := time.Parse("200601", yearMonth)
	if err != nil {
		return (&File{}).TableName()
	}
	return (&File{}).TableName() + "_" + yearMonth
}

// GetAndroidSigningJobByID 获取安卓签名任务表表名。
func GetAndroidSigningJobByID(id string) string {
	if len(id) < 6 {
		return (&AndroidSigningJob{}).TableName()
	}
	yearMonth := id[:6]
	_, err := time.Parse("200601", yearMonth)
	if err != nil {
		return (&AndroidSigningJob{}).TableName()
	}
	return (&AndroidSigningJob{}).TableName() + "_" + yearMonth
}

// GetAppleSigningJobByID 获取 Apple 签名任务表表名。
func GetAppleSigningJobByID(id string) string {
	if len(id) < 6 {
		return (&AppleSigningJob{}).TableName()
	}
	yearMonth := id[:6]
	_, err := time.Parse("200601", yearMonth)
	if err != nil {
		return (&AppleSigningJob{}).TableName()
	}
	return (&AppleSigningJob{}).TableName() + "_" + yearMonth
}

// GetWindowsSigningJobByID 获取 Windows 签名任务表表名。
func GetWindowsSigningJobByID(id string) string {
	if len(id) < 6 {
		return (&WindowsSigningJob{}).TableName()
	}
	yearMonth := id[:6]
	_, err := time.Parse("200601", yearMonth)
	if err != nil {
		return (&WindowsSigningJob{}).TableName()
	}
	return (&WindowsSigningJob{}).TableName() + "_" + yearMonth
}
