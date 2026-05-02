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

	cc "gitee.com/ivfzhou/csms/comm/consts"
)

// 运行模式。
const (
	ModeTestMachine       = "TestMachine"
	ModeControllerMachine = "ControllerMachine"
	ModeHostMachine       = "HostMachine"
)

const (
	// 添加的防火墙规则名。
	FirewallRuleName = "hlk_manager_set"
	// 测试信息文件路径。
	TestInfoFilePath = `C:\hlk_test_info.json`
	// 存放测试信息的文件夹。
	TestInfoDirectoryPath = `C:\hlk_test_info`
	// 虚拟机检查点名称。
	VirtualMachineCheckPointName = "init"
	// 服务安装脚本名称。
	InstallServiceFileName = "install.bat"
	// 证书名称。
	TestCertificateName = "Contoso.com(Test)"
	// 证书信息。
	TestCertificateEku = "1.3.6.1.5.5.7.3.3"
	// 默认池名称。
	HLKStudioDefaultPool = `Default Pool`
	// 机器状态。
	MachineReadyState = "Ready"
	// 对话框处理通信文件。
	DialogConversationFilePath = `C:\dialog_conversation.txt`
	// 需要人工交互的测试项。
	ManualTestType = "Manual"
	// WFP 配置文件路径。
	WFPLogoTxtFilePath = `C:\Windows\System32\WFPLogo.Info`
	// WFP 应答文件路径。
	WFPLogoAnswerFilePath = `C:\Windows\System32\WFPLogo.Answer`
	// 默认 WFP 应答配置。
	DefaultWFPAnswerFileData = `configurationTimer = 10;
# ArchitecturalDesign.StreamInjection.NoStreamStarvation.1
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=allow protocol=6 localip=%LOCAL_IP% remoteip=%REMOTE_IP% localport=%LOCAL_PORT% remoteport=%REMOTE_PORT% & netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=allow protocol=6 localip=%LOCAL_IP% remoteip=%REMOTE_IP% localport=%LOCAL_PORT% remoteport=%REMOTE_PORT%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# ArchitecturalDesign.StreamInjection.NoStreamStarvation.2
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule  dir=in  action=allow protocol=6 localip=%LOCAL_IP% remoteip=%REMOTE_IP% localport=%LOCAL_PORT% remoteport=%REMOTE_PORT% & netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=allow protocol=6 localip=%LOCAL_IP% remoteip=%REMOTE_IP% localport=%LOCAL_PORT% remoteport=%REMOTE_PORT%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# ArchitecturalDesign.SupportPowerManagedStates.1
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=allow program=%APPLICATION% localip=%LOCAL_IP% remoteip=%REMOTE_IP% protocol=%PROTOCOL%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# ArchitecturalDesign.SupportPowerManagedStates.2
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=block program=%APPLICATION% localip=%LOCAL_IP% remoteip=%REMOTE_IP% protocol=%PROTOCOL%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# ArchitecturalDesign.SupportPowerManagedStates.3
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=allow program=%APPLICATION% localip=%LOCAL_IP% remoteip=%REMOTE_IP% protocol=%PROTOCOL%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# ArchitecturalDesign.SupportPowerManagedStates.4
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=block program=%APPLICATION% localip=%LOCAL_IP% remoteip=%REMOTE_IP% protocol=%PROTOCOL%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# ArchitecturalDesign.SupportPowerManagedStates.5
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=allow program=%APPLICATION% localip=%LOCAL_IP% remoteip=%REMOTE_IP% protocol=%PROTOCOL%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# ArchitecturalDesign.SupportPowerManagedStates.6
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=block program=%APPLICATION% localip=%LOCAL_IP% remoteip=%REMOTE_IP% protocol=%PROTOCOL%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# ArchitecturalDesign.SupportPowerManagedStates.7
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=allow program=%APPLICATION% localip=%LOCAL_IP% remoteip=%REMOTE_IP% protocol=%PROTOCOL%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# ArchitecturalDesign.SupportPowerManagedStates.8
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=block program=%APPLICATION% localip=%LOCAL_IP% remoteip=%REMOTE_IP% protocol=%PROTOCOL%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.IPAddressExceptions.1
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=allow localip=%LOCAL_IP% remoteip=%REMOTE_IP%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.IPAddressExceptions.2
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=block localip=%LOCAL_IP% remoteip=%REMOTE_IP%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.IPAddressExceptions.3
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=allow localip=%LOCAL_IP% remoteip=%REMOTE_IP%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.IPAddressExceptions.4
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=block localip=%LOCAL_IP% remoteip=%REMOTE_IP%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.IPAddressExceptions.5
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=allow localip=%LOCAL_IP% remoteip=%REMOTE_IP%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.IPAddressExceptions.6
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=block localip=%LOCAL_IP% remoteip=%REMOTE_IP%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.IPAddressExceptions.7
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=allow localip=%LOCAL_IP% remoteip=%REMOTE_IP%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.IPAddressExceptions.8
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=block localip=%LOCAL_IP% remoteip=%REMOTE_IP%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.PortExceptions.1
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=allow protocol=%PROTOCOL% localport=%LOCAL_PORT% remoteport=%REMOTE_PORT%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.PortExceptions.2
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=block protocol=%PROTOCOL% localport=%LOCAL_PORT% remoteport=%REMOTE_PORT%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.PortExceptions.3
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=allow protocol=%PROTOCOL% localport=%LOCAL_PORT% remoteport=%REMOTE_PORT%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.PortExceptions.4
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=block protocol=%PROTOCOL% localport=%LOCAL_PORT% remoteport=%REMOTE_PORT%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.PortExceptions.5
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=allow protocol=%PROTOCOL% localport=%LOCAL_PORT% remoteport=%REMOTE_PORT%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.PortExceptions.6
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=block protocol=%PROTOCOL% localport=%LOCAL_PORT% remoteport=%REMOTE_PORT%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.PortExceptions.7
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=allow protocol=%PROTOCOL% localport=%LOCAL_PORT% remoteport=%REMOTE_PORT%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.PortExceptions.8
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=block protocol=%PROTOCOL% localport=%LOCAL_PORT% remoteport=%REMOTE_PORT%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.ProtocolExceptions.1
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=allow protocol=%PROTOCOL%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.ProtocolExceptions.2
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=block protocol=%PROTOCOL%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.ProtocolExceptions.3
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=allow protocol=%PROTOCOL%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.ProtocolExceptions.4
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=block protocol=%PROTOCOL%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.ProtocolExceptions.5
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=allow protocol=%PROTOCOL%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.ProtocolExceptions.6
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=block protocol=%PROTOCOL% ;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.ProtocolExceptions.7
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=allow protocol=%PROTOCOL%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.ProtocolExceptions.8
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=block protocol=%PROTOCOL%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.ICMPExceptions.1
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=allow protocol=icmpv4:%ICMP_TYPE%,%ICMP_CODE%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.ICMPExceptions.2
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=block protocol=icmpv4:%ICMP_TYPE%,%ICMP_CODE%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.ICMPExceptions.3
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=allow protocol=icmpv4:%ICMP_TYPE%,%ICMP_CODE%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.ICMPExceptions.4
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=block protocol=icmpv4:%ICMP_TYPE%,%ICMP_CODE%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.ICMPExceptions.5
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=allow protocol=icmpv6:%ICMP_TYPE%,%ICMP_CODE%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.ICMPExceptions.6
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=block protocol=icmpv6:%ICMP_TYPE%,%ICMP_CODE%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.ICMPExceptions.7
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=allow protocol=icmpv6:%ICMP_TYPE%,%ICMP_CODE%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.Support5TupleExceptions.ICMPExceptions.8
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=block protocol=icmpv6:%ICMP_TYPE%,%ICMP_CODE%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.SupportApplicationExceptions.1
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=allow program=%APPLICATION%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.SupportApplicationExceptions.2
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=block program=%APPLICATION%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.SupportApplicationExceptions.3
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=allow program=%APPLICATION%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.SupportApplicationExceptions.4
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=block program=%APPLICATION%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.SupportApplicationExceptions.5
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=allow program=%APPLICATION%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.SupportApplicationExceptions.6
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=block program=%APPLICATION%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.SupportApplicationExceptions.7
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=allow program=%APPLICATION%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.SupportApplicationExceptions.8
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=block program=%APPLICATION%;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.SupportMACAddressExceptions.1
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=allow;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.SupportMACAddressExceptions.2
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=block;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.SupportMACAddressExceptions.3
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=allow;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.SupportMACAddressExceptions.4
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=block;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.SupportMACAddressExceptions.5
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=allow;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.SupportMACAddressExceptions.6
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=out action=block;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.SupportMACAddressExceptions.7
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=allow;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# Firewall.SupportMACAddressExceptions.8
LOGO_ADD = netsh advfirewall firewall add rule name=certs-hlk-test-rule dir=in action=block;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# ArchitecturalDesign.PacketInjection.NoDeadlocks
numPacketInjectionCommands = 0;
# ArchitecturalDesign.PacketInjection.NoDeadlocks.1
LOGO_ADD = netsh add firewal rule advfirewall firewall add rule name=certs-hlk-test-rule dir=out localip=%LOCAL_IP% remoteip=%REMOTE_IP% action=allow;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# ArchitecturalDesign.PacketInjection.NoDeadlocks.2
LOGO_ADD = netsh add firewal rule advfirewall firewall add rule name=certs-hlk-test-rule dir=in localip=%LOCAL_IP% remoteip=%REMOTE_IP% action=allow;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# ArchitecturalDesign.PacketInjection.NoDeadlocks.3
LOGO_ADD = netsh add firewal rule advfirewall firewall add rule name=certs-hlk-test-rule dir=out localip=%LOCAL_IP% remoteip=%REMOTE_IP% action=allow;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
# ArchitecturalDesign.PacketInjection.NoDeadlocks.4
LOGO_ADD = netsh add firewal rule advfirewall firewall add rule name=certs-hlk-test-rule dir=in localip=%LOCAL_IP% remoteip=%REMOTE_IP% action=allow;
LOGO_DELETE = netsh advfirewall firewall delete rule name=certs-hlk-test-rule;
`
	// INF 模板。
	INFTemplate = `;Copyright (c) 1996-2011 Kaspersky Lab, CHSC. All rights Reserved

[Version]
Signature	= "$Windows NT$"
Class		= "System"							;This is determined by the work this filter driver does
ClassGuid 	= {4d36e97d-e325-11ce-bfc1-08002be10318}	;Class = ActivityMonitor
Provider 	= %MfgName%
;CatalogFile	= %ImageName%.cat ; Catalog files are supplied by the Microsoft Windows Hardware Quality Lab (WHQL)
DriverVer= 07/24/2012,1.0.0.121            ;mm/dd/yy
CatalogFile	= !!REPLACE_ME!!.cat



[ModelSection]
%ServiceDesc%=DefaultInstall,%ServiceName%

[ModelSection.ntamd64]
%ServiceDesc%=DefaultInstall,%ServiceName%

;; Default install sections
[DefaultInstall]
OptionDesc	= %ServiceDesc%
CopyFiles	= DriverFiles

[DefaultInstall.ntamd64]
OptionDesc	= %ServiceDesc%
CopyFiles	= DriverFiles

[DefaultInstall.Services]
AddService	= %ServiceName%,,Service

[DefaultInstall.ntamd64.Services]
AddService	= %ServiceName%,,Service


[DestinationDirs]
DriverFiles	   = 12 ;%windir%\system32\drivers
DefaultDestDir = 12

[SourceDisksNames]
1 = %Disk1%

[SourceDisksFiles]
!!REPLACE_ME!!.sys = 1

;; Services Section
[Service]
DisplayName	= %ServiceName%
Description	= %ServiceDesc%
ServiceBinary = %12%\%ImageName%.sys ;%windir%\system32\drivers\elamts.sys
ServiceType = 1                   ; SERVICE_KERNEL_DRIVER
StartType   = 0                   ; SERVICE_BOOT_START
ErrorControl= 3                   ; SERVICE_ERROR_CRITICAL
LoadOrderGroup = "Early-Launch"


;; Copy Files
[DriverFiles]
%ImageName%.sys,,,2

;; String Section
[Strings]
MfgName	    = "Tencent"
ServiceName = "!!REPLACE_ME!!"
ServiceDesc = "Tencent QQPCMgr Protection Component"
ImageName   = "!!REPLACE_ME!!"
Disk1	    = "!!REPLACE_ME!! Source Media"

`
)

