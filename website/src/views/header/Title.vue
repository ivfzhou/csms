<!--
Copyright (c) 2023 ivfzhou
csms is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
-->

<script setup>
import {ref} from "vue"
import {useUserInfoStore} from "@/stores/userInfo.js"
import {Avatar} from 'ant-design-vue'
import {UserOutlined} from '@ant-design/icons-vue'
import SvgIcon from '@jamescoyle/vue-icon'
import {mdiAccount, mdiLayersOutline, mdiLogout} from '@mdi/js'
import {userLogout} from "@/api/user_api.js"
import {isNavigationFailure, NavigationFailureType, useRouter} from "vue-router"

// 控制展示已登陆的信息。
const isShowUserInfo = ref(true)
const userInfoStore = useUserInfoStore()
isShowUserInfo.value = userInfoStore.userInfo.nameEn !== undefined
userInfoStore.$subscribe((_, state) => isShowUserInfo.value = state && state.nameEn)

// 退出登陆。
const router = useRouter()

async function logoutHandler() {
  const {ok} = await userLogout()
  if (ok) {
    // 成功退出登陆后，跳转到首页。
    setTimeout(async () => {
      const res = await router.push('/index')
      if (isNavigationFailure(res, NavigationFailureType.duplicated)) {
        location.reload()
      }
    }, 1000)
  }
}
</script>

<template>
  <div class="csms-header-title">
    <span class="csms-header-title-logo"></span>
    <span class="csms-header-title-name">数字证书签名及管理系统</span>
    <RouterLink v-show="isShowUserInfo" class="csms-header-title-index" to="/index">工作台</RouterLink>
    <Avatar v-show="isShowUserInfo" class="csms-header-title-avatar">
      <template #icon>
        <UserOutlined/>
      </template>
    </Avatar>
    <ul class="csms-header-title-usermenu">
      <li v-ripple>
        <SvgIcon :path="mdiLayersOutline" type="mdi"/>
        后台管理
      </li>
      <li v-ripple>
        <SvgIcon :path="mdiAccount" type="mdi"/>
        个人信息
      </li>
      <li v-ripple @click="logoutHandler">
        <SvgIcon :path="mdiLogout" type="mdi"/>
        退出登陆
      </li>
    </ul>
  </div>
</template>

<style scoped>
.csms-header-title {
  width: 100%;
  height: var(--title-height);
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  justify-content: center;
  position: relative;
}

.csms-header-title-logo {
  width: 30px;
  height: 30px;
  background-image: url('/favicon.ico');
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
}

.csms-header-title-name {
  font-size: 22px;
  color: #2d3845;
  flex-grow: 1;
  margin-left: 10px;
  align-self: baseline;
  user-select: none;
}

.csms-header-title-index {
  font-size: 20px;
  color: #7c8b99;
  margin-right: 10px;
  align-self: baseline;
}

.csms-header-title-index:hover {
  color: #0094f7;
  cursor: pointer;
}

.csms-header-title-avatar {
  cursor: pointer;
}

.csms-header-title-usermenu {
  list-style-type: none;
  visibility: hidden;
  position: absolute;
  top: 100%;
  right: 0;
  z-index: 1;
  background: #fff;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  box-shadow: 2px 2px 8px rgba(0, 0, 0, 0.15);
  margin: 0;
  padding: 0;
  opacity: 0;
  transform: translateY(-20px);
  transition: transform .3s ease, opacity .3s ease;
}

.csms-header-title-usermenu li {
  cursor: pointer;
  margin: 0;
  padding: 8px 24px;
  display: flex;
  align-items: center;
  user-select: none;
}

.csms-header-title-usermenu li:hover {
  background-color: #f5f5f5;
}

.csms-header-title-avatar:hover ~ .csms-header-title-usermenu,
.csms-header-title-usermenu:hover {
  visibility: visible;
  opacity: 1;
  transform: translateY(0);
}
</style>
