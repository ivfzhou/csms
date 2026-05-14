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
import {CloseOutlined} from '@ant-design/icons-vue'
import {onBeforeMount, ref} from 'vue'
import {getLastNotification} from "@/api/notify.js"

const message = ref('')
const isShow = ref(false)

// JS/CSS 实现过渡
// const isNotifyHidden = ref(false)

const closeMessage = () => {
  // JS/CSS 实现过渡：isNotifyHidden.value = true
  isShow.value = false
}

// JS/CSS 实现过渡
// const notifyClosed = (e) => {
//   if (e.currentTarget.classList.contains('notify-hidden')) {
//     isShow.value = false
//     isNotifyHidden.value = false
//   }
// }

onBeforeMount(async () => {
  const rsp = await getLastNotification()
  if (rsp.status !== 200 || rsp.rspBody.data.code) {

  } else {
    message.value = rsp.rspBody.data.message
    if (rsp.rspBody.data.message) {
      // JS/CSS 实现过渡
      // isNotifyHidden.value = true
      // isShow.value = true
      // requestAnimationFrame(() => {
      //   requestAnimationFrame(() => {
      //     isNotifyHidden.value = false
      //   })
      // })

      isShow.value = true
    }
  }
})
</script>

<template>
  <Transition name="notify">
    <!-- JS/CSS 实现过渡：<div v-if="isShow" class="csms-header-notify" :class="{ 'notify-hidden': isNotifyHidden}" @transitionend="notifyClosed"> -->
    <div v-if="isShow" class="csms-header-notify">
      <span class="csms-header-notify-content" v-html="message"></span>
      <CloseOutlined class="csms-header-notify-close" @click="closeMessage"/>
    </div>
  </Transition>
</template>

<style scoped>
.notify-leave-active {
  transition-property: opacity, transform;
  transition-duration: .5s;
  transition-timing-function: ease;
  opacity: 1;
}

.notify-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

.notify-enter-active {
  transition-property: opacity, transform;
  transition-duration: .5s;
  transition-timing-function: ease;
  opacity: 0;
  transform: translateY(-10px);
}

.notify-enter-to {
  opacity: 1;
  transform: translateY(0);
}

.csms-header-notify {
  background-color: LemonChiffon;
  height: var(--notify-height);
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 3px;
  box-shadow: 2px 2px 8px rgba(0, 0, 0, 0.15);
  border-width: 1px;
  border-color: #f0e68c;
  border-style: solid;

  /*
  JS/CSS 实现过渡
  opacity: 1;
  transform: translateY(0);
  transition: opacity .4s ease, transform .4s ease;
  */
}

/*
JS/CSS 实现过渡
.csms-header-notify.notify-hidden {
  opacity: 0;
  transform: translateY(-10px);
}
*/

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
  right: 12px;
  margin-right: 2px;
  transition-property: transform;
  transition-duration: .2s;
  transition-timing-function: cubic-bezier(0, 1.29, 1, .99);
}

.csms-header-notify-close:hover {
  color: blue;
  transform: scale(93%, 93%);
}
</style>
