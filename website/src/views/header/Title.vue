<!--
Copyright (c) 2023 ivfzhou
website is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
-->

<script setup>
import {Avatar} from 'ant-design-vue'
import {UserOutlined} from '@ant-design/icons-vue'
</script>

<template>
  <div class="csms-header-title">
    <span class="csms-header-title-logo"></span>
    <span class="csms-header-title-name">数字证书签名及管理系统</span>
    <RouterLink class="csms-header-title-index" to="/index">工作台</RouterLink>
    <span class="csms-header-title-avatar">
      <Avatar>
        <template #icon>
          <UserOutlined/>
        </template>
      </Avatar>
    </span>
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
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  position: relative;
}

.csms-header-title-logo {
  width: 30px;
  height: 30px;
  background-image: url('/favicon.ico');
  display: inline-block;
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
}

.csms-header-title-name {
  display: inline-block;
  font-size: 22px;
  line-height: 30px;
  color: #2d3845;
  flex-grow: 1;
  margin-left: 6px;
  padding-top: 6px;
}

.csms-header-title-index {
  display: inline-block;
  font-size: 20px;
  line-height: 30px;
  vertical-align: middle;
  color: #7c8b99;
  padding-top: 6px;
}

.csms-header-title-index:hover {
  color: #0094f7;
}

.csms-header-title-avatar {
  margin-left: 20px;
  display: inline-flex;
  cursor: pointer;
}

.csms-header-title-usermenu {
  list-style-type: none;
  visibility: hidden;
  position: absolute;
  top: 100%;
  right: 0;
  z-index: 1;
  margin: 0;
  padding: 8px 0;
  background: #fff;
  border: 1px solid #e8e8e8;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  white-space: nowrap;
}

.csms-header-title-usermenu li {
  padding: 8px 20px;
  cursor: pointer;
}

.csms-header-title-usermenu li:hover {
  background-color: #f5f5f5;
}

.csms-header-title-avatar:hover ~ .csms-header-title-usermenu,
.csms-header-title-usermenu:hover {
  visibility: visible;
}
</style>
