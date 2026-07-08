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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
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

type testJobInfo2 struct {
	Target          string            `json:"target,omitempty"`
	MachinePoolName string            `json:"machinePoolName,omitempty"`
	TestMachineName string            `json:"testMachineName,omitempty"`
	ProjectName     string            `json:"projectName,omitempty"`
	Tests           []*targetTestInfo `json:"tests,omitempty"`
	TestIDs         []string          `json:"testIDs,omitempty"`
}

type targetTestInfo struct {
	ID            string        `json:"id,omitempty"`
	TestType      string        `json:"testType,omitempty"`
	Name          string        `json:"name,omitempty"`
	Selected      bool          `json:"selected,omitempty"`
	EstimatedTime time.Duration `json:"estimatedTime,omitempty"`
}

// ControllerMachineStart 控制器机器程序入口。
func ControllerMachineStart(ctx context.Context, systems string) <-chan struct{} {
	arr := strings.Split(systems, ",")
	arr = util.CleanStrings(arr)
	for _, v := range arr {
		if _, ok := model.AllWHQLJobTestSystems[v]; !ok {
			log.Fatal(ctx, cc.ExitCodeHLKManagerInvalidSystem, "system version is invalid", v)
		}
	}

	tick := time.Tick(3 * time.Second)
	closeChan := make(chan struct{})
	log.Info(ctx, "start controller machine", systems)
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
				startJobTest(innerCtx, arr)
				dispatchJobTest(innerCtx, arr)
			case <-ctx.Done():
				return
			}
		}
	}()
	return closeChan
}

// 将测试任务调度起测试。
func startJobTest(ctx context.Context, systems []string) {
	defer func() {
		if p := recover(); p != nil {
			log.Error(ctx, "controller machine panic", p, util.GetStackCallers())
		}
	}()

	// 获取可以开启测试的任务。
	job, ok := getInitialledJob(ctx, systems)
	if !ok {
		return
	}

	// 将测试机移动到默认池，并重置状态为 Ready。
	machinePoolName, ok := resetTestMachineState(ctx, job.ID, job.TestMachineName)
	if !ok {
		return
	}

	// 找出 HLK Studio 中的测试目标。
	target, ok := findTestTarget(ctx, job.ID, job.TestMachineName, machinePoolName, job.TestTarget, job.ServiceName)
	if !ok {
		return
	}

	// 创建 HLK Studio 测试项目。
	projectName, ok := createTestProject(ctx, job.ID)
	if !ok {
		return
	}

	// 将测试目标添加到 HLK Studio 测试项目中。
	if ok = loadTargetToProject(ctx, job.ID, target, machinePoolName, job.TestMachineName, projectName); !ok {
		return
	}

	// 获取所有测试项。
	tests, ok := getAllTargetTests(ctx, job.ID, target, job.TestMachineName, machinePoolName, projectName)
	if !ok {
		return
	}

	// 启动测试。
	testIDs, ok := queueTargetTest(ctx, job.ID, tests, target, job.TestMachineName, machinePoolName, projectName)
	if !ok {
		return
	}

	// 记录测试信息。
	ok = recordTestInfoToDisk(ctx, job.ID, tests, target, job.TestMachineName, machinePoolName, projectName, testIDs)
	if !ok {
		return
	}

	// 更新任务状态。
	if ok = updateJobToListenTestResult(ctx, job.ID); !ok {
		return
	}
}

