<script setup>
import { ref } from 'vue'

// 接单记录（我投标的任务）
const myBids = ref([])

const formatMoney = (amount) => {
  return new Intl.NumberFormat('zh-CN').format(amount)
}

const getStatusText = (status) => {
  const statusMap = {
    'pending': '待审核',
    'accepted': '已中标',
    'rejected': '未中标',
    'completed': '已完成'
  }
  return statusMap[status] || status
}
</script>

<template>
  <div class="w-[80%] mx-auto">
    <div class="flex justify-between items-center mb-6">
      <h2 class="text-xl font-bold text-gray-800">接单记录</h2>
      <div class="flex gap-2">
        <button class="text-sm py-2 px-4 rounded border border-gray-200 bg-white text-gray-600 hover:border-blue-500 hover:text-blue-500 transition-colors">
          全部
        </button>
        <button class="text-sm py-2 px-4 rounded border border-gray-200 bg-white text-gray-600 hover:border-blue-500 hover:text-blue-500 transition-colors">
          待审核
        </button>
        <button class="text-sm py-2 px-4 rounded border border-gray-200 bg-white text-gray-600 hover:border-blue-500 hover:text-blue-500 transition-colors">
          已中标
        </button>
        <button class="text-sm py-2 px-4 rounded border border-gray-200 bg-white text-gray-600 hover:border-blue-500 hover:text-blue-500 transition-colors">
          已完成
        </button>
      </div>
    </div>

    <!-- 有数据时显示列表 -->
    <div v-if="myBids.length > 0" class="flex flex-col gap-5">
      <div
        v-for="bid in myBids"
        :key="bid.id"
        class="bg-white rounded-lg p-6 shadow-sm hover:shadow-md transition-shadow"
      >
        <div class="flex justify-between items-start mb-4">
          <div class="flex-1 min-w-0 mr-4">
            <h3 class="text-lg font-semibold text-gray-800 mb-2">{{ bid.bountyTitle }}</h3>
            <div class="flex flex-wrap gap-x-5 gap-y-1 text-xs text-gray-500">
              <span class="flex items-center gap-1">
                <i class="far fa-user"></i>{{ bid.publisher }}
              </span>
              <span class="flex items-center gap-1">
                <i class="far fa-clock"></i>投标时间: {{ bid.bidTime }}
              </span>
            </div>
          </div>
          <div class="flex items-center gap-3">
            <span
              class="px-3 py-1 rounded-full text-xs font-medium"
              :class="{
                'bg-yellow-100 text-yellow-600': bid.status === 'pending',
                'bg-green-100 text-green-600': bid.status === 'accepted',
                'bg-red-100 text-red-600': bid.status === 'rejected',
                'bg-blue-100 text-blue-600': bid.status === 'completed'
              }"
            >
              {{ getStatusText(bid.status) }}
            </span>
            <div class="bg-orange-50 text-orange-500 py-2 px-4 rounded font-bold text-lg whitespace-nowrap">
              ¥{{ formatMoney(bid.bidAmount) }}
            </div>
          </div>
        </div>

        <div class="bg-gray-50 rounded p-4 mb-4">
          <h4 class="text-sm font-medium text-gray-700 mb-2">我的报价说明</h4>
          <p class="text-sm text-gray-600 leading-relaxed">{{ bid.bidDescription }}</p>
        </div>

        <div class="flex justify-between items-center">
          <div class="flex gap-5 text-xs text-gray-500">
            <span class="flex items-center gap-1">
              <i class="fas fa-coins"></i>悬赏金额: ¥{{ formatMoney(bid.rewardAmount) }}
            </span>
            <span class="flex items-center gap-1 text-red-500 font-medium">
              <i class="far fa-calendar-times"></i>截止: {{ bid.deadline }}
            </span>
          </div>
          <button class="text-blue-500 hover:text-blue-600 text-sm font-medium py-2 px-4 border border-blue-500 rounded transition-colors">
            查看详情
          </button>
        </div>
      </div>
    </div>

    <!-- 无数据时显示空状态 -->
    <div v-else class="bg-white rounded-lg p-16 shadow-sm text-center">
      <i class="fas fa-handshake text-6xl text-gray-300 mb-6"></i>
      <h3 class="text-lg font-medium text-gray-600 mb-2">暂无接单记录</h3>
      <p class="text-sm text-gray-400 mb-6">去悬赏大厅查看需求，投标接单赚取收益</p>
      <router-link to="/" class="inline-block bg-blue-500 hover:bg-blue-600 text-white text-sm font-medium py-2.5 px-6 rounded transition-colors">
        <i class="fas fa-search mr-2"></i>浏览悬赏
      </router-link>
    </div>
  </div>
</template>
