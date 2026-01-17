<script setup>
import { ref } from 'vue'

const userInfo = ref({
  name: '王先生',
  avatar: '',
  company: '杭州优质纺织有限公司',
  phone: '138****8888',
  email: 'wang@example.com',
  location: '浙江杭州',
  joinTime: '2025-06-15',
  role: 'supplier', // supplier 供应商 / buyer 采购商
  verified: true
})

const stats = ref({
  totalBounties: 12,
  totalBids: 28,
  successRate: 85,
  completedOrders: 24
})
</script>

<template>
  <div class="w-[80%] mx-auto py-6">
    <div class="flex gap-8">
      <!-- 左侧个人信息 -->
      <div class="w-80 shrink-0">
        <div class="bg-white rounded-lg shadow-sm overflow-hidden">
          <!-- 头像区域 -->
          <div class="bg-gradient-to-r from-blue-500 to-cyan-500 p-8 text-center">
            <div class="w-24 h-24 mx-auto rounded-full bg-white text-blue-500 flex items-center justify-center text-4xl font-bold mb-4">
              {{ userInfo.name.charAt(0) }}
            </div>
            <h2 class="text-xl font-bold text-white mb-1">{{ userInfo.name }}</h2>
            <p class="text-sm text-white/80">{{ userInfo.company }}</p>
            <div v-if="userInfo.verified" class="inline-flex items-center gap-1 mt-3 bg-white/20 text-white text-xs px-3 py-1 rounded-full">
              <i class="fas fa-check-circle"></i>已认证
            </div>
          </div>

          <!-- 基本信息 -->
          <div class="p-6">
            <div class="flex flex-col gap-4 text-sm">
              <div class="flex items-center gap-3 text-gray-600">
                <i class="fas fa-phone w-5 text-gray-400"></i>
                {{ userInfo.phone }}
              </div>
              <div class="flex items-center gap-3 text-gray-600">
                <i class="fas fa-envelope w-5 text-gray-400"></i>
                {{ userInfo.email }}
              </div>
              <div class="flex items-center gap-3 text-gray-600">
                <i class="fas fa-map-marker-alt w-5 text-gray-400"></i>
                {{ userInfo.location }}
              </div>
              <div class="flex items-center gap-3 text-gray-600">
                <i class="fas fa-calendar w-5 text-gray-400"></i>
                加入时间: {{ userInfo.joinTime }}
              </div>
            </div>

            <button class="w-full mt-6 py-2.5 border border-blue-500 text-blue-500 rounded text-sm font-medium hover:bg-blue-50 transition-colors">
              <i class="fas fa-edit mr-2"></i>编辑资料
            </button>
          </div>
        </div>
      </div>

      <!-- 右侧内容区 -->
      <div class="flex-1 min-w-0">
        <!-- 数据统计 -->
        <div class="grid grid-cols-4 gap-4 mb-6">
          <div class="bg-white rounded-lg p-5 shadow-sm text-center">
            <div class="text-3xl font-bold text-blue-500 mb-1">{{ stats.totalBounties }}</div>
            <div class="text-sm text-gray-500">发布悬赏</div>
          </div>
          <div class="bg-white rounded-lg p-5 shadow-sm text-center">
            <div class="text-3xl font-bold text-green-500 mb-1">{{ stats.totalBids }}</div>
            <div class="text-sm text-gray-500">投标次数</div>
          </div>
          <div class="bg-white rounded-lg p-5 shadow-sm text-center">
            <div class="text-3xl font-bold text-orange-500 mb-1">{{ stats.successRate }}%</div>
            <div class="text-sm text-gray-500">中标率</div>
          </div>
          <div class="bg-white rounded-lg p-5 shadow-sm text-center">
            <div class="text-3xl font-bold text-purple-500 mb-1">{{ stats.completedOrders }}</div>
            <div class="text-sm text-gray-500">完成订单</div>
          </div>
        </div>

        <!-- 功能菜单 -->
        <div class="bg-white rounded-lg shadow-sm p-6">
          <h3 class="text-lg font-bold text-gray-800 mb-5">账户设置</h3>
          <div class="grid grid-cols-2 gap-4">
            <button class="flex items-center gap-4 p-4 rounded-lg border border-gray-100 hover:border-blue-200 hover:bg-blue-50 transition-colors text-left">
              <div class="w-10 h-10 rounded-lg bg-blue-100 text-blue-500 flex items-center justify-center">
                <i class="fas fa-user-edit"></i>
              </div>
              <div>
                <div class="font-medium text-gray-800">个人信息</div>
                <div class="text-xs text-gray-500">修改头像、昵称等</div>
              </div>
            </button>

            <button class="flex items-center gap-4 p-4 rounded-lg border border-gray-100 hover:border-blue-200 hover:bg-blue-50 transition-colors text-left">
              <div class="w-10 h-10 rounded-lg bg-green-100 text-green-500 flex items-center justify-center">
                <i class="fas fa-building"></i>
              </div>
              <div>
                <div class="font-medium text-gray-800">企业认证</div>
                <div class="text-xs text-gray-500">提交企业资质认证</div>
              </div>
            </button>

            <button class="flex items-center gap-4 p-4 rounded-lg border border-gray-100 hover:border-blue-200 hover:bg-blue-50 transition-colors text-left">
              <div class="w-10 h-10 rounded-lg bg-orange-100 text-orange-500 flex items-center justify-center">
                <i class="fas fa-wallet"></i>
              </div>
              <div>
                <div class="font-medium text-gray-800">账户余额</div>
                <div class="text-xs text-gray-500">充值、提现、账单</div>
              </div>
            </button>

            <button class="flex items-center gap-4 p-4 rounded-lg border border-gray-100 hover:border-blue-200 hover:bg-blue-50 transition-colors text-left">
              <div class="w-10 h-10 rounded-lg bg-purple-100 text-purple-500 flex items-center justify-center">
                <i class="fas fa-lock"></i>
              </div>
              <div>
                <div class="font-medium text-gray-800">安全设置</div>
                <div class="text-xs text-gray-500">修改密码、绑定手机</div>
              </div>
            </button>

            <button class="flex items-center gap-4 p-4 rounded-lg border border-gray-100 hover:border-blue-200 hover:bg-blue-50 transition-colors text-left">
              <div class="w-10 h-10 rounded-lg bg-cyan-100 text-cyan-500 flex items-center justify-center">
                <i class="fas fa-bell"></i>
              </div>
              <div>
                <div class="font-medium text-gray-800">消息通知</div>
                <div class="text-xs text-gray-500">设置通知偏好</div>
              </div>
            </button>

            <button class="flex items-center gap-4 p-4 rounded-lg border border-gray-100 hover:border-red-200 hover:bg-red-50 transition-colors text-left">
              <div class="w-10 h-10 rounded-lg bg-red-100 text-red-500 flex items-center justify-center">
                <i class="fas fa-sign-out-alt"></i>
              </div>
              <div>
                <div class="font-medium text-gray-800">退出登录</div>
                <div class="text-xs text-gray-500">安全退出当前账号</div>
              </div>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
