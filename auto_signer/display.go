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

package main

import (
	"fmt"
	"time"
)

// StepRunner 管理单个步骤的显示生命周期。
type StepRunner struct {
	number    int
	total     int
	name      string
	startTime time.Time
}

// NewStepRunner 创建步骤运行器。
func NewStepRunner(number, total int, name string) *StepRunner {
	return &StepRunner{number: number, total: total, name: name}
}

// PrintHeader 打印程序头部。
func PrintHeader(taskID string) {
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Printf("  CSMS 自动化签名程序  %s %s\n", Version(), taskID)
	fmt.Println("══════════════════════════════════════════════════════════════")
}

// PrintSuccessFooter 打印成功结果汇总。
func PrintSuccessFooter(duration time.Duration) {
	fmt.Println()
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Printf("  ✓ 签名成功    总耗时: %s\n", FormatDuration(duration))
	fmt.Println("══════════════════════════════════════════════════════════════")
}

// PrintErrorFooter 打印失败结果汇总。
func PrintErrorFooter(duration time.Duration) {
	fmt.Println()
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Printf("  ✗ 签名失败    总耗时: %s\n", FormatDuration(duration))
	fmt.Println("══════════════════════════════════════════════════════════════")
}

// Start 打印步骤开始行。
func (s *StepRunner) Start() {
	s.startTime = time.Now()
	fmt.Printf("◉ 步骤 %d/%d  %-30s", s.number, s.total, s.name)
}

// Done 打印步骤完成行及可选的子信息行。
func (s *StepRunner) Done(subLines ...string) {
	duration := time.Since(s.startTime)
	fmt.Printf("\r● 步骤 %d/%d  %-30s [完成]  耗时 %-100s\n", s.number, s.total, s.name, FormatDuration(duration))
	for _, line := range subLines {
		fmt.Printf("    └─ %s\n", line)
	}
}

// Fail 打印步骤失败行。
func (s *StepRunner) Fail(errMsg string) {
	duration := time.Since(s.startTime)
	fmt.Printf("\r✗ 步骤 %d/%d  %-30s [失败]  耗时 %-100s\n", s.number, s.total, s.name, FormatDuration(duration))
	if errMsg != "" {
		fmt.Printf("    错误: %s\n", errMsg)
	}
}

// UpdateRunning 使用附加信息刷新运行中状态行。
func (s *StepRunner) UpdateRunning(info string) {
	fmt.Printf("\r◉ 步骤 %d/%d  %-30s %s", s.number, s.total, s.name, info)
}