// 人工交互测试项名称。
const (
	TestNameWFP      = "WindowsFilteringPlatform_Tests"
	TestNameTT       = "TransitionTechnologies_Tests"
	TestNameHSP      = "Hardware-enforced Stack Protection Compatibility Test"
	TestNameELAMLogo = "ELAM Logo Test"
)

// 对话框内容。
const (
	CMDContentAudioCodec = "Press Enter when you are done selecting all enhancements for render device"
	CMDContentHSP        = "Hardware-enforced Stack Protection Compatibility Test"
)

// 测试项弹框标题。
const (
	DialogTitleWFPGatherInfo  = "WFPLogo Information Gathering"
	DialogTitleELAM           = "MVI Membership Verification"
	DialogTitleWFPIP          = "Firewall.Support5TupleExceptions.IPAddressExceptions"
	DialogTitleWFPMAC         = "Firewall.SupportMACAddressExceptions"
	DialogTitleWFPProtocol    = "Firewall.Support5TupleExceptions.ProtocolExceptions"
	DialogTitleWFPPort        = "Firewall.Support5TupleExceptions.PortExceptions"
	DialogTitleWFPICMP        = "Firewall.Support5TupleExceptions.ICMPExceptions"
	DialogTitleWFPNoDeadlocks = "ArchitecturalDesign.PacketInjection.NoDeadlocks"
	DialogTitleWFPPower       = "ArchitecturalDesign.SupportPowerManagedStates"
	DialogTitleWFPStream      = "ArchitecturalDesign.StreamInjection.NoStreamStarvation"
	DialogTitleWFPApp         = "Firewall.SupportApplicationExceptions"
	DialogTitleWFPInfo        = "WFPLogo.Info - Notepad"
	DialogTitleCMD            = `C:\Windows\SYSTEM32\cmd.exe`
	DialogTitleTeredo         = "Teredo WLK test"
)

