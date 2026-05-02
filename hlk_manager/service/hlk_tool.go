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
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	"gitee.com/ivfzhou/csms/hlk_manager/consts"
)

type hlkToolRsp[E any] struct {
	Result  bool   `json:"result"`
	Message string `json:"message"`
	Content E      `json:"content"`
}

type hlkPoolInfo struct {
	Name     string `json:"name"`
	Machines []*struct {
		Name          string `json:"name"`
		State         string `json:"state"`
		LastHeartBeat string `json:"lastHeartBeat"`
	} `json:"machines"`
}

type hlkMachineTargetInfo struct {
	Name string `json:"name"`
	Key  string `json:"key"`
	Type int    `json:"type"`
}

type hlkTestInfo struct {
	Name                         string   `json:"name"`
	ID                           string   `json:"id"`
	TestType                     string   `json:"testtype"`
	EstimatedRuntime             string   `json:"estimatedruntime"`
	RequiresSpecialConfiguration string   `json:"requiresspecialconfiguration"`
	RequiresSupplementalContent  string   `json:"requiressupplementalcontent"`
	ScheduleOptions              []string `json:"scheduleoptions"`
	Status                       string   `json:"status"`
	ExecutionState               string   `json:"executionstate"`
}

type hlkProjectInfo struct {
	Name             string `json:"name"`
	CreateTime       string `json:"creationtime"`
	ModifiedTime     string `json:"modifiedtime"`
	Status           string `json:"status"`
	ProductInstances []*struct {
		Name         string `json:"name"`
		OSPlatform   string `json:"osplatform"`
		TargetedPool string `json:"targetedpool"`
		Targets      []*struct {
			Name    string `json:"name"`
			Key     string `json:"key"`
			Machine string `json:"machine"`
		} `json:"targets"`
	} `json:"productinstances"`
}

type zipTestResultInfo struct {
	LogsZipPath string `json:"logszippath"`
}

type hlkTestResultInfo struct {
	Name              string         `json:"name"`
	CompletionTime    string         `json:"completiontime"`
	ScheduleTime      string         `json:"scheduletime"`
	StartTime         string         `json:"starttime"`
	Status            string         `json:"status"`
	AreFiltersApplied string         `json:"arefiltersapplied"`
	Target            string         `json:"target"`
	Tasks             []*hlkTaskInfo `json:"tasks"`
}

type hlkTaskInfo struct {
	Name             string         `json:"name"`
	Stage            string         `json:"stage"`
	Status           string         `json:"status"`
	TaskErrorMessage string         `json:"taskerrormessage"`
	TaskType         string         `json:"tasktype"`
	ChildTasks       []*hlkTaskInfo `json:"childtasks"`
}

type hlkxPackageInfo struct {
	Name               string `json:"name"`
	ProjectPackagePath string `json:"projectpackagepath"`
}

func listHLKPools() ([]*hlkPoolInfo, error) {
	// 执行脚本。
	command := exec.Command("PowerShell.exe", "-NoLogo", "-File", consts.HLKToolScriptFilePath, "List-Pools")
	errorOutput := &bytes.Buffer{}
	command.Stderr = errorOutput
	outputBytes, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}

	// 反序列化结果。
	var result hlkToolRsp[[]*hlkPoolInfo]
	if err = json.Unmarshal(outputBytes, &result); err != nil {
		return nil, fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}
	if !result.Result {
		return nil, errors.New(result.Message)
	}

	return result.Content, nil
}

func createHLKPool(machinePoolName string) error {
	// 运行脚本。
	command := exec.Command("PowerShell.exe", "-NoLogo", "-File", consts.HLKToolScriptFilePath,
		"Create-Pool,"+machinePoolName)
	errorOutput := &bytes.Buffer{}
	command.Stderr = errorOutput
	outputBytes, err := command.Output()
	if err != nil {
		return fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}

	// 反序列化结果。
	var result hlkToolRsp[any]
	if err = json.Unmarshal(outputBytes, &result); err != nil {
		return fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}
	if !result.Result {
		return errors.New(result.Message)
	}

	return nil
}

func setHLKMachineState(testMachineName, machinePoolName, machineState string) error {
	// 运行脚本。
	command := exec.Command("PowerShell.exe", "-NoLogo", "-File", consts.HLKToolScriptFilePath,
		fmt.Sprintf("Set-Machine-State,%s,%s,%s", testMachineName, machinePoolName, machineState))
	errorOutput := &bytes.Buffer{}
	command.Stderr = errorOutput
	outputBytes, err := command.Output()
	if err != nil {
		return fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}

	// 反序列化结果。
	var result hlkToolRsp[any]
	if err = json.Unmarshal(outputBytes, &result); err != nil {
		return fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}
	if !result.Result {
		return errors.New(result.Message)
	}

	return nil
}

