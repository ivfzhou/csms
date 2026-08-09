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
import {computed, provide, ref} from 'vue'
import zhCN from 'ant-design-vue/es/locale/zh_CN'
import enUS from 'ant-design-vue/es/locale/en_US'
import dayjs from 'dayjs'
import constants from '@/utils/constants.js'

// ant 组件本地化。
const antLocale = ref(zhCN)
const toggleAntLocale = () => {
  antLocale.value = antLocale.value === zhCN ? enUS : zhCN
  dayjs.locale(antLocale.value)
}

// ant 组件主题。
const isDark = ref(false)
const antTheme = computed(() => ({
  algorithm: isDark.value ? theme.darkAlgorithm : theme.defaultAlgorithm,
  components: {
    Layout: {
      colorBgHeader: isDark.value ? '#1f1f1f' : '#ffffff'
    }
  }
}))
const toggleAntTheme = () => isDark.value = !isDark.value

// ant 组件尺寸。
const antComponentSize = ref('middle');
const toggleAntComponentSize = () => {
  switch (antComponentSize.value) {
    case 'small':
      antComponentSize.value = 'middle'
      break
    case 'meddle':
      antComponentSize.value = 'large'
      break
    case 'large':
      antComponentSize.value = 'small'
      break
    default:
      antComponentSize.value = 'middle'
      break
  }
}

// ant 组件方向。
const antComponentDirection = ref('ltr')
const toggleAntComponentDirection = () => {
  antComponentDirection.value = antComponentDirection.value === 'ltr' ? 'rtl' : 'ltr'
}

provide(constants.keyAntConfig, {toggleAntLocale, toggleAntTheme, toggleAntComponentSize, toggleAntComponentDirection})
</script>

<template>
  <StyleProvider hash-priority="low">
    <ConfigProvider :locale="antLocale" :theme="antTheme" :component-size="antComponentSize"
                    :direction="antComponentDirection">
      <App>
        <RouterView/>
      </App>
    </ConfigProvider>
  </StyleProvider>
</template>

<style scoped></style>
