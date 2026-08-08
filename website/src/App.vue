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
import {App, ConfigProvider, StyleProvider, theme} from "ant-design-vue"
import {computed, ref} from 'vue'
import zhCN from 'ant-design-vue/es/locale/zh_CN'
import enUS from 'ant-design-vue/es/locale/en_US'
import dayjs from 'dayjs'
import {useLocaleStore} from "@/stores/locale.js"
import {useThemeStore} from "@/stores/theme.js"

// ant 组件本地化。
const locale = ref(zhCN)
const toggleLocale = () => {
  locale.value = locale.value === zhCN ? enUS : zhCN
  dayjs.locale(locale.value)
}
const localStore = useLocaleStore()
localStore.$patch({toggleLocale})

// ant 组件主题。
const isDark = ref(false)
const themeConfig = computed(() => ({
  algorithm: isDark.value ? theme.darkAlgorithm : theme.defaultAlgorithm,
  components: {
    Layout: {
      colorBgHeader: isDark.value ? '#1f1f1f' : '#ffffff'
    }
  }
}))
const toggleTheme = () => isDark.value = !isDark.value
const themeStore = useThemeStore()
themeStore.$patch({toggleTheme, isDark})
</script>

<template>
  <StyleProvider hash-priority="low">
    <ConfigProvider :locale="locale" :theme="themeConfig">
      <App>
        <RouterView/>
      </App>
    </ConfigProvider>
  </StyleProvider>
</template>

<style scoped></style>
