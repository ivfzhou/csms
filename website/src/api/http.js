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

import constants from "@/utils/constants.js"

/* httpGet 发送 HTTP GET 请求
 * @param reqUrl string，资源地址
 * @param query object，URL Query
 * @param headers object，请求头数据
 * @returns Promise<object | string>
 */
export function httpGet(reqUrl, query, headers) {
    if (fetch) return fetchGet(reqUrl, query, headers)
    else return xhrGet(reqUrl, query, headers)
}

/* httpGetJson 发送 HTTP GET 请求，以 JSON 格式解析响应数据
 * @param reqUrl string，资源地址
 * @param query object，URL Query
 * @param headers object，请求头数据
 * @returns Promise<object | string>
 */
export function httpGetJson(reqUrl, query, headers) {
    if (fetch) return fetchGetJson(reqUrl, query, headers)
    else return xhrGetJson(reqUrl, query, headers)
}

/* httpJsonPostJson 发送 HTTP POST 请求，发送 JSON 格式数据并以 JSON 格式解析响应数据
 * @param reqUrl string，资源地址
 * @param reqBody object，请求体
 * @param query object，URL Query
 * @param headers object，请求头数据
 * @returns Promise<object | string>
 */
export function httpJsonPostJson(reqUrl, reqBody, query, headers) {
    if (fetch) return fetchJsonPostJson(reqUrl, reqBody, query, headers)
    else return xhrJsonPostJson(reqUrl, reqBody, query, headers)
}

/* httpFormPostJson 发送 HTTP POST 请求，发送 form 格式数据并以 JSON 格式解析响应数据
 * @param reqUrl string，资源地址
 * @param reqBody object，请求体
 * @param query object，URL Query
 * @param headers object，请求头数据
 * @returns Promise<object | string>
 */
export function httpFormPostJson(reqUrl, reqBody, query, headers) {
    if (fetch) return fetchFormPostJson(reqUrl, reqBody, query, headers)
    else return xhrFormPostJson(reqUrl, reqBody, query, headers)
}

/* xhrGet 发送 HTTP GET 请求
 * @param reqUrl string，资源地址
 * @param query object，URL Query
 * @param headers object，请求头数据
 * @returns Promise<object | string>
 */
function xhrGet(reqUrl, query, headers) {
    return new Promise((resolve, reject) => {
        // 处理 query
        if (query) {
            if (reqUrl.indexOf('?') >= 0) reqUrl = `${reqUrl}&${objectToQueryString(query)}`
            else reqUrl = `${reqUrl}?${objectToQueryString(query)}`
        }

        let xhr = new XMLHttpRequest()
        xhr.open('GET', `${reqUrl}`)

        // 处理 header
        if (headers) Object.keys(headers).forEach((key) => xhr.setRequestHeader(key, headers[key]))
        xhr.setRequestHeader('X-Date', new Date().toUTCString())
        xhr.withCredentials = true

        // 处理回调
        xhr.addEventListener('load', () => {
                const headers = xhrHeadersToObject(xhr.getAllResponseHeaders())

                // 判断是否需要登陆。响应体类型是 json，响应码是 200006。
                const contentType = headers['Content-Type'] ??= headers['content-type']
                if (contentType && contentType.startsWith && contentType.startsWith('application/json')) {
                    try {
                        const rspBody = JSON.parse(xhr.response)
                        if (rspBody.code === constants.errCodeNeedLogin) {
                            const currentPath = location.pathname + location.search + location.hash

                            // 避免登陆页跳登陆页。
                            if (!currentPath.startsWith('/loginAndRegister'))
                                location.href = `/loginAndRegister?redirect=${encodeURIComponent(currentPath)}`

                            return
                        }
                    } catch (err) {
                        reject(`xhrGet failure: response body is not a json: ${err}: ${xhr.response}`)
                        return
                    }
                }

                resolve({
                    status: xhr.status,
                    headers: headers,
                    rspBody: xhr.response
                })
            }
        )
        xhr.addEventListener('error', () => reject(`xhrGet failure: ${reqUrl}`))

        // 发送请求
        xhr.send()
    })
}

/* xhrGetJson 发送 HTTP GET 请求，以 JSON 格式解析响应数据
 * @param reqUrl string，资源地址
 * @param query object，URL Query
 * @param headers object，请求头数据
 * @returns Promise<object | string>
 */