// 监听任务测试结果，开启下一个测试。
func dispatchJobTest(ctx context.Context, systems []string) {
	defer func() {
		if p := recover(); p != nil {
			log.Error(ctx, "controller machine panic", p, util.GetStackCallers())
		}
	}()

	// 获取测试中的任务。
	whqlJobs, ok := getTestingJobs(ctx, systems)
	if !ok {
		return
	}

	// 获取所有 HLK 项目。
	var projects []*hlkProjectInfo
	projects, ok = getAllHLKProjects(ctx)
	if !ok {
		return
	}

	for _, whqlJob := range whqlJobs {

		func() {
			defer func() {
				if p := recover(); p != nil {
					log.Error(ctx, "controller machine panic", p, util.GetStackCallers())
				}
			}()

			// 获取任务测试信息。
			var testInfo *testJobInfo2
			testInfo, ok = getJobTestingInfo(ctx, whqlJob.ID)
			if !ok {
				return
			}

			// 检查测试项目结果。
			var finished bool
			finished, ok = checkHLKProjectState(ctx, whqlJob.ID, testInfo, projects)
			if !ok {
				return
			}
			if !finished {
				return
			}

			// 检查测试项结果。
			var finishedTestIDs []string
			var failed bool
			finishedTestIDs, finished, failed, ok = checkTestResult(ctx, whqlJob.ID, testInfo)
			if !ok {
				return
			}

			if !failed && !finished {
				// 分配下一个测试任务。
				if ok = scheduleTestJob(ctx, whqlJob.ID, testInfo, finishedTestIDs); !ok {
					return
				}
				return
			}

			// 打包测试结果，并上传文件。
			var logFileID string
			logFileID, ok = zipJobTestResult(ctx, whqlJob.ID, whqlJob.AppID, testInfo)
			if !ok {
				return
			}

			// 获取测试日志。
			var jobLog string
			jobLog, ok = getJobTestLog(ctx, whqlJob.ID, testInfo, finishedTestIDs)
			if !ok {
				return
			}

			// 更新任务信息，结束测试。
			if failed {
				failTestJob(ctx, whqlJob.ID, logFileID, jobLog)
				return
			}

			// 生成 hlkx 包文件。
			var hlkxFileID string
			hlkxFileID, ok = createHLKXFile(
				ctx, whqlJob.ID, whqlJob.AppID, whqlJob.FileID, whqlJob.ServiceName, testInfo)
			if !ok {
				return
			}

			// 更新任务信息。
			finishTestJob(ctx, whqlJob.ID, hlkxFileID, logFileID, jobLog)
		}()

	}
}

// 获取可以开启测试的任务。
func getInitialledJob(ctx context.Context, systems []string) (job *model.WhqlJob, ok bool) {
	log.Debug(ctx, "get initialized job")

	// 请求主服务获取。
	whqlJob, err := httpBackendGetWHQLJobToStartTest(ctx, systems)
	if err != nil {
		log.Error(ctx, "getting initialized job failed", err)
		return nil, false
	}
	if whqlJob == nil {
		return nil, false
	}
	log.Info(ctx, "got initialized job", whqlJob.ID, whqlJob.ServiceName, whqlJob.TestMachineName)

	return whqlJob, true
}

