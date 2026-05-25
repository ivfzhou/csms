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

// 控制展示已登陆的信息。
const isShowUserInfo = ref(true)
const userInfoStore = useUserInfoStore()
isShowUserInfo.value = userInfoStore.userInfo.nameEn !== undefined
userInfoStore.$subscribe((_, state) => isShowUserInfo.value = state && state.nameEn)
</script>

<template>
  <div class="csms-header-title">
    <span class="csms-header-title-logo"></span>
    <span class="csms-header-title-name">数字证书签名及管理系统</span>
    <RouterLink class="csms-header-title-index" to="/index" v-show="isShowUserInfo">工作台</RouterLink>
    <Avatar class="csms-header-title-avatar" v-show="isShowUserInfo">
      <template #icon>
        <UserOutlined/>
      </template>
    </Avatar>
    <ul class="csms-header-title-usermenu">
      <li>后台管理</li>
      <li>个人信息</li>
      <li>退出登陆</li>
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
  height: var(--title-height);
  background-image: url('/favicon.ico');
  display: inline-block;
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
}

.csms-header-title-name {
  display: inline-flex;
  font-size: 22px;
  height: var(--title-height);
  color: #2d3845;
  flex-grow: 1;
  margin-left: 10px;
  align-items: flex-end;
}

.csms-header-title-index {
  display: inline-flex;
  font-size: 20px;
  height: var(--title-height);
  color: #7c8b99;
  align-items: flex-end;
  margin-right: 20px;
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
  border-radius: 3px;
  box-shadow: 2px 2px 8px rgba(0, 0, 0, 0.15);
  margin: 0;
  padding: 0;
}

.csms-header-title-usermenu li {
  cursor: pointer;
  margin: 0;
  padding: 8px 16px;
}

.csms-header-title-usermenu li:hover {
  background-color: #f5f5f5;
}

.csms-header-title-avatar:hover ~ .csms-header-title-usermenu, .csms-header-title-usermenu:hover {
  visibility: visible;
}
</style>
