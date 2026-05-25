/*
 * Copyright (c) 2023 ivfzhou
 * csms is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

import {httpGetJson, httpJsonPostJson} from "@/api/http.js"
import {useLanguageStore} from "@/stores/language.js"

// 获取用户信息。
export async function getUserInformation() {
    return await httpGetJson('/backend/web/user/getInformation', {language: useLanguageStore().language})
}

// 登陆。
export async function userLogin(payload) {
    return await httpJsonPostJson('/backend/web/user/login', payload, {language: useLanguageStore().language})
}
