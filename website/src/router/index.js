import {createRouter, createWebHistory} from 'vue-router'

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/',
            name: 'index',
            component: () => import('@/views/Index.vue')
        },
        {
            path: '',
            redirect: {name: 'index'}
        },
        {
            path: '/index.html',
            redirect: {name: 'index'}
        },
        {
            path: '/index',
            redirect: {name: 'index'}
        }
    ],
})

export default router
