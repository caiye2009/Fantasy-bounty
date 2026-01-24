<script setup>
import { ref, computed, provide, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BountyHall from './BountyHall.vue'
import MyBids from './MyBids.vue'
import MyProfile from './MyProfile.vue'
import { logout, isAuthenticated, sendVerifyCode, loginWithCode, getPhone, getUsername } from '@/api/auth'
import { Shirt, Megaphone } from 'lucide-vue-next'

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

// 用户手机号和用户名（用于显示）
const userPhone = ref('')
const userName = ref('')

// 用户菜单容器ref
const userMenuRef = ref(null)

// Banner 轮播状态
const currentSlide = ref(0)
const bannerSlides = [
  {
    title: '纺织布匹悬赏接单平台',
    subtitle: '发布您的面料需求，让优质供应商竞标接单 · 快速匹配，质量保障，交易安全',
    gradient: 'from-blue-900 via-blue-700 to-cyan-500'
  },
  {
    title: '优质供应商精准匹配',
    subtitle: '海量面料供应商资源，智能匹配您的需求 · 品质保障，价格透明',
    gradient: 'from-indigo-900 via-purple-700 to-pink-500'
  },
  {
    title: '安全交易全程保障',
    subtitle: '资金担保，质检验收，售后无忧 · 让每一笔交易都放心可靠',
    gradient: 'from-emerald-900 via-teal-700 to-cyan-400'
  }
]
let bannerTimer = null

const startBannerTimer = () => {
  bannerTimer = setInterval(() => {
    currentSlide.value = (currentSlide.value + 1) % bannerSlides.length
  }, 4000)
}

const goToSlide = (index) => {
  currentSlide.value = index
  // 重置定时器
  if (bannerTimer) clearInterval(bannerTimer)
  startBannerTimer()
}

// Banner 拖拽滑动
const isDragging = ref(false)
let dragStartX = 0

const onDragStart = (e) => {
  isDragging.value = true
  dragStartX = e.type === 'touchstart' ? e.touches[0].clientX : e.clientX
}

const onDragEnd = (e) => {
  if (!isDragging.value) return
  isDragging.value = false
  const endX = e.type === 'touchend' ? e.changedTouches[0].clientX : e.clientX
  const diff = endX - dragStartX
  if (Math.abs(diff) > 50) {
    if (diff < 0) {
      goToSlide((currentSlide.value + 1) % bannerSlides.length)
    } else {
      goToSlide((currentSlide.value - 1 + bannerSlides.length) % bannerSlides.length)
    }
  }
}

// 导航栏滑块
const navRef = ref(null)
const sliderStyle = ref({})

const updateSlider = () => {
  if (!navRef.value) return
  const activeIndex = navItems.findIndex(item => item.key === currentPage.value)
  if (activeIndex < 0) return
  const buttons = navRef.value.querySelectorAll('button')
  if (!buttons[activeIndex]) return
  const btn = buttons[activeIndex]
  const navRect = navRef.value.getBoundingClientRect()
  const btnRect = btn.getBoundingClientRect()
  sliderStyle.value = {
    width: `${btnRect.width}px`,
    transform: `translateX(${btnRect.left - navRect.left}px)`
  }
}

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
    userName.value = getUsername() || ''
  }
  startBannerTimer()
  nextTick(() => updateSlider())
})