func moveHLKMachine(testMachineName, sourcePool, destinationPool string) error {
	// 运行脚本。
	command := exec.Command("PowerShell.exe", "-NoLogo", "-File", consts.HLKToolScriptFilePath,
		fmt.Sprintf("Move-Machine,%s,%s,%s", testMachineName, sourcePool, destinationPool))
	errorOutput := &bytes.Buffer{}
	command.Stderr = errorOutput
	outputBytes, err := command.Output()
	if err != nil {
		return fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}

	// 反序列化结果。
	var result hlkToolRsp[any]
	if err = json.Unmarshal(outputBytes, &result); err != nil {
		return fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}
	if !result.Result {
		return errors.New(result.Message)
	}

	return nil
}

func listHLKMachineTargets(testMachineName, machinePoolName string) ([]*hlkMachineTargetInfo, error) {
	// 运行脚本。
	command := exec.Command("PowerShell.exe", "-NoLogo", "-File", consts.HLKToolScriptFilePath,
		fmt.Sprintf("List-Machine-Targets,%s,%s", testMachineName, machinePoolName))
	errorOutput := &bytes.Buffer{}
	command.Stderr = errorOutput
	outputBytes, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}

	// 反序列化结果。
	var result hlkToolRsp[[]*hlkMachineTargetInfo]
	if err = json.Unmarshal(outputBytes, &result); err != nil {
		return nil, fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}
	if !result.Result {
		return nil, errors.New(result.Message)
	}

	return result.Content, nil
}

func createHLKProject(projectName string, isWindowsDriverProject bool) error {
	// 运行脚本。
	command := exec.Command("PowerShell.exe", "-NoLogo", "-File", consts.HLKToolScriptFilePath,
		fmt.Sprintf(`Create-Project,%s,%v`, projectName, isWindowsDriverProject))
	errorOutput := &bytes.Buffer{}
	command.Stderr = errorOutput
	outputBytes, err := command.Output()
	if err != nil {
		return fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}

	// 反序列化结果。
	var result hlkToolRsp[any]
	if err = json.Unmarshal(outputBytes, &result); err != nil {
		return fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}
	if !result.Result {
		return errors.New(result.Message)
	}

	return nil
}

func createHLKProjectTarget(target, projectName, testMachineName, machinePoolName string) error {
	// 运行脚本。
	command := exec.Command("PowerShell.exe", "-NoLogo", "-File", consts.HLKToolScriptFilePath,
		fmt.Sprintf(`Create-Project-Target,%s,%s,%s,%s`, target, projectName, testMachineName, machinePoolName))
	errorOutput := &bytes.Buffer{}
	command.Stderr = errorOutput
	outputBytes, err := command.Output()
	if err != nil {
		return fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}

	// 反序列化结果。
	var result hlkToolRsp[any]
	if err = json.Unmarshal(outputBytes, &result); err != nil {
		return fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}
	if !result.Result {
		return errors.New(result.Message)
	}

	return nil
}

func listHLKTests(target, projectName, testMachineName, machinePoolName string) ([]*hlkTestInfo, error) {
	// 运行脚本。
	command := exec.Command("PowerShell.exe", "-NoLogo", "-File", consts.HLKToolScriptFilePath,
		fmt.Sprintf(`List-Tests,%s,%s,%s,%s`, target, projectName, testMachineName, machinePoolName))
	errorOutput := &bytes.Buffer{}
	command.Stderr = errorOutput
	outputBytes, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}

	// 反序列化结果。
	var result hlkToolRsp[[]*hlkTestInfo]
	if err = json.Unmarshal(outputBytes, &result); err != nil {
		return nil, fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}
	if !result.Result {
		return nil, errors.New(result.Message)
	}

	return result.Content, nil
}

