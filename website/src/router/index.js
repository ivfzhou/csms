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
    history: createWebHistory(import.meta.env.BASE_URL),
    strict: true,
    routes: [
        {
            path: '/',
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