// 组件卸载时清理定时器和事件监听
onUnmounted(() => {
  if (countdownTimer) {
    clearInterval(countdownTimer)
  }
  if (bannerTimer) {
    clearInterval(bannerTimer)
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

watch(currentPage, () => {
  nextTick(() => updateSlider())
})

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
    userName.value = getUsername() || ''
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
  userName.value = ''
  userPhone.value = ''
  router.push({ query: { tab: 'hall' } })
}
</script>

<template>
  <div class="min-h-screen bg-gray-100 flex flex-col gap-2">
    <header class="bg-white shadow-md sticky top-0 z-50 w-full">
      <div class="h-[70px] flex justify-between items-center px-[5%]">
        <!-- Logo -->
        <div class="flex items-center gap-3 cursor-pointer" @click="switchPage('hall')">
          <Shirt class="text-blue-500" :size="24" />
          <h1 class="text-xl font-bold text-gray-800">
            纺织<span class="text-red-500">悬赏</span>大厅
          </h1>
        </div>

        <!-- Nav -->
        <nav ref="navRef" class="relative">
          <ul class="flex gap-10">
            <li v-for="item in navItems" :key="item.key">
              <button
                @click="switchPage(item.key)"
                class="font-medium px-2 py-1 transition-colors cursor-pointer"
                :class="currentPage === item.key
                  ? 'text-blue-500'
                  : 'text-gray-600 hover:text-blue-500'"
              >
                {{ item.name }}
              </button>
            </li>
          </ul>
          <!-- 滑动指示条 -->
          <span
            v-show="currentPage === 'hall' || currentPage === 'myBids'"
            class="absolute bottom-0 h-0.5 bg-blue-500 transition-all duration-300 ease-in-out"
            :style="sliderStyle"
          ></span>
        </nav>

        <!-- User -->
        <div class="relative" ref="userMenuRef">
          <template v-if="isLoggedIn">
            <div
              @click="goToProfile()"
              class="w-9 h-9 rounded-full bg-blue-500 text-white flex items-center justify-center font-bold cursor-pointer text-sm hover:bg-blue-600 transition-colors"
              :title="userName || userPhone || '个人中心'"
            >
              {{ userName ? userName.slice(0, 2) : '用户' }}
            </div>
          </template>
          <el-button
            v-else
            type="primary"
            @click="openLoginModal"
          >
            登录
          </el-button>
        </div>
      </div>
    </header>




    <!-- Banner 轮播 -->
    <div
      class="mx-16 my-8 h-[300px] rounded-2xl relative overflow-hidden select-none"
      :class="isDragging ? 'cursor-grabbing' : 'cursor-grab'"
      @mousedown="onDragStart"
      @mouseup="onDragEnd"
      @mouseleave="onDragEnd"
      @touchstart="onDragStart"
      @touchend="onDragEnd"
    >
      <div
        v-for="(slide, index) in bannerSlides"
        :key="index"
        class="absolute inset-0 bg-gradient-to-r text-white flex items-center justify-center transition-opacity duration-700 ease-in-out"
        :class="[slide.gradient, currentSlide === index ? 'opacity-100 z-10' : 'opacity-0 z-0']"
      >
        <div class="text-center">
          <h2 class="text-3xl font-bold mb-4">{{ slide.title }}</h2>
          <p class="text-base opacity-90 mb-8">{{ slide.subtitle }}</p>
          <el-button type="danger" size="large" class="!py-5 !px-8">
            <Megaphone class="mr-2" :size="16" />发布悬赏需求
          </el-button>
        </div>
      </div>
      <!-- 轮播圆点指示器 -->
      <div class="absolute bottom-5 left-1/2 -translate-x-1/2 flex gap-2 z-20">
        <button
          v-for="(_, index) in bannerSlides"
          :key="index"
          @click="goToSlide(index)"
          class="w-2.5 h-2.5 rounded-full transition-all duration-300 cursor-pointer"
          :class="currentSlide === index ? 'bg-white scale-110' : 'bg-white/50 hover:bg-white/75'"
        ></button>
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
        <el-alert v-if="loginError" :title="loginError" type="error" show-icon :closable="false" class="!mb-4" />

        <!-- 手机号输入 -->
        <el-input
          v-model="loginForm.phone"
          maxlength="11"
          placeholder="请输入手机号"
          class="!mb-4"
          size="large"
          @keyup.enter="handleSubmit"
        />

        <!-- 验证码输入 + 发送按钮 -->
        <div class="flex gap-3 mb-4">
          <el-input
            v-model="loginForm.code"
            maxlength="6"
            placeholder="请输入验证码"
            class="flex-1"
            size="large"
            @keyup.enter="handleSubmit"
          />
          <el-button
            @click="handleSendCode"
            :disabled="countdown > 0 || codeSending"
            size="large"
            class="!w-28"
          >
            {{ codeSending ? '发送中...' : (countdown > 0 ? `${countdown}s后重发` : '获取验证码') }}
          </el-button>
        </div>

        <el-button
          type="primary"
          @click="handleSubmit"
          :disabled="loginLoading"
          :loading="loginLoading"
          size="large"
          class="!w-full !mt-2"
        >
          {{ loginLoading ? '登录中...' : '登录 / 注册' }}
        </el-button>

        <p class="text-center text-xs text-gray-400 mt-5">
          未注册的手机号将自动创建账号
        </p>
      </div>
    </div>
  </div>
</template>