func queueHLKTest(testID, target, projectName, testMachineName, machinePoolName string) error {
	// 运行脚本。
	command := exec.Command("PowerShell.exe", "-NoLogo", "-File", consts.HLKToolScriptFilePath,
		fmt.Sprintf(`Queue-Test,%s,%s,%s,%s,%s`, testID, target, projectName, testMachineName, machinePoolName))
	errorOutput := &bytes.Buffer{}
	command.Stderr = errorOutput
	outputBytes, err := command.Output()
	if err != nil {
		return fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}

	// 反序列化结果。
	var result hlkToolRsp[any]
	if err = json.Unmarshal(outputBytes, &result); err != nil {
		return fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}
	if !result.Result {
		return errors.New(result.Message)
	}

	return nil
}

func listHLKProjects() ([]*hlkProjectInfo, error) {
	// 运行脚本。
	command := exec.Command("PowerShell.exe", "-NoLogo", "-File", consts.HLKToolScriptFilePath, "List-Projects")
	errorOutput := &bytes.Buffer{}
	command.Stderr = errorOutput
	outputBytes, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}

	// 反序列化结果。
	var result hlkToolRsp[[]*hlkProjectInfo]
	if err = json.Unmarshal(outputBytes, &result); err != nil {
		return nil, fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}
	if !result.Result {
		return nil, errors.New(result.Message)
	}

	return result.Content, nil
}

func getHLKTestInfo(testID, target, projectName, testMachineName, machinePoolName string) (*hlkTestInfo, error) {
	// 运行脚本。
	command := exec.Command("PowerShell.exe", "-NoLogo", "-File", consts.HLKToolScriptFilePath,
		fmt.Sprintf(`Get-Test-Info,%s,%s,%s,%s,%s`, testID, target, projectName, testMachineName, machinePoolName))
	errorOutput := &bytes.Buffer{}
	command.Stderr = errorOutput
	outputBytes, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}

	// 反序列化结果。
	var result hlkToolRsp[*hlkTestInfo]
	if err = json.Unmarshal(outputBytes, &result); err != nil {
		return nil, fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}
	if !result.Result {
		return nil, errors.New(result.Message)
	}

	return result.Content, nil
}

func zipHLKTestResult(target, projectName, testMachineName, machinePoolName string) (*zipTestResultInfo, error) {
	// 运行脚本。
	command := exec.Command("PowerShell.exe", "-NoLogo", "-File", consts.HLKToolScriptFilePath,
		fmt.Sprintf(`Zip-Test-Result-Logs,%s,%s,%s,%s`, target, projectName, testMachineName, machinePoolName))
	errorOutput := &bytes.Buffer{}
	command.Stderr = errorOutput
	outputBytes, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}

	// 反序列化结果。
	var result hlkToolRsp[*zipTestResultInfo]
	if err = json.Unmarshal(outputBytes, &result); err != nil {
		return nil, fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}
	if !result.Result {
		return nil, errors.New(result.Message)
	}

	return result.Content, nil
}

func getHLKTestLog(testID, target, projectName, testMachineName, machinePoolName string) ([]*hlkTestResultInfo, error) {
	// 运行脚本。
	command := exec.Command("PowerShell.exe", "-NoLogo", "-File", consts.HLKToolScriptFilePath,
		fmt.Sprintf(`List-Test-Results,%s,%s,%s,%s,%s`, testID, target, projectName, testMachineName, machinePoolName))
	errorOutput := &bytes.Buffer{}
	command.Stderr = errorOutput
	outputBytes, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}

	outputBytes = bytes.ReplaceAll(outputBytes, []byte(`"permit_all"`), []byte(`permit_all`))

	// 反序列化结果。
	var result hlkToolRsp[[]*hlkTestResultInfo]
	if err = json.Unmarshal(outputBytes, &result); err != nil {
		return nil, fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}
	if !result.Result {
		return nil, errors.New(result.Message)
	}

	return result.Content, nil
}

func createHLKPackage(projectName, destinationFilePath, driverFilePath string) (*hlkxPackageInfo, error) {
	// 运行脚本。
	command := exec.Command("PowerShell.exe", "-NoLogo", "-File", consts.HLKToolScriptFilePath,
		fmt.Sprintf(`Create-Project-Package,%s,%s,%s`, projectName, destinationFilePath, driverFilePath))
	errorOutput := &bytes.Buffer{}
	command.Stderr = errorOutput
	outputBytes, err := command.Output()

	// 反序列化结果。
	var result hlkToolRsp[*hlkxPackageInfo]
	if err = json.Unmarshal(outputBytes, &result); err != nil {
		return nil, fmt.Errorf("%w %s %s", err, outputBytes, errorOutput.Bytes())
	}
	if !result.Result {
		return nil, errors.New(result.Message)
	}

	return result.Content, nil
}
