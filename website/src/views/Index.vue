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
import {onBeforeMount} from 'vue'
import {App} from "ant-design-vue"
import Notify from '@/views/header/Notify.vue'
import Title from '@/views/header/Title.vue'
import {getUserInformation} from "@/api/user_api.js"
import {useUserInfoStore} from "@/stores/userInfo.js"
import {isSuccessHttpCode} from "@/utils/utils.js"
import {useMessageStore} from "@/stores/message.js"

// 保存消息提示变量。
const {message} = App.useApp()
const messageStore = useMessageStore()
messageStore.$patch({message})

// 获取用户信息与登陆。
const userInfoStore = useUserInfoStore()
onBeforeMount(async () => {
  try {
    const rsp = await getUserInformation()
    if (!isSuccessHttpCode(rsp.status)) {
      message.error(`获取用户信息失败 ${rsp}`)
      return
    }
    if ((!rsp.rspBody || rsp.rspBody.code > 0) && rsp.rspBody.message) {
      message.warning(`${rsp.rspBody.code} ${rsp.rspBody.message}`)
      return
    }
    if (!rsp.rspBody.data) {
      message.warning(`未获取到获取用户信息 ${rsp}`)
      return
    }
    userInfoStore.$patch(rsp.rspBody.data)
  } catch (err) {
    message.error(`获取用户信息异常 ${err}`)
  }
})
</script>

<template>
  <div class="csms-root">
    <div class="csms-header">
      <div class="csms-header-inner">
        <Notify/>
        <Title/>
      </div>
    </div>
    <div class="csms-body">
      <RouterView/>
    </div>
  </div>
</template>

<style scoped>
.csms-header {
  background-color: #fff;
}

.csms-header-inner {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: var(--header-height);
  max-width: var(--content-width);
  margin: 0 auto;
}

.csms-body {
  max-width: var(--content-width);
  margin: 10px auto 0;
  background-color: #fff;
}
</style>
