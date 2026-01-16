<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BountyHall from './BountyHall.vue'
import MyBids from './MyBids.vue'
import MyProfile from './MyProfile.vue'

const route = useRoute()
const router = useRouter()

// 当前页面
const currentPage = computed(() => route.query.tab || 'hall')
const isLoggedIn = ref(false)
const showLoginModal = ref(false)

// 导航菜单（个人中心通过头像进入，不在导航栏显示）
const navItems = [
  { key: 'hall', name: '悬赏大厅' },
  { key: 'myBids', name: '接单记录' }
]

const switchPage = (key) => {
  router.push({ query: { tab: key } })
}

const goToProfile = () => {
  if (isLoggedIn.value) {
    switchPage('profile')
  } else {
    showLoginModal.value = true
  }
}

const handleLogin = () => {
  isLoggedIn.value = true
  showLoginModal.value = false
}
</script>

<template>
  <div class="min-h-screen bg-gray-100 flex flex-col gap-6">
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
        <div>
          <div
            v-if="isLoggedIn"
            @click="goToProfile"
            class="w-9 h-9 rounded-full bg-blue-500 text-white flex items-center justify-center font-bold cursor-pointer text-sm hover:bg-blue-600 transition-colors"
            title="个人中心"
          >
            王
          </div>
          <button
            v-else
            @click="showLoginModal = true"
            class="bg-blue-500 hover:bg-blue-600 text-white text-sm font-medium py-2 px-5 rounded transition-colors"
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
    <main class="flex-1 w-full py-10 flex justify-center">
      <!-- 悬赏大厅 -->
      <BountyHall v-if="currentPage === 'hall'" />

      <!-- 接单记录 -->
      <MyBids v-else-if="currentPage === 'myBids'" />

      <!-- 个人中心 -->
      <MyProfile v-else-if="currentPage === 'profile'" />
    </main>

    <!-- Login Modal -->
    <div v-if="showLoginModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showLoginModal = false">
      <div class="bg-white rounded-lg p-8 w-96 shadow-xl">
        <h3 class="text-xl font-bold text-gray-800 mb-6 text-center">登录</h3>
        <input type="text" placeholder="用户名" class="w-full mb-4 px-4 py-3 border border-gray-200 rounded text-sm focus:outline-none focus:border-blue-500">
        <input type="password" placeholder="密码" class="w-full mb-6 px-4 py-3 border border-gray-200 rounded text-sm focus:outline-none focus:border-blue-500">
        <button @click="handleLogin" class="w-full bg-blue-500 hover:bg-blue-600 text-white font-medium py-3 rounded transition-colors">
          登录
        </button>
        <p class="text-center text-sm text-gray-500 mt-5">
          没有账号？<button class="text-blue-500 hover:underline">注册</button>
        </p>
      </div>
    </div>
  </div>
</template>