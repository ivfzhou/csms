<script setup>
import {ref} from 'vue'
import {ConfigProvider, Layout, LayoutContent, LayoutHeader, StyleProvider} from "ant-design-vue";
import zhCN from 'ant-design-vue/es/locale/zh_CN'
import enUS from 'ant-design-vue/es/locale/en_US'
import dayjs from 'dayjs'
import Notification from '@/views/Notification.vue'
import Title from '@/views/Title.vue'

const locale = ref(zhCN)
const toggleLocale = () => {
  locale.value = locale.value === zhCN ? enUS : zhCN
  dayjs.locale(locale.value)
}

</script>

<template>
  <div class="csms-root">
    <StyleProvider hash-priority="low">
      <ConfigProvider :locale="locale">
        <Layout>
          <LayoutHeader class="csms-header">
            <Notification/>
            <Title/>
          </LayoutHeader>
          <LayoutContent></LayoutContent>
        </Layout>
      </ConfigProvider>
    </StyleProvider>
  </div>
</template>

<style scoped></style>
