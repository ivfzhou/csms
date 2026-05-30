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

import {httpGetJson} from "@/api/http.js"
import {useMessageStore} from "@/stores/message.js"
import {useLanguageStore} from "@/stores/language.js"
import {isSuccessHttpCode} from "@/utils/utils.js"

// 获取通知。
export async function getLastNotification() {
    const messageStore = useMessageStore()
    try {
        const rsp = await httpGetJson('/backend/web/notice/last', {language: useLanguageStore().language})
        if (!isSuccessHttpCode(rsp.status)) {
            messageStore.message.error(`获取通知失败 ${rsp.status} ${rsp}`)
            return Promise.resolve({ok: false, data: undefined})
        }
        if (!rsp.rspBody) {
            messageStore.message.error(`获取通知失败 ${rsp}`)
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
        messageStore.message.error(`获取通知异常 ${err}`)
        return Promise.resolve({ok: false, data: undefined})
    }
}
