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
import {useMessageStore} from "@/stores/message.js"

// 保存消息提示变量。
const {message} = App.useApp()
const messageStore = useMessageStore()
messageStore.$patch({message})

// 获取用户信息与登陆。
const userInfoStore = useUserInfoStore()
onBeforeMount(async () => {
  const {ok, data} = await getUserInformation()
  if (ok) userInfoStore.$patch(data)
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
  /* 设置头部背景色为白色 */
  background-color: #fff;
}

.csms-header-inner {
  /* 为子元素的绝对定位提供定位上下文 */
  position: relative;
  /* 使用弹性布局 */
  display: flex;
  /* 子元素沿垂直方向排列 */
  flex-direction: column;
  /* 子元素在垂直方向上居中 */
  justify-content: center;
  /* 子元素在水平方向上居中 */
  align-items: center;
  /* 头部高度由 CSS 变量 --header-height 控制 */
  height: var(--header-height);
  /* 内容最大宽度由 CSS 变量 --content-width 控制 */
  max-width: var(--content-width);
  /* 水平居中 */
  margin: 0 auto;
}

.csms-body {
  max-width: var(--content-width);
  margin: 10px auto 0;
  background-color: #fff;
}
</style>