function xhrGetJson(reqUrl, query, headers) {
    return new Promise((resolve, reject) => {
        // 处理 query
        if (query) {
            if (reqUrl.indexOf('?') >= 0) reqUrl = `${reqUrl}&${objectToQueryString(query)}`
            else reqUrl = `${reqUrl}?${objectToQueryString(query)}`
        }

        let xhr = new XMLHttpRequest()
        xhr.open('GET', `${reqUrl}`)

        // 处理 header
        if (headers) Object.keys(headers).forEach((key) => xhr.setRequestHeader(key, headers[key]))
        xhr.setRequestHeader('X-Date', new Date().toUTCString())
        xhr.withCredentials = true

        // 处理回调
        xhr.addEventListener('load', () => {
            let rspBody
            try {
                rspBody = JSON.parse(xhr.response)
            } catch (err) {
                reject(`xhrGetJson failure: response body is not a json: ${err}: ${xhr.response}`)
            }

            // 判断是否需要登陆。响应码是 200006。
            if (rspBody.code === constants.errCodeNeedLogin) {
                const currentPath = location.pathname + location.search + location.hash

                // 避免登陆页跳登陆页。
                if (!currentPath.startsWith('/loginAndRegister'))
                    location.href = `/loginAndRegister?redirect=${encodeURIComponent(currentPath)}`

                return
            }

            resolve({
                status: xhr.status,
                headers: xhrHeadersToObject(xhr.getAllResponseHeaders()),
                rspBody: rspBody
            })
        })
        xhr.addEventListener('error', () => reject(`xhrGetJson failure: ${reqUrl}`))

        // 发送请求
        xhr.send()
    })
}

/* xhrJsonPostJson 发送 HTTP POST 请求，发送 JSON 格式数据并以 JSON 格式解析响应数据
 * @param reqUrl string，资源地址
 * @param reqBody object，请求体
 * @param query object，URL Query
 * @param headers object，请求头数据
 * @returns Promise<object | string>
 */
function xhrJsonPostJson(reqUrl, reqBody, query, headers) {
    return new Promise((resolve, reject) => {
        // 处理 reqBody
        try {
            reqBody = JSON.stringify(reqBody)
        } catch (err) {
            reject(`xhrJsonPostJson failure: ${err}`)
            return
        }

        // 处理 query
        if (query) {
            if (reqUrl.indexOf('?') >= 0) reqUrl = `${reqUrl}&${objectToQueryString(query)}`
            else reqUrl = `${reqUrl}?${objectToQueryString(query)}`
        }

        let xhr = new XMLHttpRequest()
        xhr.open('POST', `${reqUrl}`)
        xhr.setRequestHeader('Content-Type', 'application/json')

        // 处理 header
        if (headers) Object.keys(headers).forEach((key) => xhr.setRequestHeader(key, headers[key]))
        xhr.setRequestHeader('X-Date', new Date().toUTCString())
        xhr.withCredentials = true

        // 处理回调
        xhr.addEventListener('load', () => {
            let rspBody
            try {
                rspBody = JSON.parse(xhr.response)
            } catch (err) {
                reject(`xhrJsonPostJson failure: response body is not a json: ${err}: ${xhr.response}`)
            }

            // 判断是否需要登陆。响应码是 200006。
            if (rspBody.code === constants.errCodeNeedLogin) {
                const currentPath = location.pathname + location.search + location.hash

                // 避免登陆页跳登陆页。
                if (!currentPath.startsWith('/loginAndRegister'))
                    location.href = `/loginAndRegister?redirect=${encodeURIComponent(currentPath)}`

                return
            }

            resolve({
                status: xhr.status,
                headers: xhrHeadersToObject(xhr.getAllResponseHeaders()),
                rspBody: rspBody
            })
        })
        xhr.addEventListener('error', () => reject(`xhrJsonPostJson failure ${reqUrl}`))

        // 发送请求
        xhr.send(reqBody)
    })
}

/* xhrFormPostJson 发送 HTTP POST 请求，发送 form 格式数据并以 JSON 格式解析响应数据
 * @param reqUrl string，资源地址
 * @param reqBody object，请求体
 * @param query object，URL Query
 * @param headers object，请求头数据
 * @returns Promise<object | string>
 */
