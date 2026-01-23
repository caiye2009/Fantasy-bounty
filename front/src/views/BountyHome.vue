<script setup>
import { ref, computed, provide, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BountyHall from './BountyHall.vue'
import MyBids from './MyBids.vue'
import MyProfile from './MyProfile.vue'
import { logout, isAuthenticated, sendVerifyCode, loginWithCode, getPhone } from '@/api/auth'

const route = useRoute()
const router = useRouter()

// 当前页面
const currentPage = computed(() => route.query.tab || 'hall')
const isLoggedIn = ref(false)
const showLoginModal = ref(false)
const showUserMenu = ref(false)

// 登录表单状态
const loginForm = ref({
  phone: '',
  code: ''
})
const loginError = ref('')
const loginLoading = ref(false)
const codeSending = ref(false)
const countdown = ref(0)
let countdownTimer = null

// 用户手机号（用于显示）
const userPhone = ref('')

// 用户菜单容器ref
const userMenuRef = ref(null)

// 处理点击外部关闭菜单
const handleClickOutside = (e) => {
  if (userMenuRef.value && !userMenuRef.value.contains(e.target)) {
    showUserMenu.value = false
  }
}

// 处理ESC键关闭菜单
const handleEscKey = (e) => {
  if (e.key === 'Escape') {
    showUserMenu.value = false
  }
}

// 监听菜单状态，打开时添加事件监听，关闭时移除
watch(showUserMenu, (isOpen) => {
  if (isOpen) {
    // 使用 nextTick 确保在当前点击事件完成后再添加监听
    nextTick(() => {
      document.addEventListener('click', handleClickOutside)
      document.addEventListener('keydown', handleEscKey)
    })
  } else {
    document.removeEventListener('click', handleClickOutside)
    document.removeEventListener('keydown', handleEscKey)
  }
})

// 页面加载时检测 token
onMounted(() => {
  isLoggedIn.value = isAuthenticated()
  if (isLoggedIn.value) {
    userPhone.value = getPhone() || ''
  }
})

// 组件卸载时清理定时器和事件监听
onUnmounted(() => {
  if (countdownTimer) {
    clearInterval(countdownTimer)
  }
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleEscKey)
})

// 统一的打开登录modal入口
const openLoginModal = () => {
  loginError.value = ''
  loginForm.value = { phone: '', code: '' }
  showLoginModal.value = true
}

// 验证手机号格式
const isValidPhone = (phone) => {
  return /^1[3-9]\d{9}$/.test(phone)
}

// 发送验证码
const handleSendCode = async () => {
  if (!loginForm.value.phone) {
    loginError.value = '请输入手机号'
    return
  }

  if (!isValidPhone(loginForm.value.phone)) {
    loginError.value = '请输入正确的手机号'
    return
  }

  codeSending.value = true
  loginError.value = ''

  try {
    await sendVerifyCode(loginForm.value.phone)
    // 开始60秒倒计时
    countdown.value = 60
    countdownTimer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) {
        clearInterval(countdownTimer)
        countdownTimer = null
      }
    }, 1000)
  } catch (error) {
    loginError.value = error.message || '验证码发送失败'
  } finally {
    codeSending.value = false
  }
}

// 提供给子组件使用
provide('isLoggedIn', isLoggedIn)
provide('openLoginModal', openLoginModal)

// 导航菜单（个人中心通过头像进入，不在导航栏显示）
const navItems = [
  { key: 'hall', name: '悬赏大厅' },
  { key: 'myBids', name: '我的投标' }
]

const switchPage = (key) => {
  // 需要登录的页面
  if (key === 'myBids' && !isLoggedIn.value) {
    openLoginModal()
    return
  }
  router.push({ query: { tab: key } })
}

const goToProfile = () => {
  if (isLoggedIn.value) {
    switchPage('profile')
  } else {
    openLoginModal()
  }
}

// 提交登录（手机号+验证码）
const handleSubmit = async () => {
  if (!loginForm.value.phone) {
    loginError.value = '请输入手机号'
    return
  }

  if (!isValidPhone(loginForm.value.phone)) {
    loginError.value = '请输入正确的手机号'
    return
  }

  if (!loginForm.value.code) {
    loginError.value = '请输入验证码'
    return
  }

  loginLoading.value = true
  loginError.value = ''

  try {
    await loginWithCode(loginForm.value.phone, loginForm.value.code)
    isLoggedIn.value = true
    userPhone.value = loginForm.value.phone
    showLoginModal.value = false
    loginForm.value = { phone: '', code: '' }
    // 清理倒计时
    if (countdownTimer) {
      clearInterval(countdownTimer)
      countdownTimer = null
      countdown.value = 0
    }
  } catch (error) {
    loginError.value = error.message || '登录失败，请重试'
  } finally {
    loginLoading.value = false
  }
}

// 登出
const handleLogout = () => {
  logout()
  isLoggedIn.value = false
  router.push({ query: { tab: 'hall' } })
}
</script>