// 将测试机移动到默认池，并重置状态为 Ready。
func resetTestMachineState(ctx context.Context, jobID int, testMachineName string) (machinePoolName string, ok bool) {
	log.Info(ctx, "reset test machine state")

	// 获取所有测试池。
	hlkPools, err := listHLKPools()
	if err != nil {
		log.Error(ctx, "getting hlk pools failed", err)
		failJob(ctx, jobID, "重置测试机状态发生错误，%v", err)
		return "", false
	}

	// 查找测试机。
	var needCreatePool = true
	var foundMachine bool
	var needMove bool
	var machineState string
	var machineInPool string
	machinePoolName = getMachinePoolName(testMachineName)
F:
	for _, v := range hlkPools {
		if v.Name == consts.HLKStudioDefaultPool {
			for _, machine := range v.Machines {
				if machine.Name == testMachineName {
					needMove = true
					machineState = machine.State
					foundMachine = true
					machineInPool = consts.HLKStudioDefaultPool
					break
				}
			}
		} else if v.Name == machinePoolName {
			needCreatePool = false
			for _, machine := range v.Machines {
				if machine.Name == testMachineName {
					needMove = false
					machineState = machine.State
					foundMachine = true
					machineInPool = machinePoolName
					break F
				}
			}
		} else {
			for _, machine := range v.Machines {
				if machine.Name == testMachineName {
					needMove = true
					machineState = machine.State
					foundMachine = true
					machineInPool = v.Name
					break
				}
			}
		}
	}

	// 测试机器没找到，就结束任务。
	if !foundMachine {
		log.Error(ctx, "test machine not found", testMachineName)
		failJob(ctx, jobID, "找不到测试机器")
		return "", false
	}

	// 创建测试池。
	if needCreatePool {
		if err = createHLKPool(machinePoolName); err != nil {
			log.Error(ctx, "creating pool failed", err)
			failJob(ctx, jobID, "重置测试机状态发生错误，%v", err)
			return "", false
		}
	}

	// 重置机器状态。
	if !needMove && machineState != consts.MachineReadyState {
		if err = setHLKMachineState(testMachineName, machineInPool, consts.MachineReadyState); err != nil {
			log.Error(ctx, "setting machine state failed", err)
			failJob(ctx, jobID, "重置测试机状态发生错误，%v", err)
			return "", false
		}
	}

	// 移动机器。
	if needMove {
		log.Info(ctx, "move test machine to other pool")
		if err = moveHLKMachine(testMachineName, machineInPool, machinePoolName); err != nil {
			log.Error(ctx, "moving machine failed", err)
			failJob(ctx, jobID, "重置测试机状态发生错误，%v", err)
			return "", false
		}
	}

	// 重置机器状态。
	if err = setHLKMachineState(testMachineName, machinePoolName, consts.MachineReadyState); err != nil {
		log.Error(ctx, "setting machine state failed", err)
		failJob(ctx, jobID, "重置测试机状态发生错误，%v", err)
		return "", false
	}

	return machinePoolName, true
}

// 找出 HLK Studio 中的测试目标。
func findTestTarget(ctx context.Context, jobID int, testMachineName, machinePoolName, testTarget, serviceName string) (
	target string, ok bool) {

	log.Info(ctx, "find test target")

	// 运行脚本获取所有目标。
	targets, err := listHLKMachineTargets(testMachineName, machinePoolName)
	if err != nil {
		log.Error(ctx, "getting targets failed", err)
		failJob(ctx, jobID, "查找测试目标发生错误，%v", err)
		return "", false
	}

	// 遍历比较。
	for _, v := range targets {
		if len(testTarget) > 0 &&
			(strings.EqualFold(v.Name, testTarget) || strings.EqualFold(v.Name, testTarget+".sys")) {
			return v.Key, true
		}
		if strings.EqualFold(v.Name, serviceName) || strings.EqualFold(v.Name, serviceName+".sys") {
			return v.Key, true
		}
	}

	// 没找到目标。
	failJob(ctx, jobID, "未找到测试目标，%s，%s", testTarget, serviceName)
	return "", false
}

// 创建 HLK 测试项目。
func createTestProject(ctx context.Context, jobID int) (projectName string, ok bool) {
	log.Info(ctx, "create test project")

	// 运行脚本。
	projectName = getProjectName(jobID)
	err := createHLKProject(projectName, true)
	if err != nil {
		log.Error(ctx, "creating hlk project failed", err)
		failJob(ctx, jobID, "创建 HLK 项目失败，%v", err)
		return "", false
	}

	return projectName, true
}

// 将测试目标添加到 HLK Studio 测试项目中。
func loadTargetToProject(ctx context.Context, jobID int, target, machinePoolName, testMachineName, projectName string) (
	ok bool) {

	log.Info(ctx, "load target to project")

	// 运行脚本。
	err := createHLKProjectTarget(target, projectName, testMachineName, machinePoolName)
	if err != nil {
		log.Error(ctx, "loading target to project failed", err)
		failJob(ctx, jobID, "加载测试目标失败，%v", err)
		return false
	}

	return true
}

