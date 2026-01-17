<script setup>
import { ref, onMounted, watch } from 'vue'
import { fetchMyBids } from '@/api/bid'

// 筛选条件
const statusOptions = [
  { label: '全部', value: '' },
  { label: '待审核', value: 'pending' },
  { label: '已中标', value: 'accepted' },
  { label: '未中标', value: 'rejected' },
  { label: '已完成', value: 'completed' }
]
const selectedStatus = ref('')

// 接单记录（我投标的任务）
const myBids = ref([])
const loading = ref(false)

// 悬赏类型映射
const bountyTypeMap = {
  woven: '梭织',
  knitted: '针织'
}

// 加载我的投标列表
const loadMyBids = async () => {
  loading.value = true
  try {
    const result = await fetchMyBids(selectedStatus.value, 1, 50)
    myBids.value = result.data.map(item => ({
      id: item.id,
      bountyId: item.bountyId,
      bountyTitle: item.bountyProductName || '未知悬赏',
      bountyType: item.bountyType,
      bidTime: formatDateTime(item.createdAt),
      status: item.status,
      bidAmount: item.bidPrice,
      deadline: formatDate(item.bidDeadline),
      wovenSpec: item.wovenSpec,
      knittedSpec: item.knittedSpec
    }))
  } catch (error) {
    console.error('加载投标列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 格式化日期时间
const formatDateTime = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  }).replace(/\//g, '-')
}

// 格式化日期
const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  }).replace(/\//g, '-')
}

// 监听筛选条件变化
watch(selectedStatus, () => {
  loadMyBids()
})

onMounted(() => {
  loadMyBids()
})

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

// 抽屉状态
const drawerVisible = ref(false)
const selectedBid = ref(null)

const openDrawer = (bid) => {
  selectedBid.value = bid
  drawerVisible.value = true
}

const closeDrawer = () => {
  drawerVisible.value = false
  selectedBid.value = null
}
</script>

