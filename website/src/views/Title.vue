<script setup>
import {Avatar} from 'ant-design-vue'
import {UserOutlined} from '@ant-design/icons-vue'
</script>

<template>
  <div class="csms-header-title">
    <span class="csms-header-title-logo"></span>
    <span class="csms-header-title-name">数字证书签名及管理系统</span>
    <RouterLink class="csms-header-title-index" to="/index">工作台</RouterLink>
    <Avatar class="csms-header-title-avatar">
      <template #icon>
        <UserOutlined/>
      </template>
    </Avatar>
  </div>
</template>

<style scoped></style>
