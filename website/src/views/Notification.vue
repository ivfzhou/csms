<script setup>
import {CloseOutlined} from '@ant-design/icons-vue'
import {onBeforeMount, ref} from 'vue'
import {getLastNotification} from "@/api/notification.js";

const message = ref('')
const show = ref(false)
const closeMessage = () => show.value = false

onBeforeMount(async () => {
  const rsp = await getLastNotification()
  if (rsp.status !== 200 || rsp.rspBody.data.code) {

  } else {
    message.value = rsp.rspBody.data.message
    if (rsp.rspBody.data.message) show.value = true
  }
})
</script>

<template>
  <div class="csms-header-notify" v-if="show">
    <span class="csms-header-notify-content" v-html="message"></span>
    <CloseOutlined class="csms-header-notify-close" @click="closeMessage"/>
  </div>
</template>

<style scoped></style>