// 获取所有测试项。
func getAllTargetTests(ctx context.Context, jobID int, target, testMachineName, machinePoolName, projectName string) (
	tests []*targetTestInfo, ok bool) {

	log.Info(ctx, "get all tests")

	// 运行脚本。
	hlkTests, err := listHLKTests(target, projectName, testMachineName, machinePoolName)
	if err != nil {
		log.Error(ctx, "getting all tests failed", err)
		failJob(ctx, jobID, "获取测试项发生错误，%v", err)
		return nil, false
	}

	// 转换数据。
	tests = make([]*targetTestInfo, len(hlkTests))
	for i, v := range hlkTests {
		tests[i] = &targetTestInfo{
			Name:          v.Name,
			ID:            v.ID,
			TestType:      v.TestType,
			EstimatedTime: parseEstimatedTime(ctx, v.EstimatedRuntime),
		}
	}

	return tests, true
}

// 开启测试。
func queueTargetTest(ctx context.Context, jobID int, tests []*targetTestInfo, target, testMachineName, machinePoolName,
	projectName string) (testIDs []string, ok bool) {

	log.Info(ctx, "queue target test")

	// 排序。人工交互类型和耗时久的排后面。
	sort.Slice(tests, func(i, j int) bool {
		b1 := tests[i].TestType == consts.ManualTestType
		b2 := tests[j].TestType == consts.ManualTestType
		if b1 && b2 {
			if tests[i].EstimatedTime <= 0 {
				return true
			}
			if tests[j].EstimatedTime <= 0 {
				return false
			}
			return tests[i].EstimatedTime < tests[j].EstimatedTime
		}
		if b1 {
			return false
		}
		if b2 {
			return true
		}
		if tests[i].EstimatedTime <= 0 {
			return true
		}
		if tests[j].EstimatedTime <= 0 {
			return false
		}
		return tests[i].EstimatedTime < tests[j].EstimatedTime
	})

	// 检查一边测试项。
	var firstTest *targetTestInfo
	testIDs = make([]string, 0, len(tests))
	for _, v := range tests {
		if v.TestType == consts.ManualTestType {
			switch v.Name {
			case consts.TestNameWFP:
			case consts.TestNameELAMLogo:
			case consts.TestNameTT:
			case consts.TestNameHSP:
			default:
				log.Error(ctx, "unhandled test type", v.Name)
				continue
			}
		}
		v.Selected = true
		if firstTest == nil {
			firstTest = v
		}
		testIDs = append(testIDs, v.ID)
	}
	if firstTest == nil {
		log.Error(ctx, "no test found")
		failJob(ctx, jobID, "没有找到可以开启的测试项")
		return nil, false
	}

	// 将第一个测试项加入测试序列。
	err := queueHLKTest(firstTest.ID, target, projectName, testMachineName, machinePoolName)
	if err != nil {
		log.Error(ctx, "queuing target test failed", err)
		failJob(ctx, jobID, "开启测试【%s】发生错误，%v", firstTest.Name, err)
		return nil, false
	}
	log.Info(ctx, "start test job successfully", firstTest.Name)

	return testIDs, true
}

// 记录测试信息。
func recordTestInfoToDisk(ctx context.Context, jobID int, tests []*targetTestInfo, target, testMachineName,
	machinePoolName, projectName string, testIDs []string) (ok bool) {

	log.Info(ctx, "record test info to disk")

	// 序列化。
	infoBytes, err := json.Marshal(&testJobInfo2{
		Target:          target,
		MachinePoolName: machinePoolName,
		TestMachineName: testMachineName,
		ProjectName:     projectName,
		Tests:           tests,
		TestIDs:         testIDs,
	})
	if err != nil {
		log.Error(ctx, "marshaling test info failed", err)
		failJob(ctx, jobID, "记录测试信息时发生错误，%v", err)
		return false
	}

	// 写入文件。
	filePath := getTestInfoFilePath(jobID)
	if err = os.MkdirAll(filepath.Dir(filePath), cc.DirectoryMode); err != nil {
		log.Error(ctx, "mkdir failed", err)
		failJob(ctx, jobID, "记录测试信息时发生错误，%v", err)
		return false
	}
	if err = os.WriteFile(filePath, infoBytes, cc.FileMode); err != nil {
		log.Error(ctx, "write file failed", err)
		failJob(ctx, jobID, "记录测试信息时发生错误，%v", err)
		return false
	}

	return true
}

