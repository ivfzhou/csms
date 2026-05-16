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
import {computed, provide, ref} from 'vue'
import {ConfigProvider, StyleProvider, theme} from "ant-design-vue"
import zhCN from 'ant-design-vue/es/locale/zh_CN'
import enUS from 'ant-design-vue/es/locale/en_US'
import dayjs from 'dayjs'
import Notify from '@/views/header/Notify.vue'
import Title from '@/views/header/Title.vue'
import constants from '@/utils/constants.js'

const locale = ref(zhCN)
const toggleLocale = () => {
  locale.value = locale.value === zhCN ? enUS : zhCN
  dayjs.locale(locale.value)
}

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

provide(constants.isDark, isDark)
provide(constants.toggleTheme, toggleTheme)
provide(constants.toggleLocale, toggleLocale)

</script>

<template>
  <StyleProvider hash-priority="low">
    <ConfigProvider :locale="locale" :theme="themeConfig">
      <div class="csms-header">
        <div class="csms-header-inner">
          <Notify/>
          <Title/>
        </div>
      </div>
      <div class="csms-body">
        <RouterView/>
      </div>
    </ConfigProvider>
  </StyleProvider>
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
  margin: 0 auto;
}
</style>
