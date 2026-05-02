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
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	bp "gitee.com/ivfzhou/csms/backend/protocol"
	cc "gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
	"gitee.com/ivfzhou/csms/hlk_manager/consts"
)

var (
	ipv4Matcher = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	ipv6Matcher = regexp.MustCompile(`(::[0-9a-fA-F]{1,4})|([0-9a-fA-F]{1,4}(:{1,2}[0-9a-fA-F]{1,4}){1,7})`)
	numMatcher  = regexp.MustCompile(`\d+$`)
	macMatcher  = regexp.MustCompile(`[0-9a-fA-F]{2}(-[0-9a-fA-F]{2}){5}`)
)

type testJobInfo struct {
	ID        int    `json:"id,omitempty"`
	Workspace string `json:"workspace,omitempty"`
}

// TestMachineStart 测试机器程序逻辑入口。
func TestMachineStart(ctx context.Context, system string) <-chan struct{} {
	if _, ok := model.AllWHQLJobTestSystems[system]; !ok {
		log.Fatal(ctx, cc.ExitCodeHLKManagerInvalidSystem, "system version is invalid", system)
	}
	tick := time.Tick(3 * time.Second)
	closeChan := make(chan struct{})
	log.Info(ctx, "start test machine", system)
	go func() {
		defer close(closeChan)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			innerCtx := ctxs.New()
			select {
			case <-tick:
				stopTick, running, whqlJob := initialTestMachine(innerCtx, system)
				if stopTick {
					return
				}
				if running {
					handleManualInteractionTest(innerCtx, whqlJob)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return closeChan
}

// 执行初始化测试机逻辑。
func initialTestMachine(ctx context.Context, system string) (stopTick, running bool, whqlJob *model.WhqlJob) {
	defer func() {
		if p := recover(); p != nil {
			log.Error(ctx, "test machine panic", p, util.GetStackCallers())
		}
	}()

	// 判断测试机是否在测试中。
	finished, running, whqlJob, ok := checkMachineState(ctx)
	if !ok {
		return
	}

	// 回滚测试机器。
	restored, ok := restoreMachine(ctx, finished)
	if !ok {
		stopTick = restored
		return
	}

	// 获取任务。
	job, ok := getJobToInitialMachine(ctx, system)
	if !ok {
		return
	}

	// 生成工作空间。
	workspace, ok := createWorkspace(ctx, job.ID)
	if !ok {
		return
	}

	// 清理工作区。
	defer cleanWorkspace(ctx, &ok, workspace)

	// 下载任务文件。
	sourceFilePath, ok := downloadFileToDisk(ctx, job.ID, job.FileID, workspace)
	if !ok {
		return
	}

	// 解压文件。
	isZipFile, ok := unzipFile(ctx, job.ID, workspace, sourceFilePath)
	if !ok {
		return
	}

	// 找到安装服务程序文件。
	installFilePath, ok := findInstallFile(ctx, job.ID, isZipFile, workspace)
	if !ok {
		return
	}

	// 创建添加证书。
	if ok = addCertificate(ctx, job.ID); !ok {
		return
	}

	// 给驱动程序文件签名。
	if ok = signSystemFiles(ctx, job.ID, isZipFile, workspace); !ok {
		return
	}
	if ok = signSystemFile(ctx, job.ID, isZipFile, sourceFilePath); !ok {
		return
	}

	// 安装系统服务。
	var serviceName string
	if ok = executeInstallServiceFile(ctx, job.ID, isZipFile, installFilePath); !ok {
		return
	}
	if serviceName, ok = createSystemService(ctx, job.ID, isZipFile, sourceFilePath, job.ServiceName); !ok {
		return
	}

	// 记录测试任务信息。
	if ok = persistenceJobInformation(ctx, job.ID, workspace); !ok {
		return
	}

	// 更新任务。
	if ok = pushJobToTesting(ctx, job.ID, serviceName); !ok {
		return
	}

	// 重启机器。
	if ok = rebootMachine(ctx); !ok {
		return
	}

	stopTick = true
	return
}

// 处理人工交互测试。
func handleManualInteractionTest(ctx context.Context, whqlJob *model.WhqlJob) {
	defer func() {
		if p := recover(); p != nil {
			log.Error(ctx, "test machine panic", p, util.GetStackCallers())
		}
	}()

	// 获取所有桌面对话框。
	windowDialogs, ok := listAllWindowDialogs(ctx)
	if !ok {
		return
	}

	for _, v := range windowDialogs {
		switch v {
		case consts.DialogTitleWFPGatherInfo:
			dealDialogTitleWFPGatherInfo(ctx, whqlJob.TestConfig, whqlJob.ServiceName)
		case consts.DialogTitleWFPMAC:
			dealDialogTitleWFPMAC(ctx)
		case consts.DialogTitleWFPIP:
			dealDialogTitleWFPIP(ctx)
		case consts.DialogTitleWFPProtocol:
			dealDialogTitleWFPProtocol(ctx)
		case consts.DialogTitleWFPPort:
			dealDialogTitleWFPPort(ctx)
		case consts.DialogTitleWFPICMP:
			dealDialogTitleWFPICMP(ctx)
		case consts.DialogTitleWFPApp:
			dealDialogTitleWFPApp(ctx)
		case consts.DialogTitleWFPNoDeadlocks:
			dealDialogTitleWFPNoDeadlocks(ctx)
		case consts.DialogTitleWFPPower:
			dealDialogTitleWFPPower(ctx)
		case consts.DialogTitleWFPStream:
			dealDialogTitleWFPStream(ctx)
		case consts.DialogTitleELAM:
			dealDialogTitleELAM(ctx, whqlJob.TestConfig)
		case consts.DialogTitleCMD:
			dealDialogTitleCMD(ctx, whqlJob.TestConfig)
		case consts.DialogTitleTeredo:
			dealDialogTitleTeredo(ctx)
		case "":
			dealDialogTitleBlank(ctx)
		default:
			log.Warn(ctx, "unhandled window dialog", v)
		}
	}
}

// 读取文件，判断测试机是否在测试中。
func checkMachineState(ctx context.Context) (finished, running bool, whqlJob *model.WhqlJob, ok bool) {
	log.Debug(ctx, "check machine state")

	// 读取文件。
	fileBytes, err := os.ReadFile(consts.TestInfoFilePath)
	if err != nil {
		// 文件不存在，没有任务测试。
		if os.IsNotExist(err) {
			return false, false, nil, true
		}
		log.Error(ctx, "failed to read test info file path", err)
		// 等待下次 tick。
		return false, false, nil, false
	}

	// 序列化文件内容。
	var job testJobInfo
	if err = json.Unmarshal(fileBytes, &job); err != nil {
		log.Error(ctx, "failed to unmarshal test info file path", err, fileBytes)
		// 序列化失败，回滚机器。
		return true, false, nil, true
	}

	// 查询任务信息，判断是否在测试中。
	whqlJob, err = httpBackendGetWHQLJobInformation(ctx, job.ID)
	if err != nil {
		log.Error(ctx, "failed to get whql job information", err)
		// 获取任务信息失败，等待下次轮询。
		return false, false, nil, false
	}
	if whqlJob == nil {
		// 任务已不存在。
		return true, false, nil, true
	}

	// 判断任务状态，任务在测试中。
	if util.In(whqlJob.Status, model.WHQLJobStatusInitiallingTestMachine,
		model.WHQLJobStatusFinishInitiallingTestMachine, model.WHQLJobStatusHLKTesting) {
		return false, true, whqlJob, false
	}

	// 任务需要初始化测试。回滚机器。
	if whqlJob.Status == model.WHQLJobStatusWaitingTest {
		return true, false, nil, false
	}

	// 任务测试完毕。回滚机器。
	return true, false, nil, true
}

// 获取测试机器到检查点。
func restoreMachine(ctx context.Context, finished bool) (restored, ok bool) {
	// 还没测试完，不回滚。
	if !finished {
		return false, true
	}

	log.Info(ctx, "restore test machine")

	// 请求宿主机回滚。
	if err := httpHostRestoreMachine(ctx); err != nil {
		log.Error(ctx, "test machine restore failed", err)
		return false, false
	}

	return true, false
}

// 获取任务。
func getJobToInitialMachine(ctx context.Context, system string) (whqlJob *model.WhqlJob, ok bool) {
	log.Debug(ctx, "get a new job to initial test machine")

	// 请求主服务获取任务。
	whqlJob, err := httpBackendGetWHQLJobToInitialTestMachine(ctx, system)
	if err != nil {
		log.Error(ctx, "getting job to initial failed", err)
		return nil, false
	}
	if whqlJob == nil {
		log.Debug(ctx, "no job initial")
		return nil, false
	}
	log.Info(ctx, "initial job", whqlJob.ID, whqlJob.ServiceName, whqlJob.TestTarget)

	return whqlJob, true
}

// 创建工作空间。
func createWorkspace(ctx context.Context, jobID int) (workspace string, ok bool) {
	log.Info(ctx, "create workspace")

	// 创建目录。
	workspace, err := util.GenerateTemporaryDirectory(filepath.Join(cc.ServiceNameHLK, strconv.Itoa(jobID)))
	if err != nil {
		log.Error(ctx, "failed to create workspace", err)
		failJob(ctx, jobID, "创建工作空间失败，%v", err)
		return "", false
	}

	return workspace, true
}

// 清理工作区。
func cleanWorkspace(ctx context.Context, ok *bool, workspace string) {
	// 成功处理了，不清理工作区。
	if *ok {
		return
	}

	log.Info(ctx, "clean workspace", workspace)
	util.RemoveDirectory(ctx, workspace)
}

// 下载文件。
func downloadFileToDisk(ctx context.Context, jobID int, fileID, workspace string) (sourceFilePath string, ok bool) {
	log.Info(ctx, "download file to disk")

	// 从服务器下载文件。
	sourceFilePath, _, err := httpBackendDownloadFile(ctx, fileID, workspace)
	if err != nil {
		log.Error(ctx, "downloading source file failed", err)
		failJob(ctx, jobID, "下载文件失败，%s，%v", fileID, err)
		return "", false
	}

	return sourceFilePath, true
}

// 解压文件。
func unzipFile(ctx context.Context, jobID int, workspace, sourceFilePath string) (isZipFile, ok bool) {
	// 不是 zip 文件不解压。
	if strings.ToLower(filepath.Ext(sourceFilePath)) != cc.ExtensionZIP {
		return false, true
	}

	log.Info(ctx, "unzip file to disk")

	// 解压文件。
	err := util.Unzip(ctx, sourceFilePath, workspace)
	if err != nil {
		log.Error(ctx, "unzipping file failed", err)
		failJob(ctx, jobID, "解压文件失败，%v", err)
		return true, false
	}

	return true, true
}

// 找到安装服务程序文件。
func findInstallFile(ctx context.Context, jobID int, isZipFile bool, directory string) (
	installFilePath string, ok bool) {

	// 不是 zip 包，不需要处理。
	if !isZipFile {
		return "", true
	}

	log.Info(ctx, "find install.bat file")

	// 读取文件夹。
	entries, err := os.ReadDir(directory)
	if err != nil {
		log.Error(ctx, "reading directory failed", err)
		return "", false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if entry.Name() == consts.InstallServiceFileName {
			return filepath.Join(directory, entry.Name()), true
		}
	}

	// 没有找到，结束任务。
	log.Error(ctx, "no install file found, fail this job")
	failJob(ctx, jobID, "安装脚本文件 %s 为找到", consts.InstallServiceFileName)

	return "", false
}

// 给系统添加测试证书。
func addCertificate(ctx context.Context, jobID int) (ok bool) {
	log.Info(ctx, "add a test certificate")

	// 删除旧的证书。
	for {
		command := exec.Command(consts.CertmgrFilePath, `/del`, `/s`, `PrivateCertStore`, `/c`, `/n`, consts.TestCertificateName)
		output := &bytes.Buffer{}
		command.Stdout = output
		command.Stderr = output
		writer, err := command.StdinPipe()
		if err != nil {
			log.Error(ctx, "creating stdin pipe failed", err)
			failJob(ctx, jobID, "删除系统证书失败，%v", err)
			return false
		}
		if err = command.Start(); err != nil {
			log.Error(ctx, "starting command failed", err)
			failJob(ctx, jobID, "删除系统证书失败，%v", err)
			return false
		}
		_, err = writer.Write([]byte("1\r\n"))
		if err != nil {
			log.Error(ctx, "writing command failed", err)
			failJob(ctx, jobID, "删除系统证书失败，%v", err)
			return false
		}
		if err = command.Wait(); err != nil {
			outputString := output.String()
			if strings.Contains(outputString, "Failed to find a certificate to delete") {
				break
			}
			log.Error(ctx, "waiting command failed", err)
			failJob(ctx, jobID, "删除系统证书失败，%v，%s", err, outputString)
			return false
		}
	}

	// 添加新证书。
	outputBytes, err := exec.Command(consts.MakecertFilePath, `/r`, `/pe`, `/ss`, `PrivateCertStore`, `/n`,
		`CN=`+consts.TestCertificateName, `/eku`, consts.TestCertificateEku).CombinedOutput()
	outputBytes = util.GbkToUtf8(outputBytes)
	if err != nil {
		log.Error(ctx, "adding certificate failed", err)
		failJob(ctx, jobID, "添加系统证书失败，%v，%s", err, outputBytes)
		return false
	}
	log.Debug(ctx, "adding certificate succeeded", outputBytes)

	return true
}

// 给驱动程序文件签名。
func signSystemFiles(ctx context.Context, jobID int, isZipFile bool, workspace string) (ok bool) {
	// 不是 zip 包，不处理。
	if !isZipFile {
		return true
	}

	log.Info(ctx, "sign system binary files")

	// 读取出所有 .sys 文件。
	entries, err := os.ReadDir(workspace)
	if err != nil {
		log.Error(ctx, "reading directory failed", err)
		return false
	}
	systemFilePaths := make([]string, 0, len(entries)/2)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.ToLower(filepath.Ext(entry.Name())) == cc.ExtensionSYS {
			systemFilePaths = append(systemFilePaths, filepath.Join(workspace, entry.Name()))
		}
	}

	// 给 .sys 文件签名。
	args := []string{"sign", "/v", "/fd", "sha256", "/s", "PrivateCertStore", "/n", consts.TestCertificateName}
	args = append(args, systemFilePaths...)
	outputBytes, err := exec.Command(consts.SigntoolFilePath, args...).CombinedOutput()
	if err != nil {
		log.Error(ctx, "sign command failed", err, outputBytes)
		failJob(ctx, jobID, "为驱动程序签名失败，%v，%s", err, outputBytes)
		return false
	}
	if !bytes.Contains(outputBytes,
		fmt.Appendf(nil, "Number of files successfully Signed: %d", len(systemFilePaths))) {
		log.Error(ctx, "some system files are not signed", outputBytes)
		failJob(ctx, jobID, "一些驱动程序签名失败，%s", outputBytes)
		return false
	}

	return true
}

// 给驱动程序文件签名。
func signSystemFile(ctx context.Context, jobID int, isZipFile bool, sourceFilePath string) (ok bool) {
	// 是 zip 包，不处理。
	if isZipFile {
		return true
	}

	log.Info(ctx, "sign system binary file")

	// 给 .sys 文件签名。
	outputBytes, err := exec.Command(consts.SigntoolFilePath, "sign", "/v", "/fd", "sha256", "/s", "PrivateCertStore",
		"/n", consts.TestCertificateName, sourceFilePath).CombinedOutput()
	outputBytes = util.GbkToUtf8(outputBytes)
	if err != nil {
		log.Error(ctx, "sign command failed", err, outputBytes)
		failJob(ctx, jobID, "为驱动程序签名失败，%v，%s", err, outputBytes)
		return false
	}
	if !bytes.Contains(outputBytes, []byte("Number of files successfully Signed: 1")) {
		log.Error(ctx, "sign command failed", outputBytes)
		failJob(ctx, jobID, "驱动程序签名失败，%s", outputBytes)
		return false
	}

	return true
}

// 运行安装服务脚本。
func executeInstallServiceFile(ctx context.Context, jobID int, isZipFile bool, installServiceFilePath string) (
	ok bool) {

	// 不是 zip 包，不处理。
	if !isZipFile {
		return true
	}

	log.Info(ctx, "execute install service file")

	// 执行系统命令。
	command := exec.Command(installServiceFilePath)
	path, err := filepath.Abs(filepath.Dir(installServiceFilePath))
	if err != nil {
		log.Error(ctx, "getting absolute file path failed", err)
		failJob(ctx, jobID, "获取文件路径发生错误，%v", err)
		return false
	}
	command.Dir = path
	outputBytes, err := command.CombinedOutput()
	if err != nil {
		log.Error(ctx, "running install service command failed", err, outputBytes)
		failJob(ctx, jobID, "执行 %s 失败，%v，%s", consts.InstallServiceFileName, err, outputBytes)
		return false
	}
	log.Debug(ctx, "running install service succeeded", outputBytes)

	return true
}

// 创建系统服务，并运行。
func createSystemService(ctx context.Context, jobID int, isZipFile bool, systemFilePath, serviceName string) (
	newServiceName string, ok bool) {

	// 是 zip 包，不处理。
	if isZipFile {
		return "", true
	}

	log.Info(ctx, "create system service")

	// 创建服务。
	if len(serviceName) <= 0 {
		serviceName = filepath.Base(systemFilePath)
		serviceName = strings.TrimSuffix(serviceName, filepath.Ext(serviceName))
	}
	quotedServiceName := util.QuoteArguments(serviceName)[0]
	outputBytes, err := exec.Command("sc", "create", quotedServiceName,
		fmt.Sprintf(`binPath=%s`, systemFilePath), "start=system", "error=ignore", "type=kernel",
		"DisplayName="+quotedServiceName).CombinedOutput()
	outputBytes = util.GbkToUtf8(outputBytes)
	if err != nil {
		log.Error(ctx, "creating system service failed", err, outputBytes)
		failJob(ctx, jobID, "创建系统服务失败，%v，%s", err, outputBytes)
		return "", false
	}
	log.Debug(ctx, "creating system service succeeded", outputBytes)

	return serviceName, true
}

// 将任务信息序列化到文件。
func persistenceJobInformation(ctx context.Context, jobID int, workspace string) (ok bool) {
	log.Info(ctx, "persistence job information")

	// 序列化数据。
	info := &testJobInfo{
		ID:        jobID,
		Workspace: workspace,
	}
	jobInfoBytes, err := json.Marshal(&info)
	if err != nil {
		log.Error(ctx, "marshalling job information failed", err)
		failJob(ctx, jobID, "序列化数据发生错误，%v", err)
		return false
	}

	// 将序列化数据写入文件。
	fileStream, err := os.OpenFile(consts.TestInfoFilePath, os.O_CREATE|os.O_WRONLY, cc.FileMode)
	if err != nil {
		log.Error(ctx, "opening test info file failed", err)
		failJob(ctx, jobID, "记录测试信息发生错误，%v", err)
		return false
	}
	defer util.CloseIO(ctx, fileStream)
	written, err := io.Copy(fileStream, bytes.NewBuffer(jobInfoBytes))
	if err != nil {
		log.Error(ctx, "failed to write file", err)
		failJob(ctx, jobID, "记录测试信息发生错误，%v", err)
		return false
	}
	if written != int64(len(jobInfoBytes)) {
		log.Error(ctx, "the size of the bytes written does not meet expectations", len(jobInfoBytes), written)
		failJob(ctx, jobID, "记录测试信息发生错误，数据未完全保存：%v", err)
		return false
	}

	return true
}

// 更新任务。
func pushJobToTesting(ctx context.Context, jobID int, serviceName string) (ok bool) {
	log.Info(ctx, "success initial test machine")

	// 请求主服务。
	err := httpBackendUpdateWHQLJob(ctx, &bp.WindowsInternalUpdateWHQLJobReq{
		JobID:           jobID,
		AppendLog:       formatJobLog(log.LevelInfo, "初始化测试机器完毕"),
		Status:          model.WHQLJobStatusFinishInitiallingTestMachine,
		TestMachineName: getMachineName(),
		ServiceName:     serviceName,
	})
	if err != nil {
		log.Error(ctx, "push job to testing failed", err)
		return false
	}

	return true
}

// 重启机器。
func rebootMachine(ctx context.Context) (ok bool) {
	log.Info(ctx, "reboot machine")

	// 运行系统命令。
	outputBytes, err := exec.Command("cmd", "/C", "shutdown", "/r", "/t", "10").CombinedOutput()
	if err != nil {
		log.Error(ctx, "reboot machine failed", err, outputBytes)
		return false
	}
	log.Debug(ctx, "reboot machine succeeded", outputBytes)

	return true
}

// 获取所有桌面对话框。
func listAllWindowDialogs(ctx context.Context) (windowDialogs []string, ok bool) {
	log.Debug(ctx, "list all window dialogs")

	windowDialogs, err := listAllWindows(ctx)
	if err != nil {
		log.Error(ctx, "list all window dialogs failed", err)
		return nil, false
	}
	if len(windowDialogs) <= 0 {
		return nil, false
	}

	return windowDialogs, true
}

// 处理交互对话框。
func dealDialogTitleWFPGatherInfo(ctx context.Context, testConfig, serviceName string) {
	log.Info(ctx, "deal dialog title", consts.DialogTitleWFPGatherInfo)

	// 关闭对话框。
	err := closeWindow(ctx, consts.DialogTitleWFPInfo, 0)
	if err != nil {
		log.Error(ctx, "close window failed", err)
	}

	// 处理配置。
	var config cc.WHQLConfig
	err = json.Unmarshal([]byte(testConfig), &config)
	if err != nil {
		log.Warn(ctx, "unmarshalling config failed", err)
	}
	bs, _ := json.Marshal(config.WFPConfig)
	var stringToAny map[string]any
	_ = json.Unmarshal(bs, &stringToAny)
	stringToString := make(map[string]string, len(stringToAny))
	for k, v := range stringToAny {
		switch v {
		case cc.WHQLConfigOn:
			stringToString[k] = "1"
		case cc.WHQLConfigOff:
			stringToString[k] = "0"
		default:
			stringToString[k] = fmt.Sprint(v)
		}
	}
	for i := 0; i < len(consts.DefaultWFPLogo)-1; i += 2 {
		key := fmt.Sprint(consts.DefaultWFPLogo[i])
		if _, ok := stringToString[key]; !ok {
			stringToString[key] = fmt.Sprint(consts.DefaultWFPLogo[i+1])
		}
	}
	if _, ok := stringToString["DriverName"]; !ok {
		stringToString["DriverName"] = fmt.Sprintf(`"%s"`, serviceName)
	}
	stringToString["RunBy"] = fmt.Sprintf(`"%s"`, cc.ServiceNameHLK)
	stringToString["UseAnswerFile"] = "1"
	var configString strings.Builder
	for k, v := range stringToString {
		_, _ = fmt.Fprintf(&configString, "%s = %s;\r\n", k, v)
	}
	err = os.WriteFile(consts.WFPLogoTxtFilePath, []byte(configString.String()), cc.FileMode)
	if err != nil {
		log.Error(ctx, "failed to write wfp configuration file", err)
		return
	}
	err = os.WriteFile(consts.WFPLogoAnswerFilePath, []byte(consts.DefaultWFPAnswerFileData), cc.FileMode)
	if err != nil {
		log.Error(ctx, "failed to write wfp answer file", err)
		return
	}

	// 点击 OK 按钮。
	err = clickButton(ctx, consts.DialogTitleWFPGatherInfo, "OK", 0)
	if err != nil {
		log.Error(ctx, "click button failed", err)
	}
}

// 处理交互对话框。
func dealDialogTitleWFPMAC(ctx context.Context) {
	log.Info(ctx, "deal dialog title", consts.DialogTitleWFPMAC)

	// 获取对话框。
	dialogs, err := getDialogs(ctx, consts.DialogTitleWFPMAC)
	if err != nil {
		log.Error(ctx, "get dialogs failed", err)
		return
	}
	var firstNonBlankDialog string
	var firstNonBlankDialogIndex int
	for i, v := range dialogs {
		if len(v) > 0 {
			firstNonBlankDialog = v
			firstNonBlankDialogIndex = util.Atoi(i)
			break
		}
	}
	messages := strings.Split(firstNonBlankDialog, "\n")
	messages = util.ListTo(messages, func(e string) string { return util.TrimBlank(e) })

	// 获取防火墙参数。
	action, direct, needRemove, localMAC, remoteMAC := "", "", false, "", ""
	for _, v := range messages {
		if strings.Contains(v, "PERMIT") {
			action = "allow"
		}
		if strings.Contains(v, "BLOCK") {
			action = "block"
		}
		if strings.Contains(v, "outbound") {
			direct = "out"
		}
		if strings.Contains(v, "inbound") {
			direct = "in"
		}
		if strings.Contains(v, "Remote") && strings.Contains(v, "Address") {
			remoteMAC = macMatcher.FindString(v)
		}
		if strings.Contains(v, "Local") && strings.Contains(v, "Address") {
			localMAC = macMatcher.FindString(v)
		}
		if strings.Contains(v, "remove") && strings.Contains(v, "configuration") &&
			strings.Contains(v, "previous step") {
			needRemove = true
			break
		}
	}

	// 删除防火墙规则。
	if needRemove {
		if err = util.DeleteWindowsFirewallRule(consts.FirewallRuleName); err != nil {
			log.Error(ctx, "delete firewall rule failed", err)
		}
	} else {
		// 添加防火墙规则。
		err = util.CreateWindowsFirewallRule(&util.CreateWindowsFirewallRuleParams{
			Name:   consts.FirewallRuleName,
			Direct: direct,
			Action: action,
		})
		if err != nil {
			log.Error(ctx, "create firewall rule failed", err)
		}
	}
	_, _ = localMAC, remoteMAC

	// 点击按钮关闭对话框。
	err = clickButton(ctx, consts.DialogTitleWFPMAC, "OK", firstNonBlankDialogIndex)
	if err != nil {
		log.Error(ctx, "click button failed", err)
	}
}

// 处理交互对话框。
func dealDialogTitleWFPIP(ctx context.Context) {
	log.Info(ctx, "deal dialog title", consts.DialogTitleWFPIP)

	// 获取对话框。
	dialogs, err := getDialogs(ctx, consts.DialogTitleWFPIP)
	if err != nil {
		log.Error(ctx, "get dialogs failed", err)
		return
	}
	var firstNonBlankDialog string
	var firstNonBlankDialogIndex int
	for i, v := range dialogs {
		if len(v) > 0 {
			firstNonBlankDialog = v
			firstNonBlankDialogIndex = util.Atoi(i)
			break
		}
	}
	messages := strings.Split(firstNonBlankDialog, "\n")
	messages = util.ListTo(messages, func(e string) string { return util.TrimBlank(e) })

	// 获取防火墙参数。
	action, direct, remoteIP, localIP, needRemove, isIPv6 := "", "", "", "", false, false
	for _, v := range messages {
		if strings.Contains(v, "PERMIT") {
			action = "allow"
		}
		if strings.Contains(v, "BLOCK") {
			action = "block"
		}
		if strings.Contains(v, "IPv6") {
			isIPv6 = true
		}
		if strings.Contains(v, "outbound") {
			direct = "out"
		}
		if strings.Contains(v, "inbound") {
			direct = "in"
		}
		if strings.Contains(v, "Remote") && strings.Contains(v, "Address") {
			if isIPv6 {
				remoteIP = ipv6Matcher.FindString(v)
			} else {
				remoteIP = ipv4Matcher.FindString(v)
			}
		}
		if strings.Contains(v, "Local") && strings.Contains(v, "Address") {
			if isIPv6 {
				localIP = ipv6Matcher.FindString(v)
			} else {
				localIP = ipv4Matcher.FindString(v)
			}
		}
		if strings.Contains(v, "remove") && strings.Contains(v, "configuration") &&
			strings.Contains(v, "previous step") {
			needRemove = true
			break
		}
	}

	// 删除防火墙规则。
	if needRemove {
		if err = util.DeleteWindowsFirewallRule(consts.FirewallRuleName); err != nil {
			log.Error(ctx, "delete firewall rule failed", err)
		}
	} else {
		// 添加防火墙规则。
		err = util.CreateWindowsFirewallRule(&util.CreateWindowsFirewallRuleParams{
			Name:     consts.FirewallRuleName,
			Direct:   direct,
			LocalIP:  localIP,
			RemoteIP: remoteIP,
			Action:   action,
		})
		if err != nil {
			log.Error(ctx, "create firewall rule failed", err)
		}
	}

	// 点击按钮关闭对话框。
	err = clickButton(ctx, consts.DialogTitleWFPIP, "OK", firstNonBlankDialogIndex)
	if err != nil {
		log.Error(ctx, "click button failed", err)
	}
}

// 处理交互对话框。
func dealDialogTitleWFPProtocol(ctx context.Context) {
	log.Info(ctx, "deal dialog title", consts.DialogTitleWFPProtocol)

	// 获取对话框。
	dialogs, err := getDialogs(ctx, consts.DialogTitleWFPProtocol)
	if err != nil {
		log.Error(ctx, "get dialogs failed", err)
		return
	}
	var firstNonBlankDialog string
	var firstNonBlankDialogIndex int
	for i, v := range dialogs {
		if len(v) > 0 {
			firstNonBlankDialog = v
			firstNonBlankDialogIndex = util.Atoi(i)
			break
		}
	}
	messages := strings.Split(firstNonBlankDialog, "\n")
	messages = util.ListTo(messages, func(e string) string { return util.TrimBlank(e) })

	// 获取防火墙参数。
	action, direct, needRemove, protocol := "", "", false, ""
	for _, v := range messages {
		if strings.Contains(v, "PERMIT") {
			action = "allow"
		}
		if strings.Contains(v, "BLOCK") {
			action = "block"
		}
		if strings.Contains(v, "outbound") {
			direct = "out"
		}
		if strings.Contains(v, "inbound") {
			direct = "in"
		}
		if strings.Contains(v, "Protocol") {
			if strings.Contains(v, "IP Raw (255)") {
				protocol = "255"
			}
			if strings.Contains(v, "TCP (6)") {
				protocol = "6"
			}
			if strings.Contains(v, "UDP (17)") {
				protocol = "17"
			}
		}
		if strings.Contains(v, "remove") && strings.Contains(v, "configuration") &&
			strings.Contains(v, "previous step") {
			needRemove = true
			break
		}
	}

	// 删除防火墙规则。
	if needRemove {
		if err = util.DeleteWindowsFirewallRule(consts.FirewallRuleName); err != nil {
			log.Error(ctx, "delete firewall rule failed", err)
		}
	} else {
		// 添加防火墙规则。
		err = util.CreateWindowsFirewallRule(&util.CreateWindowsFirewallRuleParams{
			Name:     consts.FirewallRuleName,
			Direct:   direct,
			Protocol: protocol,
			Action:   action,
		})
		if err != nil {
			log.Error(ctx, "create firewall rule failed", err)
		}
	}

	// 点击按钮关闭对话框。
	err = clickButton(ctx, consts.DialogTitleWFPProtocol, "OK", firstNonBlankDialogIndex)
	if err != nil {
		log.Error(ctx, "click button failed", err)
	}
}

// 处理交互对话框。
func dealDialogTitleWFPPort(ctx context.Context) {
	log.Info(ctx, "deal dialog title", consts.DialogTitleWFPPort)

	// 获取对话框。
	dialogs, err := getDialogs(ctx, consts.DialogTitleWFPPort)
	if err != nil {
		log.Error(ctx, "get dialogs failed", err)
		return
	}
	var firstNonBlankDialog string
	var firstNonBlankDialogIndex int
	for i, v := range dialogs {
		if len(v) > 0 {
			firstNonBlankDialog = v
			firstNonBlankDialogIndex = util.Atoi(i)
			break
		}
	}
	messages := strings.Split(firstNonBlankDialog, "\n")
	messages = util.ListTo(messages, func(e string) string { return util.TrimBlank(e) })

	// 获取防火墙参数。
	action, direct, needRemove, protocol, localPort, remotePort := "", "", false, "", "", ""
	for _, v := range messages {
		if strings.Contains(v, "PERMIT") {
			action = "allow"
		}
		if strings.Contains(v, "BLOCK") {
			action = "block"
		}
		if strings.Contains(v, "outbound") {
			direct = "out"
		}
		if strings.Contains(v, "inbound") {
			direct = "in"
		}
		if strings.Contains(v, "Remote") && strings.Contains(v, "Port") {
			remotePort = numMatcher.FindString(v)
		}
		if strings.Contains(v, "Local") && strings.Contains(v, "Port") {
			localPort = numMatcher.FindString(v)
		}
		if strings.Contains(v, "Protocol") {
			if strings.Contains(v, "IP Raw (255)") {
				protocol = "255"
			}
			if strings.Contains(v, "TCP (6)") {
				protocol = "6"
			}
			if strings.Contains(v, "UDP (17)") {
				protocol = "17"
			}
		}
		if strings.Contains(v, "remove") && strings.Contains(v, "configuration") &&
			strings.Contains(v, "previous step") {
			needRemove = true
			break
		}
	}

	// 删除防火墙规则。
	if needRemove {
		if err = util.DeleteWindowsFirewallRule(consts.FirewallRuleName); err != nil {
			log.Error(ctx, "delete firewall rule failed", err)
		}
	} else {
		// 添加防火墙规则。
		err = util.CreateWindowsFirewallRule(&util.CreateWindowsFirewallRuleParams{
			Name:       consts.FirewallRuleName,
			Direct:     direct,
			Protocol:   protocol,
			Action:     action,
			LocalPort:  localPort,
			RemotePort: remotePort,
		})
		if err != nil {
			log.Error(ctx, "create firewall rule failed", err)
		}
	}

	// 点击按钮关闭对话框。
	err = clickButton(ctx, consts.DialogTitleWFPPort, "OK", firstNonBlankDialogIndex)
	if err != nil {
		log.Error(ctx, "click button failed", err)
	}
}

// 处理交互对话框。
func dealDialogTitleWFPICMP(ctx context.Context) {
	log.Info(ctx, "deal dialog title", consts.DialogTitleWFPICMP)

	// 获取对话框。
	dialogs, err := getDialogs(ctx, consts.DialogTitleWFPICMP)
	if err != nil {
		log.Error(ctx, "get dialogs failed", err)
		return
	}
	var firstNonBlankDialog string
	var firstNonBlankDialogIndex int
	for i, v := range dialogs {
		if len(v) > 0 {
			firstNonBlankDialog = v
			firstNonBlankDialogIndex = util.Atoi(i)
			break
		}
	}
	messages := strings.Split(firstNonBlankDialog, "\n")
	messages = util.ListTo(messages, func(e string) string { return util.TrimBlank(e) })

	// 获取防火墙参数。
	action, direct, needRemove, icmpVer, typ, code := "", "", false, "", "", ""
	for _, v := range messages {
		if strings.Contains(v, "PERMIT") {
			action = "allow"
		}
		if strings.Contains(v, "BLOCK") {
			action = "block"
		}
		if strings.Contains(v, "ICMPv6") {
			icmpVer = "icmpv6"
		}
		if strings.Contains(v, "ICMPv4") {
			icmpVer = "icmpv4"
		}
		if strings.Contains(v, "outbound") {
			direct = "out"
		}
		if strings.Contains(v, "inbound") {
			direct = "in"
		}
		if strings.Contains(v, "Type:") {
			typ = numMatcher.FindString(v)
		}
		if strings.Contains(v, "Code:") {
			code = numMatcher.FindString(v)
		}
		if strings.Contains(v, "remove") && strings.Contains(v, "configuration") &&
			strings.Contains(v, "previous step") {
			needRemove = true
			break
		}
	}

	// 删除防火墙规则。
	if needRemove {
		if err = util.DeleteWindowsFirewallRule(consts.FirewallRuleName); err != nil {
			log.Error(ctx, "delete firewall rule failed", err)
		}
	} else {
		// 添加防火墙规则。
		err = util.CreateWindowsFirewallRule(&util.CreateWindowsFirewallRuleParams{
			Name:     consts.FirewallRuleName,
			Direct:   direct,
			Action:   action,
			Protocol: fmt.Sprintf("%s:%s,%s", icmpVer, typ, code),
		})
		if err != nil {
			log.Error(ctx, "create firewall rule failed", err)
		}
	}

	// 点击按钮关闭对话框。
	err = clickButton(ctx, consts.DialogTitleWFPICMP, "OK", firstNonBlankDialogIndex)
	if err != nil {
		log.Error(ctx, "click button failed", err)
	}
}

// 处理交互对话框。
func dealDialogTitleWFPApp(ctx context.Context) {
	log.Info(ctx, "deal dialog title", consts.DialogTitleWFPApp)

	// 获取对话框。
	dialogs, err := getDialogs(ctx, consts.DialogTitleWFPApp)
	if err != nil {
		log.Error(ctx, "get dialogs failed", err)
		return
	}
	var firstNonBlankDialog string
	var firstNonBlankDialogIndex int
	for i, v := range dialogs {
		if len(v) > 0 {
			firstNonBlankDialog = v
			firstNonBlankDialogIndex = util.Atoi(i)
			break
		}
	}
	messages := strings.Split(firstNonBlankDialog, "\n")
	messages = util.ListTo(messages, func(e string) string { return util.TrimBlank(e) })

	// 获取防火墙参数。
	action, direct, needRemove, isIPv6, localIP, remoteIP := "allow", "", false, false, "", ""
	for _, v := range messages {
		if strings.Contains(v, "IPv6") {
			isIPv6 = true
		}
		if strings.Contains(v, "Outbound") {
			direct = "out"
		}
		if strings.Contains(v, "Inbound") {
			direct = "in"
		}
		if strings.Contains(v, "Source") && strings.Contains(v, "Address") {
			index := strings.Index(v, "Source")
			sub := v[index:]
			if isIPv6 {
				switch direct {
				case "out":
					localIP = ipv6Matcher.FindString(sub)
				case "in":
					remoteIP = ipv6Matcher.FindString(sub)
				}
			} else {
				switch direct {
				case "out":
					localIP = ipv4Matcher.FindString(sub)
				case "in":
					remoteIP = ipv4Matcher.FindString(sub)
				}
			}
		}
		if strings.Contains(v, "Destination") && strings.Contains(v, "Address") {
			index := strings.Index(v, "Destination")
			sub := v[index:]
			if isIPv6 {
				switch direct {
				case "out":
					remoteIP = ipv6Matcher.FindString(sub)
				case "in":
					localIP = ipv6Matcher.FindString(sub)
				}
			} else {
				switch direct {
				case "out":
					remoteIP = ipv4Matcher.FindString(sub)
				case "in":
					localIP = ipv4Matcher.FindString(sub)
				}
			}
		}
		if strings.Contains(v, "remove") && strings.Contains(v, "configuration") &&
			strings.Contains(v, "previous step") {
			needRemove = true
			break
		}
	}

	// 删除防火墙规则。
	if needRemove {
		if err = util.DeleteWindowsFirewallRule(consts.FirewallRuleName); err != nil {
			log.Error(ctx, "delete firewall rule failed", err)
		}
	} else {
		// 添加防火墙规则。
		err = util.CreateWindowsFirewallRule(&util.CreateWindowsFirewallRuleParams{
			Name:     consts.FirewallRuleName,
			Direct:   direct,
			Action:   action,
			LocalIP:  localIP,
			RemoteIP: remoteIP,
		})
		if err != nil {
			log.Error(ctx, "create firewall rule failed", err)
		}
	}

	// 点击按钮关闭对话框。
	err = clickButton(ctx, consts.DialogTitleWFPApp, "OK", firstNonBlankDialogIndex)
	if err != nil {
		log.Error(ctx, "click button failed", err)
	}
}

// 处理交互对话框。
func dealDialogTitleWFPNoDeadlocks(ctx context.Context) {
	log.Info(ctx, "deal dialog title", consts.DialogTitleWFPNoDeadlocks)

	// 获取对话框。
	dialogs, err := getDialogs(ctx, consts.DialogTitleWFPNoDeadlocks)
	if err != nil {
		log.Error(ctx, "get dialogs failed", err)
		return
	}
	var firstNonBlankDialog string
	var firstNonBlankDialogIndex int
	for i, v := range dialogs {
		if len(v) > 0 {
			firstNonBlankDialog = v
			firstNonBlankDialogIndex = util.Atoi(i)
			break
		}
	}
	messages := strings.Split(firstNonBlankDialog, "\n")
	messages = util.ListTo(messages, func(e string) string { return util.TrimBlank(e) })

	// 获取防火墙参数。
	action, direct, needRemove, program := "", "", false, ""
	for _, v := range messages {
		if strings.Contains(v, "PERMIT") {
			action = "allow"
		}
		if strings.Contains(v, "BLOCK") {
			action = "block"
		}
		if strings.Contains(v, "outbound") {
			direct = "out"
		}
		if strings.Contains(v, "inbound") {
			direct = "in"
		}
		if strings.Contains(v, "Application:") {
			program = strings.Replace(v, "Application:", "", 1)
			program = strings.TrimSpace(program)
			program = strings.Trim(program, "\t")
			program = strings.Trim(program, "\r")
		}
		if strings.Contains(v, "remove") && strings.Contains(v, "configuration") &&
			strings.Contains(v, "previous step") {
			needRemove = true
			break
		}
	}

	// 删除防火墙规则。
	if needRemove {
		if err = util.DeleteWindowsFirewallRule(consts.FirewallRuleName); err != nil {
			log.Error(ctx, "delete firewall rule failed", err)
		}
	} else {
		// 添加防火墙规则。
		err = util.CreateWindowsFirewallRule(&util.CreateWindowsFirewallRuleParams{
			Name:    consts.FirewallRuleName,
			Program: program,
			Direct:  direct,
			Action:  action,
		})
		if err != nil {
			log.Error(ctx, "create firewall rule failed", err)
		}
	}

	// 点击按钮关闭对话框。
	err = clickButton(ctx, consts.DialogTitleWFPNoDeadlocks, "OK", firstNonBlankDialogIndex)
	if err != nil {
		log.Error(ctx, "click button failed", err)
	}
}

// 处理交互对话框。
func dealDialogTitleWFPPower(ctx context.Context) {
	log.Info(ctx, "deal dialog title", consts.DialogTitleWFPPower)

	// 获取对话框。
	dialogs, err := getDialogs(ctx, consts.DialogTitleWFPPower)
	if err != nil {
		log.Error(ctx, "get dialogs failed", err)
		return
	}
	var firstNonBlankDialog string
	var firstNonBlankDialogIndex int
	for i, v := range dialogs {
		if len(v) > 0 {
			firstNonBlankDialog = v
			firstNonBlankDialogIndex = util.Atoi(i)
			break
		}
	}
	messages := strings.Split(firstNonBlankDialog, "\n")
	messages = util.ListTo(messages, func(e string) string { return util.TrimBlank(e) })

	// 获取防火墙参数。
	action, direct, remoteIP, localIP, needRemove, isIPv6, protocol, program := "", "", "", "", false, false, "", ""
	for _, v := range messages {
		if strings.Contains(v, "PERMIT") {
			action = "allow"
		}
		if strings.Contains(v, "BLOCK") {
			action = "block"
		}
		if strings.Contains(v, "IPv6") {
			isIPv6 = true
		}
		if strings.Contains(v, "outbound") {
			direct = "out"
		}
		if strings.Contains(v, "inbound") {
			direct = "in"
		}
		if strings.Contains(v, "Application:") {
			program = strings.Replace(v, "Application:", "", 1)
			program = strings.TrimSpace(program)
			program = strings.Trim(program, "\t")
			program = strings.Trim(program, "\r")
		}
		if strings.Contains(v, "Protocol") {
			if strings.Contains(v, "IP Raw (255)") {
				protocol = "255"
			}
			if strings.Contains(v, "TCP (6)") {
				protocol = "6"
			}
			if strings.Contains(v, "UDP (17)") {
				protocol = "17"
			}
		}
		if strings.Contains(v, "Remote") && strings.Contains(v, "Address") {
			if isIPv6 {
				remoteIP = ipv6Matcher.FindString(v)
			} else {
				remoteIP = ipv4Matcher.FindString(v)
			}
		}
		if strings.Contains(v, "Local") && strings.Contains(v, "Address") {
			if isIPv6 {
				localIP = ipv6Matcher.FindString(v)
			} else {
				localIP = ipv4Matcher.FindString(v)
			}
		}
		if strings.Contains(v, "remove") && strings.Contains(v, "configuration") &&
			strings.Contains(v, "previous step") {
			needRemove = true
			break
		}
	}

	// 删除防火墙规则。
	if needRemove {
		if err = util.DeleteWindowsFirewallRule(consts.FirewallRuleName); err != nil {
			log.Error(ctx, "delete firewall rule failed", err)
		}
	} else {
		// 添加防火墙规则。
		err = util.CreateWindowsFirewallRule(&util.CreateWindowsFirewallRuleParams{
			Name:     consts.FirewallRuleName,
			Program:  program,
			Direct:   direct,
			Action:   action,
			LocalIP:  localIP,
			RemoteIP: remoteIP,
			Protocol: protocol,
		})
		if err != nil {
			log.Error(ctx, "create firewall rule failed", err)
		}
	}

	// 点击按钮关闭对话框。
	err = clickButton(ctx, consts.DialogTitleWFPPower, "OK", firstNonBlankDialogIndex)
	if err != nil {
		log.Error(ctx, "click button failed", err)
	}
}

// 处理交互对话框。
func dealDialogTitleWFPStream(ctx context.Context) {
	log.Info(ctx, "deal dialog title", consts.DialogTitleWFPStream)

	// 获取对话框。
	dialogs, err := getDialogs(ctx, consts.DialogTitleWFPStream)
	if err != nil {
		log.Error(ctx, "get dialogs failed", err)
		return
	}
	var firstNonBlankDialog string
	var firstNonBlankDialogIndex int
	for i, v := range dialogs {
		if len(v) > 0 {
			firstNonBlankDialog = v
			firstNonBlankDialogIndex = util.Atoi(i)
			break
		}
	}
	messages := strings.Split(firstNonBlankDialog, "\n")
	messages = util.ListTo(messages, func(e string) string { return util.TrimBlank(e) })

	// 获取防火墙参数。
	action, needRemove, isIPv6, localIP, remoteIP, remotePort, localPort, protocol :=
		"", false, false, "", "", "", "", ""
	for _, v := range messages {
		if strings.Contains(v, "PERMIT") {
			action = "allow"
		}
		if strings.Contains(v, "BLOCK") {
			action = "block"
		}
		if strings.Contains(v, "IPv6") {
			isIPv6 = true
		}
		if strings.Contains(v, "Local") && strings.Contains(v, "Address") {
			if isIPv6 {
				localIP = ipv6Matcher.FindString(v)
			} else {
				localIP = ipv4Matcher.FindString(v)
			}
		}
		if strings.Contains(v, "Remote") && strings.Contains(v, "Port") {
			remotePort = numMatcher.FindString(v)
		}
		if strings.Contains(v, "Local") && strings.Contains(v, "Port") {
			localPort = numMatcher.FindString(v)
		}
		if strings.Contains(v, "Protocol") {
			if strings.Contains(v, "IP Raw (255)") {
				protocol = "255"
			}
			if strings.Contains(v, "TCP (6)") {
				protocol = "6"
			}
			if strings.Contains(v, "UDP (17)") {
				protocol = "17"
			}
		}
		if strings.Contains(v, "remove") && strings.Contains(v, "configuration") &&
			strings.Contains(v, "previous step") {
			needRemove = true
			break
		}
	}

	// 删除防火墙规则。
	if needRemove {
		if err = util.DeleteWindowsFirewallRule(consts.FirewallRuleName); err != nil {
			log.Error(ctx, "delete firewall rule failed", err)
		}
	} else {
		// 添加防火墙规则。
		err = util.CreateWindowsFirewallRule(&util.CreateWindowsFirewallRuleParams{
			Name:       consts.FirewallRuleName,
			Protocol:   protocol,
			Direct:     "in",
			LocalIP:    localIP,
			RemoteIP:   remoteIP,
			Action:     action,
			LocalPort:  localPort,
			RemotePort: remotePort,
		})
		if err != nil {
			log.Error(ctx, "create firewall rule failed", err)
		}
		err = util.CreateWindowsFirewallRule(&util.CreateWindowsFirewallRuleParams{
			Name:       consts.FirewallRuleName,
			Protocol:   protocol,
			Direct:     "out",
			LocalIP:    localIP,
			RemoteIP:   remoteIP,
			Action:     action,
			LocalPort:  localPort,
			RemotePort: remotePort,
		})
		if err != nil {
			log.Error(ctx, "create firewall rule failed", err)
		}
	}

	// 点击按钮关闭对话框。
	err = clickButton(ctx, consts.DialogTitleWFPStream, "OK", firstNonBlankDialogIndex)
	if err != nil {
		log.Error(ctx, "click button failed", err)
	}
}

// 处理交互对话框。
func dealDialogTitleELAM(ctx context.Context, testConfig string) {
	log.Info(ctx, "deal dialog title", consts.DialogTitleELAM)

	// 获取配置。
	var config cc.WHQLConfig
	err := json.Unmarshal([]byte(testConfig), &config)
	if err != nil {
		log.Warn(ctx, "unmarshalling config failed", err)
	}
	button := "Yes"
	switch config.ELAMConfig.IsWdBootMVIMember {
	case cc.WHQLConfigOff:
		button = "No"
	case cc.WHQLConfigOn:
		button = "Yes"
	}

	// 点击按钮。
	err = clickButton(ctx, consts.DialogTitleELAM, button, 0)
	if err != nil {
		log.Error(ctx, "click button failed", err)
	}
}

// 处理交互对话框。
func dealDialogTitleCMD(ctx context.Context, testConfig string) {
	log.Info(ctx, "deal dialog title", consts.DialogTitleCMD)

	index := 0
	for {
		index++

		// 获取 cmd 内容。
		content, err := getCmdContent(ctx, index)
		if err != nil {
			log.Error(ctx, "get cmd content failed", err)
			continue
		}
		if len(content) <= 0 {
			continue
		}

		// 根据内容操作。
		switch {
		case strings.Contains(content, consts.CMDContentAudioCodec):
			// 获取配置。
			var config cc.WHQLConfig
			err = json.Unmarshal([]byte(testConfig), &config)
			if err != nil {
				log.Warn(ctx, "unmarshalling config failed", err)
			}

			// 输入到 cmd 框中。
			key := "{Enter}"
			if config.AudioCodecVerifyAudioEffectsDiscovery {
				key += "Y{Enter}"
			} else {
				key += "N{Enter}"
			}
			err = sendCmdContent(ctx, key, index)
			if err != nil {
				log.Error(ctx, "send cmd content failed", err)
			}
		case strings.Contains(content, consts.CMDContentHSP):
			// 获取配置。
			var config cc.WHQLConfig
			err = json.Unmarshal([]byte(testConfig), &config)
			if err != nil {
				log.Warn(ctx, "unmarshalling config failed", err)
			}

			// 输入到 cmd 框中。
			key := "N{Enter}"
			if config.IsHSPCompatibility {
				key = "Y{Enter}"
			}
			err = sendCmdContent(ctx, key, index)
			if err != nil {
				log.Error(ctx, "send cmd content failed", err)
			}
		default:
			log.Warn(ctx, "unhandled cmd content", content)
		}

		return
	}
}

// 处理交互对话框。
func dealDialogTitleTeredo(ctx context.Context) {
	log.Info(ctx, "deal dialog title", consts.DialogTitleTeredo)

	// 获取对话框。
	dialogs, err := getDialogs(ctx, consts.DialogTitleTeredo)
	if err != nil {
		log.Error(ctx, "get dialogs failed", err)
		return
	}
	var firstNonBlankDialog string
	var firstNonBlankDialogIndex int
	for i, v := range dialogs {
		if len(v) > 0 {
			firstNonBlankDialog = v
			firstNonBlankDialogIndex = util.Atoi(i)
			break
		}
	}

	// 区分处理。
	switch {
	case strings.Contains(firstNonBlankDialog, "remove"):
		// 删除防火墙规则。
		if err = util.DeleteWindowsFirewallRule(consts.FirewallRuleName); err != nil {
			log.Error(ctx, "delete firewall rule failed", err)
		}
	case strings.Contains(firstNonBlankDialog, "7000") && strings.Contains(firstNonBlankDialog, "5001-65535"):
		// 添加防火墙规则。
		err = util.CreateWindowsFirewallRule(&util.CreateWindowsFirewallRuleParams{
			Name:      consts.FirewallRuleName,
			Protocol:  "udp",
			Direct:    "in",
			Action:    "allow",
			LocalPort: "7000",
		})
		if err != nil {
			log.Error(ctx, "create firewall rule failed", err)
		}
		err = util.CreateWindowsFirewallRule(&util.CreateWindowsFirewallRuleParams{
			Name:      consts.FirewallRuleName,
			Protocol:  "udp",
			Direct:    "out",
			Action:    "allow",
			LocalPort: "5001-65535",
		})
		if err != nil {
			log.Error(ctx, "create firewall rule failed", err)
		}
	case strings.Contains(firstNonBlankDialog, "7000"):
		// 添加防火墙规则。
		err = util.CreateWindowsFirewallRule(&util.CreateWindowsFirewallRuleParams{
			Name:      consts.FirewallRuleName,
			Protocol:  "udp",
			Direct:    "in",
			Action:    "allow",
			LocalPort: "7000",
		})
		if err != nil {
			log.Error(ctx, "create firewall rule failed", err)
		}
	}

	// 点击按钮关闭对话框。
	err = clickButton(ctx, consts.DialogTitleTeredo, "OK", firstNonBlankDialogIndex)
	if err != nil {
		log.Error(ctx, "click button failed", err)
	}
}

// 处理交互对话框。
func dealDialogTitleBlank(ctx context.Context) {
	log.Info(ctx, "deal blank dialog title")

	// 获取对话框。
	dialogs, err := getDialogs(ctx, consts.DialogTitleTeredo)
	if err != nil {
		log.Error(ctx, "get dialogs failed", err)
		return
	}

	// 区分处理。
	for i, v := range dialogs {
		index := util.Atoi(i)

		// TransitionTechnologies_Tests 中的一个对话框处理删除防火墙规则
		if strings.Contains(v, " remove ") && strings.Contains(v, "Teredo") {
			// 删除防火墙。
			if err = util.DeleteWindowsFirewallRule(consts.FirewallRuleName); err != nil {
				log.Error(ctx, "delete firewall rule failed", err)
			}

			// 关闭对话框。
			err = clickButton(ctx, "", "OK", index)
			if err != nil {
				log.Error(ctx, "click button failed", err)
			}
		}
	}
}
