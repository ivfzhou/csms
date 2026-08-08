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

import {createRouter, createWebHistory} from 'vue-router'

const router = createRouter({
    // import.meta 本身是 ES2020 标准中定义的 JavaScript 语法，它是模块元信息对象，由 JS 宿主（浏览器/Node.js）实现。
    // import.meta.env 是 Vite 在构建时注入的，它包含了环境变量（类似 Webpack 中的 process.env）。
    // import.meta.env.BASE_URL 是 Vite 特有的环境变量，对应 Vite 配置中的 base 选项（默认是 /），用于指定应用部署的基础路径。
    history: createWebHistory(import.meta.env.BASE_URL),
    strict: true,
    routes: [
        {
            path: '/',
            // import() 是 JavaScript 语法，它是 ES2020 标准中定义的动态导入（Dynamic Import）语法，返回一个 Promise，用于在运行时动态加载模块。
            component: () => import('@/views/Index.vue'),
            children: [
                {
                    path: '',
                    component: () => import('@/views/body/Dashboard.vue')
                },
                {
                    path: 'loginAndRegister',
                    component: () => import('@/views/body/LoginAndRegister.vue'),
                    props: route => ({redirect: route.query.redirect, isLogin: route.query.isLogin}),
                }
            ]
        },
        {
            path: '/index.html',
            redirect: '/'
        },
        {
            path: '/index',
            redirect: '/'
        }
    ],
})

export default router
