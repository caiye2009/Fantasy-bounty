<script setup>
import { onMounted, onUnmounted, ref, inject, watch } from 'vue'
import { Loader2, Tag, Clock, CalendarX, Handshake, Search, X, Barcode, Calendar } from 'lucide-vue-next'
import { useMyBids } from '@/composables/useMyBids'
import { bountyTypeMap, formatDate, formatMoney, getStatusText } from '@/utils/format'

const currentPage = inject('currentPage')

// ========== 数据层：composable ==========
const {
  selectedStatus,
  myBids,
  loading,
  error,
  loadMyBids,
  statusOptions,
} = useMyBids()

// 切到本页时刷新数据
watch(currentPage, (page) => {
  if (page === 'myBids') {
    loadMyBids()
  }
})

// ========== 抽屉 ==========
const drawerVisible = ref(false)
const selectedBid = ref(null)

const openDrawer = (bid) => {
  selectedBid.value = bid
  drawerVisible.value = true
  document.body.style.overflow = 'hidden'
}

const closeDrawer = () => {
  drawerVisible.value = false
  selectedBid.value = null
  document.body.style.overflow = ''
}

// ========== 键盘事件 ==========
const handleKeydown = (e) => {
  if (e.key === 'Escape' && drawerVisible.value) {
    closeDrawer()
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})
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
              <el-select v-model="selectedStatus" placeholder="全部" class="flex-1">
                <el-option v-for="opt in statusOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
              </el-select>
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
          <Loader2 class="text-gray-300 mb-4 animate-spin mx-auto" :size="40" />
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
                    <Tag :size="12" />{{ bountyTypeMap[bid.bountyType] || bid.bountyType }}
                  </span>
                  <span class="flex items-center gap-1">
                    <Clock :size="12" />投标时间: {{ bid.bidTime }}
                  </span>
                </div>
              </div>
              <div class="flex items-center gap-3">
                <span
                  class="px-3 py-1 rounded-full text-xs font-medium"
                  :class="{
                    'bg-yellow-100 text-yellow-600': bid.status === 'pending',
                    'bg-blue-100 text-blue-600': bid.status === 'in_progress',
                    'bg-orange-100 text-orange-600': bid.status === 'pending_acceptance',
                    'bg-green-100 text-green-600': bid.status === 'completed'
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
                  <CalendarX :size="12" />投标截止: {{ bid.deadline }}
                </span>
              </div>
              <el-button
                type="primary"
                plain
                size="small"
                @click.stop="openDrawer(bid)"
              >
                查看详情
              </el-button>
            </div>
          </div>
        </template>

        <!-- 错误状态 -->
        <div v-else-if="error" class="bg-white rounded-lg p-16 shadow-sm text-center">
          <Handshake class="text-gray-300 mb-6 mx-auto" :size="60" />
          <h3 class="text-lg font-medium text-gray-600 mb-2">加载失败</h3>
          <p class="text-sm text-gray-400">请稍后再试</p>
        </div>

        <!-- 空状态 -->
        <div v-else class="bg-white rounded-lg p-16 shadow-sm text-center">
          <Handshake class="text-gray-300 mb-6 mx-auto" :size="60" />
          <h3 class="text-lg font-medium text-gray-600 mb-2">暂无投标记录</h3>
          <p class="text-sm text-gray-400 mb-6">去悬赏大厅查看需求</p>
          <router-link to="/">
            <el-button type="primary">
              <Search class="mr-2" :size="14" />浏览悬赏
            </el-button>
          </router-link>
        </div>
      </div>
    </div>

    <!-- 抽屉遮罩 -->
    <Transition name="fade">
      <div
        v-if="drawerVisible"
        class="fixed inset-0 bg-black/30 z-[100]"
        @click="closeDrawer"
      ></div>
    </Transition>

    <!-- 抽屉 -->
    <Transition name="slide">
      <div
        v-if="drawerVisible && selectedBid"
        class="fixed top-18 bottom-0 right-0 w-[600px] bg-white shadow-2xl z-[101] flex flex-col rounded-tl-lg"
      >
        <!-- 抽屉头部 -->
        <div class="flex items-center justify-between p-6 border-b border-gray-100">
          <h2 class="text-lg font-semibold text-gray-800">投标详情</h2>
          <button
            @click="closeDrawer"
            class="p-2 hover:bg-gray-100 rounded-full transition-colors"
          >
            <X class="text-gray-500" :size="16" />
          </button>
        </div>

        <!-- 抽屉内容 -->
        <div class="flex-1 overflow-y-auto p-6">
          <!-- 悬赏标题 -->
          <h3 class="text-xl font-bold text-gray-900 mb-2">{{ selectedBid.bountyTitle }}</h3>

          <!-- 产品编码 -->
          <p v-if="selectedBid.bountyProductCode" class="text-gray-500 text-sm mb-3">
            <Barcode class="inline-block mr-2" :size="14" />产品编码：{{ selectedBid.bountyProductCode }}
          </p>

          <!-- 投标时间和状态 -->
          <div class="flex items-center gap-4 mb-4">
            <p class="text-gray-500 text-sm">
              <Clock class="inline-block mr-2" :size="14" />投标时间：{{ selectedBid.bidTime }}
            </p>
            <span
              class="px-3 py-1 rounded-full text-xs font-medium"
              :class="{
                'bg-yellow-100 text-yellow-600': selectedBid.status === 'pending',
                'bg-blue-100 text-blue-600': selectedBid.status === 'in_progress',
                'bg-orange-100 text-orange-600': selectedBid.status === 'pending_acceptance',
                'bg-green-100 text-green-600': selectedBid.status === 'completed'
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
                <div v-if="selectedBid.wovenSpec.deliveryMethod">
                  <span class="text-xs text-gray-500">交货方式</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedBid.wovenSpec.deliveryMethod }}</p>
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
                <div v-if="selectedBid.knittedSpec.deliveryMethod">
                  <span class="text-xs text-gray-500">交货方式</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedBid.knittedSpec.deliveryMethod }}</p>
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
                <Calendar :size="14" />
                <span class="text-sm font-medium">投标截止</span>
              </div>
              <p class="text-orange-700 font-semibold mt-1">{{ selectedBid.deadline }}</p>
            </div>
          </div>
        </div>

        <!-- 抽屉底部 -->
        <div class="p-6 border-t border-gray-100">
          <el-button @click="closeDrawer" class="!w-full !py-5 !text-base">
            关闭
          </el-button>
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