// 更新任务。
func updateJobToListenTestResult(ctx context.Context, jobID int) (ok bool) {
	log.Info(ctx, "update job to listen test result")

	// 请求主服务。
	err := httpBackendUpdateWHQLJob(ctx, &bp.WindowsInternalUpdateWHQLJobReq{
		JobID:     jobID,
		AppendLog: formatJobLog(log.LevelInfo, "已启动 HLK 测试"),
		Status:    model.WHQLJobStatusHLKTesting,
	})
	if err != nil {
		log.Error(ctx, "updating job failed", err)
		return false
	}

	return true
}

// 解析测试项任务耗时。
func parseEstimatedTime(ctx context.Context, s string) time.Duration {
	d := time.Duration(0)
	arr := strings.Split(s, ":")
	if len(arr) == 3 {
		hour, _ := strconv.Atoi(arr[0])
		minute, _ := strconv.Atoi(arr[1])
		second, _ := strconv.Atoi(arr[2])
		d = time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute + time.Duration(second)*time.Second
	} else {
		log.Error(ctx, "unknown estimated time", s)
	}
	return d
}

// 获取测试中的任务。
func getTestingJobs(ctx context.Context, systems []string) ([]*model.WhqlJob, bool) {
	log.Debug(ctx, "get whql testing jobs")

	// 请求主服务。
	whqlJobs, err := httpBackendGetTestingWHQLJobs(ctx, systems)
	if err != nil {
		log.Error(ctx, "get testing jobs failed", err)
		return nil, false
	}

	if len(whqlJobs) <= 0 {
		return nil, false
	}

	return whqlJobs, true
}

// 获取所有 HLK 项目。
func getAllHLKProjects(ctx context.Context) ([]*hlkProjectInfo, bool) {
	log.Debug(ctx, "get all hlk projects")

	// 运行 PowerShell 脚本。
	hlkProjects, err := listHLKProjects()
	if err != nil {
		log.Error(ctx, "list hlk projects failed", err)
		return nil, false
	}

	return hlkProjects, true
}

// 获取任务测试信息。
func getJobTestingInfo(ctx context.Context, jobID int) (*testJobInfo2, bool) {
	log.Debug(ctx, "get job testing info")

	// 读取文件。
	filePath := getTestInfoFilePath(jobID)
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		log.Error(ctx, "read file failed", err)
		failJob(ctx, jobID, "获取任务测试信息发生错误：%v", err)
		return nil, false
	}

	// 反序列化数据。
	var info testJobInfo2
	err = json.Unmarshal(fileData, &info)
	if err != nil {
		log.Error(ctx, "unmarshal test info failed", err)
		failJob(ctx, jobID, "获取任务测试信息发生错误：%v", err)
		return nil, false
	}

	return &info, true
}

// 检查测试项目结果。
func checkHLKProjectState(ctx context.Context, jobID int, testInfo *testJobInfo2, projects []*hlkProjectInfo) (
	finished, ok bool) {

	log.Debug(ctx, "check hlk project state")

	// 找出测试项目。
	var project *hlkProjectInfo
	for _, v := range projects {
		if strings.EqualFold(v.Name, testInfo.ProjectName) {
			project = v
			break
		}
	}
	if project == nil {
		log.Error(ctx, "project not found", testInfo.ProjectName)
		failJob(ctx, jobID, "HLK Studio 项目未找到")
		return false, false
	}

	// 检查项目状态。
	return project.Status == "NotRunning" && project.ModifiedTime >= project.CreateTime, true
}

