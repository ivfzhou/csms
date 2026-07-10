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

export function initWebAPI() {
    request.variables.set('date', new Date().toUTCString())
    let language = 'zh'
    if (Math.random() >= 0.5) language = 'en'
    request.variables.set('language', language)
}

export function checkAPIResult(httpCode, serverCode) {
    client.assert(response.status === httpCode, `http status is not ${httpCode}, actual is ${response.status}`)
    if (serverCode === undefined && !response.body) return
    client.assert(response.body.code === serverCode, `server code is not ${serverCode}, actual is ${response.body.code} ${response.body.message} ${response.headers.valueOf('X-Csms-Request-Id')}`)
}

export function parseAndSaveWebAuthorization(response) {
    const list = response.headers.valuesOf('Set-Cookie')
    let name = ''
    let session = ''
    list.find(it => {
        it.split(';').forEach(it => {
            it = it.trim()
            if (!session && it.startsWith('csms_session')) {
                const pair = it.split('=')
                const value = pair[1].trim()
                if (pair.length === 2 && value.length > 0) {
                    session = value
                }
            } else if (!name && it.startsWith('csms_user')) {
                const pair = it.split('=')
                const value = pair[1].trim()
                if (pair.length === 2 && value.length > 0) {
                    name = value
                }
            }
        })
        return name && session
    })
    return [name, session]
}

export function createAuthorization(appId, apiAccountId, secret) {
    const option = {
        algorithm: 'HS256',
        header: {
            "typ": 'JWT',
            "alg": 'HS256'
        }
    }
    const now = Number((Date.now() / 1000).toFixed(0)) - 180
    const claim = {
        exp: Math.floor(Date.now() / 1000) + 3600,
        iss: appId,
        sub: apiAccountId,
        iat: now,
        nbf: now
    }
    return jwt.sign(claim, secret, option)
}