<template>
  <div class="w-[80%] mx-auto py-6">
    <!-- 左右布局容器 -->
    <div class="flex gap-6">
      <!-- 左侧：筛选区域 -->
      <aside class="w-[30%] shrink-0 sticky top-24 h-[calc(100vh-7rem)] flex flex-col justify-between">
        <!-- 筛选框 -->
        <div class="bg-white rounded-lg p-8 shadow-sm">
          <h3 class="text-lg font-semibold text-gray-800 mb-6">筛选条件</h3>

          <!-- 状态筛选 -->
          <div class="flex flex-col gap-5">
            <div class="flex items-center justify-between gap-3">
              <span class="text-sm text-gray-600 whitespace-nowrap">投标状态</span>
              <select
                v-model="selectedStatus"
                class="px-3 py-1.5 border border-gray-200 rounded text-sm text-gray-600 focus:outline-none focus:border-blue-500 flex-1"
              >
                <option v-for="opt in statusOptions" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </option>
              </select>
            </div>
          </div>
        </div>

        <!-- 页脚信息 -->
        <div class="p-6 text-sm">
          <h3 class="font-semibold text-gray-800 mb-2">我的投标</h3>
          <p class="text-gray-500 text-xs mb-4 leading-relaxed">查看和管理您参与的所有投标记录，跟踪投标状态和结果。</p>

          <div class="border-t border-gray-100 pt-3">
            <h4 class="text-gray-700 font-medium mb-2">投标须知</h4>
            <ul class="text-gray-500 text-xs space-y-1">
              <li>认真阅读需求详情</li>
              <li>合理报价，诚信投标</li>
              <li>中标后按时交付</li>
            </ul>
          </div>
        </div>
      </aside>

      <!-- 右侧：投标列表 -->
      <div class="flex-1 min-w-0 flex flex-col gap-5">
        <!-- 加载状态 -->
        <div v-if="loading" class="bg-white rounded-lg p-16 shadow-sm text-center">
          <i class="fas fa-spinner fa-spin text-4xl text-gray-300 mb-4"></i>
          <p class="text-gray-500">加载中...</p>
        </div>

        <!-- 有数据时显示列表 -->
        <template v-else-if="myBids.length > 0">
          <div
            v-for="bid in myBids"
            :key="bid.id"
            class="bg-white rounded-lg p-6 shadow-sm hover:shadow-md transition-shadow cursor-pointer"
            @click="openDrawer(bid)"
          >
            <div class="flex justify-between items-start mb-4">
              <div class="flex-1 min-w-0 mr-4">
                <h3 class="text-lg font-semibold text-gray-800 mb-2">{{ bid.bountyTitle }}</h3>
                <div class="flex flex-wrap gap-x-5 gap-y-1 text-xs text-gray-500">
                  <span class="flex items-center gap-1">
                    <i class="fas fa-tag"></i>{{ bountyTypeMap[bid.bountyType] || bid.bountyType }}
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

            <div class="flex justify-between items-center">
              <div class="flex gap-5 text-xs text-gray-500">
                <span class="flex items-center gap-1 text-red-500 font-medium">
                  <i class="far fa-calendar-times"></i>投标截止: {{ bid.deadline }}
                </span>
              </div>
              <button
                class="text-blue-500 hover:text-blue-600 text-sm font-medium py-2 px-4 border border-blue-500 rounded transition-colors"
                @click.stop="openDrawer(bid)"
              >
                查看详情
              </button>
            </div>
          </div>
        </template>

        <!-- 无数据时显示空状态 -->
        <div v-else class="bg-white rounded-lg p-16 shadow-sm text-center">
          <i class="fas fa-handshake text-6xl text-gray-300 mb-6"></i>
          <h3 class="text-lg font-medium text-gray-600 mb-2">暂无投标记录</h3>
          <p class="text-sm text-gray-400 mb-6">去悬赏大厅查看需求</p>
          <router-link to="/" class="inline-block bg-blue-500 hover:bg-blue-600 text-white text-sm font-medium py-2.5 px-6 rounded transition-colors">
            <i class="fas fa-search mr-2"></i>浏览悬赏
          </router-link>
        </div>
      </div>
    </div>

    <!-- 抽屉遮罩 -->
    <Transition name="fade">
      <div
        v-if="drawerVisible"
        class="fixed inset-0 bg-black/30 z-40"
        @click="closeDrawer"
      ></div>
    </Transition>

    <!-- 抽屉 -->
    <Transition name="slide">
      <div
        v-if="drawerVisible && selectedBid"
        class="fixed top-18 bottom-0 right-0 w-[600px] bg-white shadow-2xl z-50 flex flex-col rounded-tl-lg"
      >
        <!-- 抽屉头部 -->
        <div class="flex items-center justify-between p-6 border-b border-gray-100">
          <h2 class="text-lg font-semibold text-gray-800">投标详情</h2>
          <button
            @click="closeDrawer"
            class="p-2 hover:bg-gray-100 rounded-full transition-colors"
          >
            <i class="fas fa-times text-gray-500"></i>
          </button>
        </div>

        <!-- 抽屉内容 -->
        <div class="flex-1 overflow-y-auto p-6">
          <!-- 悬赏标题 -->
          <h3 class="text-xl font-bold text-gray-900 mb-3">{{ selectedBid.bountyTitle }}</h3>

          <!-- 投标时间和状态 -->
          <div class="flex items-center gap-4 mb-4">
            <p class="text-gray-500 text-sm">
              <i class="fas fa-clock mr-2"></i>投标时间：{{ selectedBid.bidTime }}
            </p>
            <span
              class="px-3 py-1 rounded-full text-xs font-medium"
              :class="{
                'bg-yellow-100 text-yellow-600': selectedBid.status === 'pending',
                'bg-green-100 text-green-600': selectedBid.status === 'accepted',
                'bg-red-100 text-red-600': selectedBid.status === 'rejected',
                'bg-blue-100 text-blue-600': selectedBid.status === 'completed'
              }"
            >
              {{ getStatusText(selectedBid.status) }}
            </span>
          </div>

          <!-- 标签 -->
          <div class="flex flex-wrap gap-2 mb-6">
            <span class="bg-blue-50 text-blue-600 text-xs px-3 py-1.5 rounded-full">
              {{ bountyTypeMap[selectedBid.bountyType] || selectedBid.bountyType }}
            </span>
          </div>

          <!-- 投标金额 -->
          <div class="mb-6">
            <h4 class="text-sm font-semibold text-gray-700 mb-3">我的投标金额</h4>
            <div class="bg-orange-50 rounded-lg p-4">
              <p class="text-orange-600 text-2xl font-bold">¥{{ formatMoney(selectedBid.bidAmount) }}</p>
            </div>
          </div>

          <!-- 投标规格信息 -->
          <div class="mb-6">
            <h4 class="text-sm font-semibold text-gray-700 mb-3">投标规格</h4>

            <!-- 梭织规格 -->
            <div v-if="selectedBid.bountyType === 'woven' && selectedBid.wovenSpec" class="bg-gray-50 rounded-lg p-4 space-y-3">
              <div class="grid grid-cols-2 gap-4">
                <div v-if="selectedBid.wovenSpec.sizeLength">
                  <span class="text-xs text-gray-500">尺码（长度）</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedBid.wovenSpec.sizeLength }} 米</p>
                </div>
                <div v-if="selectedBid.wovenSpec.greigeFabricType">
                  <span class="text-xs text-gray-500">胚布类型</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedBid.wovenSpec.greigeFabricType }}</p>
                </div>
                <div v-if="selectedBid.wovenSpec.greigeDeliveryDate">
                  <span class="text-xs text-gray-500">胚布交期</span>
                  <p class="text-sm font-medium text-gray-800">{{ formatDate(selectedBid.wovenSpec.greigeDeliveryDate) }}</p>
                </div>
              </div>
            </div>

            <!-- 针织规格 -->
            <div v-else-if="selectedBid.bountyType === 'knitted' && selectedBid.knittedSpec" class="bg-gray-50 rounded-lg p-4 space-y-3">
              <div class="grid grid-cols-2 gap-4">
                <div v-if="selectedBid.knittedSpec.sizeWeight">
                  <span class="text-xs text-gray-500">尺码（重量）</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedBid.knittedSpec.sizeWeight }} kg</p>
                </div>
                <div v-if="selectedBid.knittedSpec.greigeFabricType">
                  <span class="text-xs text-gray-500">胚布类型</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedBid.knittedSpec.greigeFabricType }}</p>
                </div>
                <div v-if="selectedBid.knittedSpec.greigeDeliveryDate">
                  <span class="text-xs text-gray-500">胚布交期</span>
                  <p class="text-sm font-medium text-gray-800">{{ formatDate(selectedBid.knittedSpec.greigeDeliveryDate) }}</p>
                </div>
              </div>
            </div>

            <!-- 无规格信息 -->
            <div v-else class="bg-gray-50 rounded-lg p-4">
              <p class="text-sm text-gray-500">暂无投标规格信息</p>
            </div>
          </div>

          <!-- 截止日期 -->
          <div class="grid grid-cols-1 gap-4">
            <div class="bg-orange-50 rounded-lg p-4">
              <div class="flex items-center gap-2 text-orange-600">
                <i class="fas fa-calendar-alt"></i>
                <span class="text-sm font-medium">投标截止</span>
              </div>
              <p class="text-orange-700 font-semibold mt-1">{{ selectedBid.deadline }}</p>
            </div>
          </div>
        </div>

        <!-- 抽屉底部 -->
        <div class="p-6 border-t border-gray-100">
          <button
            @click="closeDrawer"
            class="w-full py-3 bg-gray-100 text-gray-700 rounded-lg font-medium hover:bg-gray-200 transition-colors"
          >
            关闭
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.slide-enter-active,
.slide-leave-active {
  transition: transform 0.3s ease;
}
.slide-enter-from,
.slide-leave-to {
  transform: translateX(100%);
}
</style>