// 检查测试结果。
func checkTestResult(ctx context.Context, jobID int, testInfo *testJobInfo2) (
	finishedTestIDs []string, finished, failed, ok bool) {

	log.Debug(ctx, "check test result")

	// 获取测试项测试结果。
	finished = true
	finishedTestIDs = make([]string, 0, len(testInfo.TestIDs))
	for _, v := range testInfo.TestIDs {
		info, err := getHLKTestInfo(
			v, testInfo.Target, testInfo.ProjectName, testInfo.TestMachineName, testInfo.MachinePoolName)
		if err != nil {
			log.Error(ctx, "get test info failed", err)
			failJob(ctx, jobID, "获取测试信息发生错误：%v", err)
			return nil, false, false, false
		}
		switch info.Status {
		case "InQueue", "Running", "NotRun":
			finished = false
		case "Passed":
			finishedTestIDs = append(finishedTestIDs, v)
		case "Failed":
			failed = true
			finishedTestIDs = append(finishedTestIDs, v)
		default:
			log.Error(ctx, "unknown test status", info.Status)
			failJob(ctx, jobID, "未知的测试项状态：%v", info.Status)
			return nil, false, false, false
		}
	}

	ok = true
	return
}

// 开启下一个测试任务。
func scheduleTestJob(ctx context.Context, jobID int, testInfo *testJobInfo2, finishedTestIDs []string) (ok bool) {
	log.Info(ctx, "schedule test job", jobID)

	// 获取要开启的测试任务 ID。
	var nextTestID, testName string
	for _, v := range testInfo.TestIDs {
		if slices.Contains(finishedTestIDs, v) {
			continue
		}
		nextTestID = v
		break
	}
	if len(nextTestID) <= 0 {
		log.Error(ctx, "test id not found")
		failJob(ctx, jobID, "分配测试任务发生错误，未找到可开启测试的任务")
		return false
	}
	for _, v := range testInfo.Tests {
		if v.ID == nextTestID {
			testName = v.Name
			break
		}
	}

	// 开启下一个测试任务项。
	err := queueHLKTest(
		nextTestID, testInfo.Target, testInfo.ProjectName, testInfo.TestMachineName, testInfo.MachinePoolName)
	if err != nil {
		log.Error(ctx, "queuing target test failed", err)
		failJob(ctx, jobID, "开启测试【%s】发生错误，%v", testName, err)
		return false
	}
	log.Info(ctx, "start test job successfully", testName)

	return true
}

// 打包测试结果，并上传文件。
func zipJobTestResult(ctx context.Context, jobID, appID int, testInfo *testJobInfo2) (logFileID string, ok bool) {
	log.Info(ctx, "zip job test result")

	// 运行脚本。
	result, err := zipHLKTestResult(
		testInfo.Target, testInfo.ProjectName, testInfo.TestMachineName, testInfo.MachinePoolName)
	if err != nil {
		log.Error(ctx, "zip test result failed", err)
		failJob(ctx, jobID, "打包测试结果发生错：%v", err)
		return "", false
	}
	defer util.RemoveFile(ctx, result.LogsZipPath)

	// 上传文件。
	fileStream, err := os.Open(result.LogsZipPath)
	if err != nil {
		log.Error(ctx, "open file failed", err)
		failJob(ctx, jobID, "打包测试结果发生错：%v", err)
		return "", false
	}
	fileInfo, err := os.Stat(result.LogsZipPath)
	if err != nil {
		log.Error(ctx, "stat file failed", err)
		failJob(ctx, jobID, "打包测试结果发生错：%v", err)
		return "", false
	}
	logFileID, err = httpBackendUploadFile(
		ctx, model.FileTypeHLKLog, filepath.Base(result.LogsZipPath), appID, fileStream, fileInfo.Size())
	if err != nil {
		log.Error(ctx, "upload file failed", err)
		failJob(ctx, jobID, "打包测试结果发生错：%v", err)
		return "", false
	}

	ok = true
	return
}