<template>
  <div class="min-h-screen bg-gray-100 flex flex-col gap-2">
    <header class="bg-white shadow-md sticky top-0 z-50 w-full">
      <div class="h-[70px] flex justify-between items-center px-[5%]">
        <!-- Logo -->
        <div class="flex items-center gap-3 cursor-pointer" @click="switchPage('hall')">
          <i class="fas fa-vest text-blue-500 text-2xl"></i>
          <h1 class="text-xl font-bold text-gray-800">
            纺织<span class="text-red-500">悬赏</span>大厅
          </h1>
        </div>

        <!-- Nav -->
        <nav>
          <ul class="flex gap-10">
            <li v-for="item in navItems" :key="item.key">
              <button
                @click="switchPage(item.key)"
                class="font-medium px-2 py-1 transition-colors relative"
                :class="currentPage === item.key
                  ? 'text-blue-500'
                  : 'text-gray-600 hover:text-blue-500'"
              >
                {{ item.name }}
                <span
                  v-if="currentPage === item.key"
                  class="absolute bottom-0 left-0 w-full h-0.5 bg-blue-500"
                ></span>
              </button>
            </li>
          </ul>
        </nav>

        <!-- User -->
        <div class="relative" ref="userMenuRef">
          <template v-if="isLoggedIn">
            <div
              @click.stop="showUserMenu = !showUserMenu"
              class="w-9 h-9 rounded-full bg-blue-500 text-white flex items-center justify-center font-bold cursor-pointer text-sm hover:bg-blue-600 transition-colors"
              :title="userPhone ? userPhone : '个人中心'"
            >
              {{ userPhone ? userPhone.slice(-2) : 'U' }}
            </div>
            <!-- 用户菜单 -->
            <div
              v-if="showUserMenu"
              class="absolute right-0 top-12 bg-white rounded-lg shadow-lg py-2 w-36 z-50"
            >
              <button
                @click="goToProfile(); showUserMenu = false"
                class="w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
              >
                <i class="fas fa-user mr-2"></i>个人中心
              </button>
              <button
                @click="handleLogout(); showUserMenu = false"
                class="w-full px-4 py-2 text-left text-sm text-red-500 hover:bg-gray-100"
              >
                <i class="fas fa-sign-out-alt mr-2"></i>退出登录
              </button>
            </div>
          </template>
          <button
            v-else
            @click="openLoginModal"
            class="bg-blue-500 hover:bg-blue-600 text-white text-sm font-medium py-2 px-5 rounded-md transition-colors"
          >
            登录
          </button>
        </div>
      </div>
    </header>




    <!-- Banner -->
    <div class="mx-16 my-8 h-[300px] rounded-2xl bg-gradient-to-r from-blue-900 via-blue-700 to-cyan-500 text-white flex items-center justify-center">
      <div class="text-center">
        <h2 class="text-3xl font-bold mb-4">纺织布匹悬赏接单平台</h2>
        <p class="text-base opacity-90 mb-8">
          发布您的面料需求，让优质供应商竞标接单 · 快速匹配，质量保障，交易安全
        </p>
        <button class="bg-red-500 hover:bg-red-600 text-white font-medium py-3 px-8 rounded transition-colors">
          <i class="fas fa-bullhorn mr-2"></i>发布悬赏需求
        </button>
      </div>
    </div>

    <!-- Main Content -->
    <main class="flex-1 w-full flex justify-center">
      <!-- 悬赏大厅 -->
      <BountyHall v-if="currentPage === 'hall'" />

      <!-- 接单记录 -->
      <MyBids v-else-if="currentPage === 'myBids'" />

      <!-- 个人中心 -->
      <MyProfile v-else-if="currentPage === 'profile'" />

    </main>

    <!-- Login Modal (手机号+验证码) -->
    <div v-if="showLoginModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showLoginModal = false">
      <div class="bg-white rounded-lg p-8 w-96 shadow-xl">
        <h3 class="text-xl font-bold text-gray-800 mb-6 text-center">
          登录 / 注册
        </h3>

        <!-- 错误提示 -->
        <div v-if="loginError" class="mb-4 p-3 bg-red-50 border border-red-200 rounded text-red-600 text-sm">
          {{ loginError }}
        </div>

        <!-- 手机号输入 -->
        <input
          v-model="loginForm.phone"
          type="tel"
          maxlength="11"
          placeholder="请输入手机号"
          class="w-full mb-4 px-4 py-3 border border-gray-200 rounded text-sm focus:outline-none focus:border-blue-500"
          @keyup.enter="handleSubmit"
        >

        <!-- 验证码输入 + 发送按钮 -->
        <div class="flex gap-3 mb-4">
          <input
            v-model="loginForm.code"
            type="text"
            maxlength="6"
            placeholder="请输入验证码"
            class="flex-1 px-4 py-3 border border-gray-200 rounded text-sm focus:outline-none focus:border-blue-500"
            @keyup.enter="handleSubmit"
          >
          <button
            @click="handleSendCode"
            :disabled="countdown > 0 || codeSending"
            class="w-28 px-3 py-3 text-sm font-medium rounded transition-colors"
            :class="countdown > 0 || codeSending
              ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
              : 'bg-blue-50 text-blue-500 hover:bg-blue-100'"
          >
            {{ codeSending ? '发送中...' : (countdown > 0 ? `${countdown}s后重发` : '获取验证码') }}
          </button>
        </div>

        <button
          @click="handleSubmit"
          :disabled="loginLoading"
          class="w-full bg-blue-500 hover:bg-blue-600 text-white font-medium py-3 rounded transition-colors disabled:bg-blue-300 mt-2"
        >
          {{ loginLoading ? '登录中...' : '登录 / 注册' }}
        </button>

        <p class="text-center text-xs text-gray-400 mt-5">
          未注册的手机号将自动创建账号
        </p>
      </div>
    </div>
  </div>
</template>