function xhrFormPostJson(reqUrl, reqBody, query, headers) {
    return new Promise((resolve, reject) => {
        // 处理 reqBody
        let formData
        if (reqBody instanceof FormData) {
            formData = reqBody
        } else {
            formData = new FormData()
            for (let [key, value] of Object.entries(reqBody)) formData.append(key, value)
        }

        // 处理 query
        if (query) {
            if (reqUrl.indexOf('?') >= 0) reqUrl = `${reqUrl}&${objectToQueryString(query)}`
            else reqUrl = `${reqUrl}?${objectToQueryString(query)}`
        }

        let xhr = new XMLHttpRequest()
        xhr.open('POST', `${reqUrl}`)

        // 处理 header
        if (headers) Object.keys(headers).forEach((key) => xhr.setRequestHeader(key, headers[key]))
        xhr.setRequestHeader('X-Date', new Date().toUTCString())
        xhr.withCredentials = true

        // 处理回调
        xhr.addEventListener('load', () => {
            let rspBody
            try {
                rspBody = JSON.parse(xhr.response)
            } catch (err) {
                reject(`xhrFormPostJson failure: response body is not a json: ${err}: ${xhr.response}`)
            }

            // 判断是否需要登陆。响应码是 200006。
            if (rspBody.code === constants.errCodeNeedLogin) {
                const currentPath = location.pathname + location.search + location.hash

                // 避免登陆页跳登陆页。
                if (!currentPath.startsWith('/loginAndRegister'))
                    location.href = `/loginAndRegister?redirect=${encodeURIComponent(currentPath)}`

                return
            }

            resolve({
                status: xhr.status,
                headers: xhrHeadersToObject(xhr.getAllResponseHeaders()),
                rspBody: rspBody
            })
        })
        xhr.addEventListener('error', () => reject(`xhrFormPostJson failure ${reqUrl}`))

        // 发送请求
        xhr.send(formData)
    })
}

/* fetchGet 发送 HTTP GET 请求
 * @param reqUrl string，资源地址
 * @param query object，URL Query
 * @param headers object，请求头数据
 * @returns Promise<object | string>
 */
function fetchGet(reqUrl, query, headers) {
    // 处理 query
    if (query) {
        if (reqUrl.indexOf('?') >= 0) reqUrl = `${reqUrl}&${objectToQueryString(query)}`
        else reqUrl = `${reqUrl}?${objectToQueryString(query)}`
    }

    return new Promise((resolve, reject) => {
        fetch(`${reqUrl}`, {
            method: 'GET',
            headers: mergeDateHeader(headers),
            credentials: 'same-origin'
        }).then(async (res) => {
                const headers = fetchHeadersToObject(res.headers)

                // 判断是否需要登陆。响应体类型是 json，响应码是 200006。
                const contentType = headers['Content-Type'] ??= headers['content-type']
                if (contentType && contentType.startsWith && contentType.startsWith('application/json')) {
                    try {
                        const rspBody = await res.json()
                        if (rspBody.code === constants.errCodeNeedLogin) {
                            const currentPath = location.pathname + location.search + location.hash

                            // 避免登陆页跳登陆页。
                            if (!currentPath.startsWith('/loginAndRegister'))
                                location.href = `/loginAndRegister?redirect=${encodeURIComponent(currentPath)}`

                            return
                        }
                    } catch (err) {
                        reject(`fetchGet failure: response body is not a json: ${err}`)
                        return
                    }
                }

                resolve({
                    status: res.status,
                    headers: headers,
                    rspBody: await res.arrayBuffer()
                })
            }
        ).catch((err) => reject(`fetchGet failure: ${err}`))
    })
}

/* fetchGetJson 发送 HTTP GET 请求，以 JSON 格式解析响应数据
 * @param reqUrl string，资源地址
 * @param query object，URL Query
 * @param headers object，请求头数据
 * @returns Promise<object | string>
 */
function fetchGetJson(reqUrl, query, headers) {
    // 处理 query
    if (query) {
        if (reqUrl.indexOf('?') >= 0) reqUrl = `${reqUrl}&${objectToQueryString(query)}`
        else reqUrl = `${reqUrl}?${objectToQueryString(query)}`
    }

    return new Promise((resolve, reject) => {
        fetch(`${reqUrl}`, {
            method: 'GET',
            headers: mergeDateHeader(headers),
            credentials: 'same-origin'
        }).then(async (res) => {
                const rspBody = await res.json()

                // 判断是否需要登陆。响应码是 200006。
                if (rspBody.code === constants.errCodeNeedLogin) {
                    const currentPath = location.pathname + location.search + location.hash

                    // 避免登陆页跳登陆页。
                    if (!currentPath.startsWith('/loginAndRegister'))
                        location.href = `/loginAndRegister?redirect=${encodeURIComponent(currentPath)}`

                    return
                }

                resolve({
                    status: res.status,
                    headers: fetchHeadersToObject(res.headers),
                    rspBody: rspBody
                })
            }
        ).catch((err) => reject(`fetchGetJson failure: ${err}`))
    })
}

/* fetchJsonPostJson 发送 HTTP POST 请求，发送 JSON 格式数据并以 JSON 格式解析响应数据
 * @param reqUrl string，资源地址
 * @param reqBody object，请求体
 * @param query object，URL Query
 * @param headers object，请求头数据
 * @returns Promise<object | string>
 */
