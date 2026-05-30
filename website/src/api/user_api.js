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

import {httpFormPostJson, httpGetJson, httpJsonPostJson} from "@/api/http.js"
import {useLanguageStore} from "@/stores/language.js"
import {useMessageStore} from "@/stores/message.js"
import {isSuccessHttpCode} from "@/utils/utils.js"

// 获取用户信息。
export async function getUserInformation() {
    const messageStore = useMessageStore()
    try {
        const rsp = await httpGetJson('/backend/web/user/getInformation', {language: useLanguageStore().language})
        if (!isSuccessHttpCode(rsp.status)) {
            messageStore.message.error(`获取用户信息失败 ${rsp.status} ${rsp}`)
            return Promise.resolve({ok: false, data: undefined})
        }
        if (!rsp.rspBody) {
            messageStore.message.error(`获取用户信息失败 ${rsp}`)
            return Promise.resolve({ok: false, data: undefined})
        }
        if (rsp.rspBody.code > 0) {
            messageStore.message.warning(`${rsp.rspBody.code} ${rsp.rspBody.message}`)
            return Promise.resolve({ok: false, data: undefined})
        }
        if (rsp.rspBody.code < 0) {
            messageStore.message.success(`${rsp.rspBody.code} ${rsp.rspBody.message}`)
        }

        return Promise.resolve({ok: true, data: rsp.rspBody.data})
    } catch (err) {
        messageStore.message.error(`获取用户信息异常 ${err}`)
        return Promise.resolve({ok: false, data: undefined})
    }
}

// 登陆。
export async function userLogin(payload) {
    const messageStore = useMessageStore()
    try {
        const rsp = await httpJsonPostJson('/backend/web/user/login', payload, {language: useLanguageStore().language})
        if (!isSuccessHttpCode(rsp.status)) {
            messageStore.message.error(`登陆失败 ${rsp.status} ${rsp}`)
            return Promise.resolve({ok: false, data: undefined})
        }
        if (!rsp.rspBody) {
            messageStore.message.error(`登陆失败 ${rsp}`)
            return Promise.resolve({ok: false, data: undefined})
        }
        if (rsp.rspBody.code > 0) {
            messageStore.message.warning(`${rsp.rspBody.code} ${rsp.rspBody.message}`)
            return Promise.resolve({ok: false, data: undefined})
        }
        if (rsp.rspBody.code < 0) {
            messageStore.message.success(`${rsp.rspBody.code} ${rsp.rspBody.message}`)
        }

        return Promise.resolve({ok: true, data: rsp.rspBody.data})
    } catch (err) {
        messageStore.message.error(`登陆异常 ${err}`)
        return Promise.resolve({ok: false, data: undefined})
    }
}

// 注册。
export async function userRegister(payload) {
    const messageStore = useMessageStore()
    try {
        const rsp = await httpFormPostJson('/backend/web/user/register', payload, {language: useLanguageStore().language})
        if (!isSuccessHttpCode(rsp.status)) {
            messageStore.message.error(`注册失败 ${rsp.status} ${rsp}`)
            return Promise.resolve({ok: false, data: undefined})
        }
        if (!rsp.rspBody) {
            messageStore.message.error(`注册失败 ${rsp}`)
            return Promise.resolve({ok: false, data: undefined})
        }
        if (rsp.rspBody.code > 0) {
            messageStore.message.warning(`${rsp.rspBody.code} ${rsp.rspBody.message}`)
            return Promise.resolve({ok: false, data: undefined})
        }
        if (rsp.rspBody.code < 0) {
            messageStore.message.success(`${rsp.rspBody.code} ${rsp.rspBody.message}`)
        }

        return Promise.resolve({ok: true, data: rsp.rspBody.data})
    } catch (err) {
        messageStore.message.error(`注册异常 ${err}`)
        return Promise.resolve({ok: false, data: undefined})
    }
}