// 获取测试日志。
func getJobTestLog(ctx context.Context, jobID int, testInfo *testJobInfo2, finishedTestIDs []string) (string, bool) {
	log.Info(ctx, "get job hlk test log")

	// 调用脚本获取。
	testLogInfos := make([]*hlkTestResultInfo, 0, len(finishedTestIDs))
	for _, v := range finishedTestIDs {
		testLog, err := getHLKTestLog(
			v, testInfo.Target, testInfo.ProjectName, testInfo.TestMachineName, testInfo.MachinePoolName)
		if err != nil {
			log.Error(ctx, "get test log failed", err)
			failJob(ctx, jobID, "获取任务测试日志发生错误：%v", err)
			return "", false
		}
		testLogInfos = append(testLogInfos, testLog...)
	}

	// 处理日志。
	sb := &strings.Builder{}
	for i, v := range testLogInfos {
		sb.WriteString(v.Name)
		sb.WriteString(": \n")
		_, _ = fmt.Fprintf(sb, "  %d. %s\n", i+1, v.Name)
		for _, ts := range v.Tasks {
			if len(ts.TaskErrorMessage) > 0 {
				_, _ = fmt.Fprintf(sb, "    %s: %s\n", ts.Name, ts.TaskErrorMessage)
			} else {
				_, _ = fmt.Fprintf(sb, "    %s: %s\n", ts.Name, ts.Status)
			}
			for _, cts := range ts.ChildTasks {
				if len(ts.TaskErrorMessage) > 0 {
					_, _ = fmt.Fprintf(sb, "      %s: %s\n", cts.Name, cts.TaskErrorMessage)
				} else {
					_, _ = fmt.Fprintf(sb, "      %s: %s\n", cts.Name, cts.Status)
				}
			}
		}
	}

	return sb.String(), true
}

// 更新任务信息，结束测试。
func failTestJob(ctx context.Context, jobID int, logFileID, jobLog string) {
	log.Info(ctx, "fail test job")

	// 请求主服务。
	err := httpBackendUpdateWHQLJob(ctx, &bp.WindowsInternalUpdateWHQLJobReq{
		JobID:        jobID,
		AppendLog:    formatJobLog(log.LevelError, "HLK 测试失败\n%s", jobLog),
		Status:       model.WHQLJobStatusFailure,
		HLKLogFileID: logFileID,
	})
	if err != nil {
		log.Error(ctx, "failed to update test job to failed", err)
		return
	}
}

