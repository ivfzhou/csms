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
import {useRoute, useRouter} from "vue-router"
import {onBeforeMount, reactive, ref, watch} from "vue"
import {App, Button, Form, FormItem, Input, InputPassword, Modal, Upload} from 'ant-design-vue'
import {
  HomeOutlined,
  IdcardOutlined,
  LoadingOutlined,
  LockOutlined,
  PlusOutlined,
  UserOutlined
} from "@ant-design/icons-vue"
import {useUserInfoStore} from "@/stores/userInfo.js"
import {getUserInformation, userLogin, userRegister} from "@/api/user_api.js"
import {getBase64} from "@/utils/utils.js"

// 定义组件 props。
const props = defineProps({
  redirect: {
    type: String,
    default: '/'
  },
  isLogin: Boolean
})

// 校验是否已经登陆了，若已登陆就跳转到原页面。
const {message} = App.useApp()
const userInfoStore = useUserInfoStore()
const router = useRouter()
onBeforeMount(async () => {
  if (userInfoStore.userInfo) {
    const {ok, data} = await getUserInformation()
    if (ok) {
      userInfoStore.$patch(data)
      message.info(`已登陆，即将跳转`)
      setTimeout(() => router.push(props.redirect), 2000)
    }
  }
})

// 控制登陆/注册模式下浏览器链接地址。
const isLogin = ref(props.isLogin)
const route = useRoute()
watch(isLogin, (value) => {
  router.replace({query: {...route.query, isLogin: value ? '' : undefined}})
})

// 表单数据、校验和提交逻辑。
const formState = reactive({
  avatar: [],
  username: '',
  nickname: '',
  password: '',
  passwordConfirm: '',
  department: ''
})
const formRules = {
  avatar: {
    required: true,
    message: '请上传头像'
  },
  username: {
    trigger: 'change',
    validator: async (_, value) => {
      if (!value) return Promise.reject('请输入用户名')

      if (value.length < 6 || value.length > 32) return Promise.reject('用户名至少 6 位字符，最多 32 位字符')

      if (!/^[a-zA-Z][a-zA-Z0-9]+$/.test(value)) return Promise.reject('用户名由数字和字母组成，第一个字符需为字母')

      return Promise.resolve()
    }
  },
  nickname: {
    trigger: 'change',
    validator: async (_, value) => {
      if (!value) return Promise.reject('请输入中文名')

      if (value.length < 2 || value.length > 16) return Promise.reject('中文名至少 2 位字符，最多 16 位字符')

      if (!/^\p{Unified_Ideograph}+$/u.test(value)) return Promise.reject('中文名只能包含中文字符')

      return Promise.resolve()
    }
  },
  password: {
    trigger: 'change',
    validator: async (_, value) => {
      if (!value) return Promise.reject('请输入密码')

      if (value.length < 6) return Promise.reject('密码至少 6 位字符')

      if (/\p{C}/u.test(value) || value.includes('�')) return Promise.reject('密码只能包含可打印字符')

      return Promise.resolve()
    }
  },
  passwordConfirm: {
    trigger: 'change',
    validator: async (_, value) => {
      if (!value) return Promise.reject('请输入密码')

      if (value.length < 6) return Promise.reject('密码至少 6 位字符')

      if (/\p{C}/u.test(value) || value.includes('�')) return Promise.reject('密码只能包含可打印字符')

      if (value !== formState.password) return Promise.reject('两次输入的密码须相同')

      return Promise.resolve()
    }
  },
  department: {
    trigger: 'change',
    validator: async (_, value) => {
      if (!value) return Promise.reject('请输入用户所在部门')

      if (value.length > 1024) return Promise.reject('部门字符数不能超过 1024 位')

      if (/\p{C}/u.test(value) || value.includes('�')) return Promise.reject('部门只能包含可打印字符')

      return Promise.resolve()
    }
  }
}
const finishForm = async (value) => {
  if (isLogin.value) {
    const {ok, _} = await userLogin({nameEn: value.username, password: value.password})
    if (ok) {
      // 跳转到原页面。
      setTimeout(() => router.push(props.redirect), 500)
    }
    return
  }

  const data = new FormData()
  data.append('nameZh', formState.nickname)
  data.append('nameEn', formState.username)
  data.append('password', formState.password)
  data.append('passwordConfirmation', formState.passwordConfirm)
  data.append('department', formState.department)
  data.append('avatar', formState.avatar[0], formState.avatar[0].name)
  const {ok} = await userRegister(data)
  if (ok) {
    // 切换到登陆页面。
    isLogin.value = true
  }
}

