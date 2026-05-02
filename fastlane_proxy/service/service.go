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
	"os"
	"os/exec"
	"strings"
)

func executeFastlaneCommand(args ...string) (string, error) {
	command := exec.Command("fastlane", args...)
	command.Env = append(command.Env, os.Environ()...)
	command.Env = append(command.Env, "FASTLANE_HIDE_TIMESTAMP=1", "FASTLANE_DISABLE_COLORS=1",
		"FASTLANE_SKIP_UPDATE_CHECK=1")
	output, err := command.CombinedOutput()
	return string(output), err
}

func extractOutput(action string, str string) string {
	index := strings.Index(str, action)
	if index < 0 {
		return ""
	}
	str = str[index:]
	index = strings.Index(str, "\n{")
	if index < 0 {
		return ""
	}
	str = str[index+1:]
	index = strings.Index(str, "}\n")
	if index < 0 {
		return ""
	}
	return str[:index+1]
}
