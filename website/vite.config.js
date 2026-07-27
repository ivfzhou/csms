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

import {fileURLToPath, URL} from 'node:url'

import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
    plugins: [
        vue(),
        vueDevTools(),
    ],
    resolve: {
        alias: {
            // 将 @ 映射到 src 目录的绝对路径，方便用 @/components/xxx 代替相对路径导入。
            '@': fileURLToPath(new URL('./src', import.meta.url))
        },
    },
    server: {
        // 开发环境代理：将以 /backend 开头的请求转发到后端服务，解决跨域问题。
        proxy: {
            '/backend': {
                target: 'https://127.0.0.1',   // 后端目标地址。
                changeOrigin: true,             // 修改请求头中的 Host 为目标地址。
                secure: false,                  // 不验证 HTTPS 证书（允许自签名证书）。
            }
        }
    }
})
