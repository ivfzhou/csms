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

package consts

import (
	"flag"
	"time"

	cc "gitee.com/ivfzhou/csms/comm/consts"
)

// 服务模式。
const (
	ModeWindows = "Windows"
	ModeAndroid = "Android"
	ModeApple   = "Apple"
)

// 签名方案。
const (
	SignatureSchemaVersion1 = 1
	SignatureSchemaVersion2 = 2
	SignatureSchemaVersion3 = 3
	SignatureSchemaVersion4 = 4
)

const (
	// 时间戳服务器地址。
	DefaultTimeServer = "http://timestamp.digicert.com"
	// 消息出队延迟。
	MessageDelayTime = 30 * time.Second
	// 任务最大出队次数。
	MaximinOutQueueTimes = 10
	// inf 模板内容。
	INFTemplate = `;Copyright (c) 1996-2011 Kaspersky Lab, CHSC. All rights Reserved

[Version]
Signature	= "$Windows NT$"
Class		= "System"							;This is determined by the work this filter driver does
ClassGuid 	= {4d36e97d-e325-11ce-bfc1-08002be10318}	;Class = ActivityMonitor
Provider 	= %%MfgName%%
;CatalogFile	= %%ImageName%%.cat ; Catalog files are supplied by the Microsoft Windows Hardware Quality Lab (WHQL)
DriverVer= 07/24/2012,1.0.0.121            ;mm/dd/yy
CatalogFile	= %[1]s.cat
PnpLockDown = 1

;; Default install sections
[DefaultInstall.NTx86]
OptionDesc	= %%ServiceDesc%%
CopyFiles	= DriverFiles

[DefaultInstall.NTamd64]
OptionDesc	= %%ServiceDesc%%
CopyFiles	= DriverFiles

[DefaultInstall.NTx86.Services]
AddService	= %%ServiceName%%,,Service

[DefaultInstall.NTamd64.Services]
AddService	= %%ServiceName%%,,Service

[DestinationDirs]
DriverFiles	   = 12 ;%%windir%%\system32\drivers
DefaultDestDir = 12

[SourceDisksNames]
1 = %%Disk1%%

[SourceDisksFiles]
%[1]s.sys = 1

;; Services Section
[Service]
DisplayName	= %%ServiceName%%
Description	= %%ServiceDesc%%
ServiceBinary = %%12%%\%%ImageName%%.sys ;%%windir%%\system32\drivers\elamts.sys
ServiceType = 1                   ; SERVICE_KERNEL_DRIVER
StartType   = 0                   ; SERVICE_BOOT_START
ErrorControl= 3                   ; SERVICE_ERROR_CRITICAL
LoadOrderGroup = "Early-Launch"

;; Copy Files
[DriverFiles]
%%ImageName%%.sys,,,2

;; String Section
[Strings]
MfgName	    = ""
ServiceName = "%[1]s"
ServiceDesc = ""
ImageName   = "%[1]s"
Disk1	    = "%[1]s Source Media"
`
	// ddf 模板。
	DDFTemplate = `; *** amd64.ddf example ***
;
.OPTION EXPLICIT     ; Generate errors
.Set CabinetFileCountThreshold=0
.Set FolderFileCountThreshold=0
.Set FolderSizeThreshold=0
.Set MaxCabinetSize=0
.Set MaxDiskFileCount=0
.Set MaxDiskSize=0
.Set CompressionType=MSZIP
.Set Cabinet=on
.Set Compress=on
.Set DiskDirectoryTemplate=%[1]s
; Specify file name for new cab file
.Set CabinetNameTemplate=%[2]s.cab
; Specify the subdirectory for the files.  
; Your cab file should not have files at the root level, 
; and each driver package must be in a separate subfolder.
.Set DestinationDir=%[2]s
;Add file path here
%[3]s
%[4]s
%[5]s
`
)

var (
	// signtool.exe 程序文件路径。
	SigntoolFilePath string
	// winevsigner.exe 程序文件路径。
	WinevsignerFilePath string
	// apksigner 文件路径。
	ApksignerFilePath string
	// jarsigner 文件路径。
	JarsignerFilePath string
	// Java 家目录。
	JavaHomeFilePath string
	// inf2Cat.exe 文件路径。
	Inf2CatFilePath string
	// makecab.exe 文件路径。
	MakeecabFilePath string
	// zsign 文件路径。
	ZsignFilePath string
)

// AddCommandFlag 增加程序命令参数。
func AddCommandFlag() {
	defaultSigntoolFilePath := ".\\signtool.exe"
	if cc.PointerSize == 4 {
		defaultSigntoolFilePath = ".\\signtool32.exe"
	}
	flag.StringVar(&SigntoolFilePath, cc.CommandFlagSigntoolFilePath, defaultSigntoolFilePath, "signtool.exe file path")
	flag.StringVar(&JavaHomeFilePath, cc.CommandFlagJavaHomeFilePath, "", "java home file path")
	flag.StringVar(&ApksignerFilePath, cc.CommandFlagApksignerFilePath, "apksigner", "apksigner file path")
	flag.StringVar(&Inf2CatFilePath, cc.CommandFlagInf2CatFilePath, "inf2Cat.exe", "inf2Cat.exe file path")
	flag.StringVar(&MakeecabFilePath, cc.CommandFlagMakecabFilePath, ".\\makecab.exe", "makecab.exe file path")
	flag.StringVar(&JarsignerFilePath, cc.CommandFlagJarsignerFilePath, "jarsigner", "jarsigner file path")
	flag.StringVar(&WinevsignerFilePath, cc.CommandFlagWinevsignerFilePath, ".\\winevsigner.exe",
		"winevsigner.exe file path")
	flag.StringVar(&ZsignFilePath, cc.CommandFlagZsignFilePath, "./zsign", "zsign file path")
}