var (
	// certmgr.exe 程序文件路径。
	CertmgrFilePath string
	// makecert.exe 程序文件路径。
	MakecertFilePath string
	// signtool.exe 程序文件路径。
	SigntoolFilePath string
	// HLK 工具脚本文件路径。
	HLKToolScriptFilePath string
	// 默认 WFP 测试配置。
	DefaultWFPLogo = [...]any{
		"CompanyName", `"none"`,
		"ProductName", `"none"`,
		"EnableDriverVerifier", 1,
		"CalloutDriver", 1,
		"IsAFirewall", 1,
		"LayeredOnMicrosoftWindowsFirewall", 1,
		"DoesMACFiltering", 1,
		"DoesVSwitchFiltering", 1,
		"DoesPacketInjection", 1,
		"DoesStreamInjection", 1,
		"DoesConnectionProxying", 1,
		"SupportModernApplications", 1,
		"CleanUninstall", 1,
		"NoProxyDeadlocks", 1,
		"IdentifyingProvider", 1,
		"AssociateProvider", 1,
		"TerminatingFilter", 1,
		"UseOwnSubLayer", 1,
		"MaintainHelperClass", 1,
		"NoAVs", 1,
		"NonTampered3rdPartyObjects", 1,
		"NoPacketInjectionDeadlocks", 1,
		"NoStreamStarvation", 1,
		"SupportPowerManagement", 1,
		"WFPObjectEnumAndACLs", 1,
		"MaxWinSock", 1,
		"ProperlyDisableWindowsFirewall", 1,
		"NoPermitBlockAll", 1,
		"SupportTupleExceptions", 1,
		"SupportAppExceptions", 1,
		"SupportMACAddressExceptions", 1,
		"UseWFP", 1,
		"SupportARP", 1,
		"SupportNeighborDiscovery", 1,
		"SupportDHCP", 1,
		"SupportIPv4", 1,
		"SupportIPv6", 1,
		"SupportDNS", 1,
		"Support6To4", 1,
		"SupportAutomaticUpdates", 1,
		"SupportBasicWebsiteBrowsing", 1,
		"SupportFileAndPrinterSharing", 1,
		"SupportICMPErrorMesages", 1,
		"SupportInternetStreaming", 1,
		"SupportMediaExtenderStreaming", 1,
		"SupportMobileBroadBand", 1,
		"SupportPeerNameResolution", 1,
		"SupportRemoteAssistance", 1,
		"SupportRemoteDesktop", 1,
		"SupportTeredo", 1,
		"SupportVirtualPrivateNetworking", 1,
		"InteropWithOtherExtensions", 1,
		"NoEgressModification", 1,
		"SupportLiveMigration", 1,
		"SupportRemoval", 1,
		"SupportReordering", 1,
		"InteropWithWFPSampler", 1,
	}
)

// AddCommandFlag 增加程序命令参数。
func AddCommandFlag() {
	defaultCertmgrFilePath := "./certmgr.exe"
	defaultMakecertFilePath := "./makecert.exe"
	defaultSigntoolFilePath := "./signtool.exe"
	if cc.PointerSize == 4 {
		defaultCertmgrFilePath = "./certmgr32.exe"
		defaultMakecertFilePath = "./makecert32.exe"
		defaultSigntoolFilePath = "./signtool32.exe"
	}
	flag.StringVar(&CertmgrFilePath, cc.CommandFlagCertmgrFilePath, defaultCertmgrFilePath, "certmgr.exe file path")
	flag.StringVar(&MakecertFilePath, cc.CommandFlagMakecertFilePath, defaultMakecertFilePath, "makecert.exe file path")
	flag.StringVar(&SigntoolFilePath, cc.CommandFlagSigntoolFilePath, defaultSigntoolFilePath, "signtool.exe file path")
	flag.StringVar(&HLKToolScriptFilePath, cc.CommandFlagHLKToolScriptFilePath, "./hlk_tool.ps1",
		"HLK tool script file path")
}