function fetchJsonPostJson(reqUrl, reqBody, query, headers) {
    // 处理 query
    if (query) {
        if (reqUrl.indexOf('?') >= 0) reqUrl = `${reqUrl}&${objectToQueryString(query)}`
        else reqUrl = `${reqUrl}?${objectToQueryString(query)}`
    }

    // 处理 headers
    if (!headers) headers = {'Content-Type': 'application/json'}
    else if (!headers['Content-Type']) headers['Content-Type'] = 'application/json'

    return new Promise((resolve, reject) => {
        // 处理 reqBody
        try {
            reqBody = JSON.stringify(reqBody)
        } catch (err) {
            reject(`fetchJsonPostJson failure: ${err}`)
            return
        }

        fetch(reqUrl, {
            method: 'POST',
            headers: mergeDateHeader(headers),
            body: reqBody,
            credentials: 'same-origin'
        }).then(async (res) => {
                const rspBody = await res.json()

                // 判断是否需要登陆。响应码是 200006。
                if (rspBody.code === constants.errCodeNeedLogin) {
                    const currentPath = location.pathname + location.search + location.hash

                    // 避免登陆页跳登陆页。
                    if (!currentPath.startsWith('/loginAndRegister'))
                        location.href = `/loginAndRegister?redirect=${encodeURIComponent(currentPath)}`

                    return
                }

                resolve({
                    status: res.status,
                    headers: fetchHeadersToObject(res.headers),
                    rspBody: rspBody
                })
            }
        ).catch((err) => reject(`fetchJsonPostJson failure: ${err}`))
    })
}

/* fetchFormPostJson 发送 HTTP POST 请求，发送 form 格式数据并以 JSON 格式解析响应数据
 * @param reqUrl string，资源地址
 * @param reqBody object，请求体
 * @param query object，URL Query
 * @param headers object，请求头数据
 * @returns Promise<object | string>
 */
function fetchFormPostJson(reqUrl, reqBody, query, headers) {
    // 处理 query
    if (query) {
        if (reqUrl.indexOf('?') >= 0) reqUrl = `${reqUrl}&${objectToQueryString(query)}`
        else reqUrl = `${reqUrl}?${objectToQueryString(query)}`
    }

    // 处理 reqBody
    let formData
    if (reqBody instanceof FormData) {
        formData = reqBody
    } else {
        formData = new FormData()
        for (let [key, value] of Object.entries(reqBody)) formData.append(key, value)
    }

    return new Promise((resolve, reject) => {
        fetch(`${reqUrl}`, {
            method: 'POST',
            headers: mergeDateHeader(headers),
            body: formData,
            credentials: 'same-origin'
        }).then(async (res) => {
                const rspBody = await res.json()

                // 判断是否需要登陆。响应码是 200006。
                if (rspBody.code === constants.errCodeNeedLogin) {
                    const currentPath = location.pathname + location.search + location.hash

                    // 避免登陆页跳登陆页。
                    if (!currentPath.startsWith('/loginAndRegister'))
                        location.href = `/loginAndRegister?redirect=${encodeURIComponent(currentPath)}`

                    return
                }

                resolve({
                    status: res.status,
                    headers: fetchHeadersToObject(res.headers),
                    rspBody: rspBody
                })
            }
        ).catch((err) => reject(`fetchFormPostJson failure: ${err}`))
    })
}

/*
 * objectToQueryString 将对象转化成 URL Query 字符串
 * @param obj object
 * @returns string
 */
function objectToQueryString(obj) {
    let parts = []

    for (let [key, value] of Object.entries(obj)) {
        key = encodeURIComponent(key)
        if (Array.isArray(value)) value.forEach((value) => parts.push(`${key}=${encodeURIComponent(value)}`))
        else parts.push(`${key}=${encodeURIComponent(value)}`)
    }

    return parts.join('&')
}

/*
 * xhrHeadersToObject 将响应头字符串转化成对象
 * @param str string
 * @returns object
 */
function xhrHeadersToObject(str) {
    let obj = {}

    str.split('\r\n')
        .filter((value) => value.length > 0)
        .forEach((value) => {
            let [key, rest] = value.split(':', 2)
            obj[key.trim()] = rest.trim()
        })

    return obj
}

/*
 * fetchHeadersToObject 将响应头对象转化成对象
 * @param headers Headers
 * @returns object
 */
function fetchHeadersToObject(headers) {
    let obj = {}

    for (let [key, value] of headers.entries()) obj[key] = value

    return obj
}

/*
 * mergeDateHeader 合并 Date 请求头
 * @param headers object | undefined
 * @returns object
 */
function mergeDateHeader(headers) {
    if (!headers) headers = {}
    headers['X-Date'] = new Date().toUTCString()
    return headers
}