// 控制输入框过渡效果。
const onBeforeEnter = (el) => {
  const computedStyle = getComputedStyle(el)
  el.style.overflow = 'hidden'
  el.style.height = '0'
  el.style.opacity = '0'
  el.style.margin = '0'
  el.style.padding = '0'
  el.style.transition = 'opacity .3s linear, height .3s linear, margin .3s linear, padding .3s linear'

  el.dataset.originPadding = computedStyle.padding
  el.dataset.originMargin = computedStyle.margin
}
const onEnter = (el, done) => {
  el.style.height = `${el.scrollHeight}px`
  el.style.padding = el.dataset.originPadding
  el.style.margin = el.dataset.originMargin
  el.style.opacity = '1'

  const onTransitionEnd = (event) => {
    if (event.propertyName === 'height') {
      el.removeEventListener('transitionend', onTransitionEnd)

      delete el.dataset.originPadding
      delete el.dataset.originMargin

      done()
    }
  }
  el.addEventListener('transitionend', onTransitionEnd)
}
const onLeave = (el, done) => {
  const computedStyle = getComputedStyle(el)
  el.style.overflow = 'hidden'
  el.style.height = `${el.scrollHeight}px`
  el.style.opacity = '1'
  el.style.margin = computedStyle.margin
  el.style.padding = computedStyle.padding
  el.style.transition = 'opacity .3s linear, height .3s linear, margin .3s linear, padding .3s linear'

  requestAnimationFrame(() => {
    el.style.height = '0'
    el.style.opacity = '0'
    el.style.margin = '0'
    el.style.padding = '0'
  })

  const onTransitionEnd = (event) => {
    if (event.propertyName === 'height') {
      el.removeEventListener('transitionend', onTransitionEnd)
      done()
    }
  }
  el.addEventListener('transitionend', onTransitionEnd)
}

// 控制头像数据。
const isAvatarLoading = ref(false)
const avatarBeforeUpload = async (file, fileList) => {
  try {
    isAvatarLoading.value = true

    // 校验文件格式和大小。
    const maximumFileSize = 1 << 20
    if (file.size > maximumFileSize) {
      message.warning(`头像过大，最大允许 ${maximumFileSize} 字节`)
      formState.avatar = []
      return false
    }
    if (!['image/jpeg', 'image/png'].includes(file.type)) {
      message.warning(`头像格式非法，只允许上传 .jpg/.png 格式头像`)
      formState.avatar = []
      return false
    }

    // 手动生成预览 URL，让 Upload 组件能显示图片
    file.url = URL.createObjectURL(file)

    formState.avatar = [file]
    return false
  } finally {
    isAvatarLoading.value = false
  }
}
const removeAvatar = (file) => formState.avatar.shift()

// 控制头像预览。
const previewVisible = ref(false)
const previewTitle = ref('')
const previewImage = ref('')
const handleCancel = () => {
  previewVisible.value = false
  previewTitle.value = ''
  previewImage.value = ''
}
const handlePreview = async file => {
  previewImage.value = file.url
  if (!previewImage.value) {
    const fileObj = file.originFileObj || file
    previewImage.value = await getBase64(fileObj)
  }
  previewTitle.value = file.name || file.url.substring(file.url.lastIndexOf('/') + 1)
  previewVisible.value = true
};
</script>

