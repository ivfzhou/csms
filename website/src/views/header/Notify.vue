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
import {CloseOutlined} from '@ant-design/icons-vue'
import {onBeforeMount, ref} from 'vue'
import {getLastNotification} from "@/api/notify.js"
import {message} from 'ant-design-vue'

const content = ref('')
const isShow = ref(false)
const close = () => isShow.value = false
const [messageApi, ContentHolder] = message.useMessage();

onBeforeMount(async () => {
  const rsp = await getLastNotification()
  if (rsp.status !== 200 || rsp.rspBody.code) {
    messageApi.error(`获取通知失败：${rsp.rspBody.code} ${rsp.rspBody.message}`)
  } else {
    content.value = rsp.rspBody.data.message
    if (rsp.rspBody.data.message) {
      isShow.value = true
    }
  }
})
</script>

<template>
  <div class="csms-header-notify" :class="{ 'notify-hidden': !isShow }">
    <ContentHolder/>
    <span class="csms-header-notify-content" v-html="content"></span>
    <CloseOutlined class="csms-header-notify-close" @click="close"/>
  </div>
</template>

<style scoped>
.csms-header-notify {
  background-color: LemonChiffon;
  display: flex;
  width: 100%;
  max-height: var(--notify-height);
  opacity: 1;
  overflow: hidden;
  transition: max-height .5s ease, opacity .5s ease;
  align-items: center;
  justify-content: center;
  border-radius: 3px;
  box-shadow: 2px 2px 8px rgba(0, 0, 0, 0.15);
  border: 1px solid #f0e68c;
  box-sizing: border-box;
}

.csms-header-notify.notify-hidden {
  max-height: 0;
  opacity: 0;
}

.csms-header-notify-content {
  flex-grow: 1;
  text-align: center;
  line-height: var(--notify-height);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-left: 10px;
  margin-right: 2px;
}

.csms-header-notify-close {
  cursor: pointer;
  margin-right: 2px;
  transition: transform .2s cubic-bezier(0, 1, 1, 1);
}

.csms-header-notify-close:hover {
  color: blue;
  transform: scale(80%, 80%);
}
</style>
