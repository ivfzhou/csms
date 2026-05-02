<script setup>
import {CloseOutlined} from '@ant-design/icons-vue'
import {ref} from 'vue'

const message = ref('test')
const show = ref(true)
const closeMessage = () => show.value = false
</script>

<template>
  <div class="csms-header-notify" v-if="show">
    <span class="csms-header-notify-content" v-html="message"></span>
    <CloseOutlined class="csms-header-notify-close" @click="closeMessage"/>
  </div>
</template>

<style scoped></style>