<template>
  <div class="csms-body-loginandregister">
    <Transition name="title" mode="out-in">
      <div class="csms-body-loginandregister-title" :key="isLogin">{{ isLogin ? '用户登陆' : '注册用户' }}</div>
    </Transition>
    <Form class="csms-body-loginandregister-form" @finish="finishForm" :rules="formRules" :model="formState"
          :label-col="{span: 6}" :wrapper-col="{span: 18}" validateFirst autocomplete="on">
      <Transition @beforeEnter="onBeforeEnter" @enter="onEnter" @leave="onLeave">
        <FormItem label="头像" name="avatar" required v-if="!isLogin">
          <Upload :fileList="formState.avatar" listType="picture-card" :beforeUpload="avatarBeforeUpload"
                  @remove="removeAvatar" accept="image/png,image/jpeg,image/jpg" @preview="handlePreview">
            <div v-if="formState.avatar.length <= 0">
              <LoadingOutlined v-if="isAvatarLoading"/>
              <PlusOutlined v-else/>
            </div>
          </Upload>
        </FormItem>
      </Transition>
      <FormItem label="用户名" name="username" hasFeedback required validateFirst>
        <Input v-model:value="formState.username" autocomplete="username"
               placeholder="请输入用户名，6 到 32 位字符，由数字和字母组成，第一个字符需为字母">
          <template #prefix>
            <UserOutlined/>
          </template>
        </Input>
      </FormItem>
      <Transition @beforeEnter="onBeforeEnter" @enter="onEnter" @leave="onLeave">
        <FormItem label="中文名" name="nickname" hasFeedback validateFirst required v-if="!isLogin">
          <Input v-model:value="formState.nickname" placeholder="请输入中文名，2 到 16 位汉字">
            <template #prefix>
              <IdcardOutlined/>
            </template>
          </Input>
        </FormItem>
      </Transition>
      <FormItem label="密码" name="password" hasFeedback validateFirst required>
        <InputPassword v-model:value="formState.password" autocomplete="current-password" :visibilityToggle="isLogin"
                       placeholder="请输入密码，至少 6 位可打印字符">
          <template #prefix>
            <LockOutlined/>
          </template>
        </InputPassword>
      </FormItem>
      <Transition @beforeEnter="onBeforeEnter" @enter="onEnter" @leave="onLeave">
        <FormItem label="确认密码" name="passwordConfirm" hasFeedback validateFirst required v-if="!isLogin">
          <InputPassword v-model:value="formState.passwordConfirm" placeholder="请再次输入密码"
                         :visibilityToggle="isLogin">
            <template #prefix>
              <LockOutlined/>
            </template>
          </InputPassword>
        </FormItem>
      </Transition>
      <Transition @beforeEnter="onBeforeEnter" @enter="onEnter" @leave="onLeave" required>
        <FormItem label="部门" name="department" hasFeedback validateFirst v-if="!isLogin">
          <Input v-model:value="formState.department" placeholder="请输入部门信息，最多 1024 个字符，组织单元间以 / 分隔">
            <template #prefix>
              <HomeOutlined/>
            </template>
          </Input>
        </FormItem>
      </Transition>
      <FormItem :wrapperCol="{span: 18, offset: 6}">
        <div class="csms-body-loginandregister-form-button-inner">
          <Button type="primary" htmlType="submit">{{ isLogin ? '登陆' : '注册' }}</Button>
          <a @click="isLogin = !isLogin">{{ isLogin ? '没账号？去注册' : '有账号？去登陆' }}</a>
        </div>
      </FormItem>
    </Form>
    <Modal :open="previewVisible" :title="previewTitle" :footer="null" @cancel="handleCancel">
      <img alt="头像" style="width: 100%" :src="previewImage"/>
    </Modal>
  </div>
</template>

<style scoped>
.csms-body-loginandregister {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}

.csms-body-loginandregister-title {
  font-size: 24px;
  font-style: oblique;
  font-weight: 550;
  margin-top: 10px;
}

.csms-body-loginandregister-form {
  margin-top: 20px;
  width: 680px;
}

.csms-body-loginandregister-form-button-inner {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title-enter-active, .title-leave-active {
  transition: opacity .15s linear, transform .15s linear;
}

.title-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}

.title-leave-to {
  opacity: 0;
  transform: translateY(10px);
}
</style>