// 生成 hlkx 包文件。
func createHLKXFile(ctx context.Context, jobID, appID int, sourceFileID, serviceName string, testInfo *testJobInfo2) (
	hlkxFileID string, ok bool) {

	log.Info(ctx, "create hlkx file")

	// 准备文件夹。
	filesDirectory, err := util.GenerateTemporaryDirectory(filepath.Join(cc.ServiceNameHLK, strconv.Itoa(jobID)))
	if err != nil {
		log.Error(ctx, "failed to create directory", err)
		failJob(ctx, jobID, "生成 HLKX 文件发生错误：%v", err)
		return "", false
	}
	util.RemoveDirectory(ctx, filesDirectory)
	filesDirectory, err = util.GenerateTemporaryDirectory(filepath.Join(cc.ServiceNameHLK, strconv.Itoa(jobID)))
	if err != nil {
		log.Error(ctx, "failed to create directory", err)
		failJob(ctx, jobID, "生成 HLKX 文件发生错误：%v", err)
		return "", false
	}
	defer util.RemoveDirectory(ctx, filesDirectory)

	// 下载驱动程序。
	filePath, _, err := httpBackendDownloadFile(ctx, sourceFileID, filesDirectory)
	if err != nil {
		log.Error(ctx, "failed to download source file", err)
		failJob(ctx, jobID, "生成 HLKX 文件发生错误：%v", err)
		return "", false
	}
	if strings.ToLower(filepath.Ext(filePath)) == cc.ExtensionZIP {
		log.Info(ctx, "unzip file")
		if err = util.Unzip(ctx, filePath, filesDirectory); err != nil {
			log.Error(ctx, "failed to unzip file", err)
			failJob(ctx, jobID, "生成 HLKX 文件发生错误：%v", err)
			return "", false
		}
	}

	// 生成 inf 文件。
	entries, err := os.ReadDir(filesDirectory)
	if err != nil {
		log.Error(ctx, "failed to read directory", err)
		failJob(ctx, jobID, "生成 HLKX 文件发生错误：%v", err)
		return "", false
	}
	fileName := filepath.Base(filePath)
	infFileName := fileName + cc.ExtensionINF
	needCreateInfFile := true
	for _, v := range entries {
		if strings.EqualFold(v.Name(), infFileName) {
			log.Warn(ctx, "inf file already exists", infFileName)
			needCreateInfFile = false
			break
		}
	}
	if needCreateInfFile {
		content := strings.ReplaceAll(consts.INFTemplate, "!!REPLACE_ME!!", serviceName)
		err = os.WriteFile(filepath.Join(filesDirectory, infFileName), []byte(content), cc.FileMode)
		if err != nil {
			log.Error(ctx, "failed to write inf file", err)
			failJob(ctx, jobID, "生成 HLKX 文件发生错误：%v", err)
			return "", false
		}
	}

	// 运行脚本创建。
	destinationFilePath := filepath.Join(filesDirectory, fileName+cc.ExtensionHLKX)
	_, err = createHLKPackage(testInfo.ProjectName, destinationFilePath, filesDirectory)
	if err != nil {
		log.Error(ctx, "failed to create hlk package", err)
		failJob(ctx, jobID, "生成 HLKX 文件发生错误：%v", err)
		return "", false
	}

	// 上传文件到主服务。
	fileStream, err := os.Open(destinationFilePath)
	if err != nil {
		log.Error(ctx, "failed open hlkx file", err, destinationFilePath)
		failJob(ctx, jobID, "生成 HLKX 文件发生错误：%v", err)
		return "", false
	}
	fileInfo, err := os.Stat(destinationFilePath)
	if err != nil {
		log.Error(ctx, "failed to stat hlkx file", err, destinationFilePath)
		failJob(ctx, jobID, "生成 HLKX 文件发生错误：%v", err)
		return "", false
	}
	hlkxFileID, err = httpBackendUploadFile(
		ctx, model.FileTypeWindowsSigning, filepath.Base(destinationFilePath), appID, fileStream, fileInfo.Size())
	if err != nil {
		log.Error(ctx, "failed to upload hlkx file", err, destinationFilePath)
		failJob(ctx, jobID, "生成 HLKX 文件发生错误：%v", err)
		return "", false
	}

	ok = true
	return
}

// 测试通过，更新任务信息。
func finishTestJob(ctx context.Context, jobID int, hlkxFileID, logFileID, jobLog string) {
	log.Info(ctx, "finish test job", jobID)

	err := httpBackendUpdateWHQLJob(ctx, &bp.WindowsInternalUpdateWHQLJobReq{
		JobID:            jobID,
		AppendLog:        formatJobLog(log.LevelInfo, "HLK 测试通过\n%s", jobLog),
		Status:           model.WHQLJobStatusFinishTest,
		HLKXFileID:       hlkxFileID,
		HLKLogFileID:     logFileID,
		FinishedTestTime: bp.Time(time.Now()),
	})
	if err != nil {
		log.Error(ctx, "failed to http backend to finish test job", err)
		return
	}
}
