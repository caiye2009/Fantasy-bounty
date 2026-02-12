<script setup>
import { ref, onMounted, onUnmounted, inject, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Phone, Mail, Loader2, Inbox, X, Calendar, Truck, Tag } from 'lucide-vue-next'
import BountyCard from '@/components/BountyCard.vue'
import { placeBid } from '@/api/bid'
import { useBountyHall } from '@/composables/useBountyHall'
import { formatDate, formatComposition } from '@/utils/format'

// 从父组件注入登录状态和登录modal控制
const isLoggedIn = inject('isLoggedIn')
const openLoginModal = inject('openLoginModal')
const currentPage = inject('currentPage')

// ========== 数据层：composable ==========
const {
  searchKeyword,
  filterBeginDate,
  filterEndDate,
  filterIncludeEnd,
  displayTasks,
  loading,
  total,
  error,
  loadBountyList,
  clearFilters,
  hasFilters,
  init,
} = useBountyHall(isLoggedIn)

// ========== 登录检查 ==========
const requireLogin = (callback) => {
  if (!isLoggedIn.value) {
    openLoginModal()
    return false
  }
  if (callback) callback()
  return true
}

// ========== 搜索交互 ==========
let searchTimeout = null
const handleSearchInput = () => {
  if (!requireLogin()) return
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    loadBountyList()
  }, 400)
}

const handleSearchEnter = () => {
  if (!requireLogin()) return
  if (searchTimeout) clearTimeout(searchTimeout)
  loadBountyList()
}

const handleFilterChange = () => {
  if (!requireLogin()) return
  loadBountyList()
}

const handleClearFilters = () => {
  clearFilters()
}

const onFilterClick = (event) => {
  if (!isLoggedIn.value) {
    event.preventDefault()
    event.stopPropagation()
    openLoginModal()
  }
}

// ========== 抽屉 ==========
const drawerVisible = ref(false)
const selectedTask = ref(null)

const openDrawer = (task) => {
  if (!requireLogin()) return
  selectedTask.value = task
  drawerVisible.value = true
  document.body.style.overflow = 'hidden'
}

const closeDrawer = () => {
  drawerVisible.value = false
  selectedTask.value = null
  document.body.style.overflow = ''
}

// ========== 投标 Modal ==========
const bidModalVisible = ref(false)
const bidAmount = ref('')
const bidSubmitting = ref(false)

// 梭织投标字段
const wovenSizeLength = ref('')
const wovenGreigeFabricType = ref('')
const wovenGreigeDeliveryDate = ref('')
const wovenDeliveryMethod = ref('')

// 针织投标字段
const knittedSizeWeight = ref('')
const knittedGreigeFabricType = ref('')
const knittedGreigeDeliveryDate = ref('')
const knittedDeliveryMethod = ref('')

// 下拉选项
const greigeFabricTypeOptions = ['现货', '定织']
const deliveryMethodOptions = ['竞标确认后', '签订合同后', '收到预付款后']

const openBidModal = () => {
  bidAmount.value = ''
  wovenSizeLength.value = ''
  wovenGreigeFabricType.value = ''
  wovenGreigeDeliveryDate.value = ''
  wovenDeliveryMethod.value = ''
  knittedSizeWeight.value = ''
  knittedGreigeFabricType.value = ''
  knittedGreigeDeliveryDate.value = ''
  knittedDeliveryMethod.value = ''
  bidModalVisible.value = true
}

const closeBidModal = () => {
  bidModalVisible.value = false
  bidAmount.value = ''
}

const submitBid = async () => {
  if (!bidAmount.value || parseFloat(bidAmount.value) <= 0) {
    ElMessage.warning('请输入有效的投标金额')
    return
  }

  const bountyType = selectedTask.value.bountyType

  if (bountyType === 'woven') {
    if (!wovenSizeLength.value || !wovenGreigeFabricType.value || !wovenGreigeDeliveryDate.value || !wovenDeliveryMethod.value) {
      ElMessage.warning('请填写完整的梭织规格信息')
      return
    }
  } else if (bountyType === 'knitted') {
    if (!knittedSizeWeight.value || !knittedGreigeFabricType.value || !knittedGreigeDeliveryDate.value || !knittedDeliveryMethod.value) {
      ElMessage.warning('请填写完整的针织规格信息')
      return
    }
  }

  bidSubmitting.value = true
  try {
    const bidData = {
      bountyId: selectedTask.value.id,
      bidPrice: parseFloat(bidAmount.value)
    }

    if (bountyType === 'woven') {
      bidData.wovenSpec = {
        sizeLength: parseFloat(wovenSizeLength.value),
        greigeFabricType: wovenGreigeFabricType.value,
        greigeDeliveryDate: wovenGreigeDeliveryDate.value,
        deliveryMethod: wovenDeliveryMethod.value
      }
    } else if (bountyType === 'knitted') {
      bidData.knittedSpec = {
        sizeWeight: parseFloat(knittedSizeWeight.value),
        greigeFabricType: knittedGreigeFabricType.value,
        greigeDeliveryDate: knittedGreigeDeliveryDate.value,
        deliveryMethod: knittedDeliveryMethod.value
      }
    }

    await placeBid(bidData)
    ElMessage.success('投标成功！')
    closeBidModal()
    closeDrawer()
  } catch (err) {
    console.error('投标失败:', err)
    ElMessage.error('投标失败，请重试')
  } finally {
    bidSubmitting.value = false
  }
}

const handleBid = () => {
  openBidModal()
}

// ========== 键盘事件 ==========
const handleKeydown = (e) => {
  if (e.key === 'Escape') {
    if (bidModalVisible.value) {
      closeBidModal()
    } else if (drawerVisible.value) {
      closeDrawer()
    }
  }
}

// 切回本页时刷新数据
watch(currentPage, (page) => {
  if (page === 'hall') {
    init()
  }
})

onMounted(() => {
  init()
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
      <!-- 左侧：搜索筛选 + 页脚 -->
      <aside class="w-[30%] shrink-0 sticky top-24 h-[calc(100vh-7rem)] flex flex-col justify-between overflow-y-auto">
        <!-- 搜索和筛选 -->
        <div class="bg-white rounded-lg p-6 shadow-sm mb-4">
          <!-- 搜索框 -->
          <div class="relative mb-4">
            <input
              v-model="searchKeyword"
              type="text"
              placeholder="搜索产品名称..."
              class="w-full px-4 py-2.5 pl-10 border border-gray-200 rounded-lg text-sm focus:outline-none focus:border-blue-500"
              @mousedown="onFilterClick"
              @input="handleSearchInput"
              @keyup.enter="handleSearchEnter"
            >
            <Search class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400" :size="14" />
          </div>

          <!-- 筛选器 -->
          <div class="space-y-3">
            <div class="flex items-center gap-2">
              <span class="text-sm text-gray-600 whitespace-nowrap w-16">开始日期</span>
              <el-date-picker
                v-model="filterBeginDate"
                type="date"
                placeholder="不限"
                value-format="YYYY-MM-DD"
                class="!flex-1"
                @change="handleFilterChange"
              />
            </div>

            <div class="flex items-center gap-2">
              <span class="text-sm text-gray-600 whitespace-nowrap w-16">结束日期</span>
              <el-date-picker
                v-model="filterEndDate"
                type="date"
                placeholder="不限"
                value-format="YYYY-MM-DD"
                class="!flex-1"
                @change="handleFilterChange"
              />
            </div>

            <div class="flex items-center gap-2">
              <span class="text-sm text-gray-600 whitespace-nowrap w-16">已截止</span>
              <el-select
                v-model="filterIncludeEnd"
                class="flex-1"
                @change="handleFilterChange"
              >
                <el-option label="不包含" value="0" />
                <el-option label="包含" value="1" />
              </el-select>
            </div>
          </div>

          <!-- 搜索统计 + 清除 -->
          <div v-if="isLoggedIn" class="mt-4 text-xs text-gray-400">
            共 {{ total }} 条结果
            <span v-if="hasFilters()" class="ml-2">
              · <el-button type="primary" link size="small" @click="handleClearFilters">清除筛选</el-button>
            </span>
          </div>
        </div>

        <!-- 页脚信息 -->
        <div class="p-6 text-sm">
          <h3 class="font-semibold text-gray-800 mb-2">纺织悬赏大厅</h3>
          <p class="text-gray-500 text-xs mb-4 leading-relaxed">连接纺织业采购商与供应商的专业平台，让面料采购更高效、更透明。</p>

          <div class="grid grid-cols-2 gap-4 mb-4">
            <div>
              <h4 class="text-gray-700 font-medium mb-2">采购商服务</h4>
              <ul class="text-gray-500 text-xs space-y-1">
                <li>发布悬赏</li>
                <li>供应商筛选</li>
                <li>验厂服务</li>
              </ul>
            </div>
            <div>
              <h4 class="text-gray-700 font-medium mb-2">供应商服务</h4>
              <ul class="text-gray-500 text-xs space-y-1">
                <li>接单投标</li>
                <li>企业认证</li>
                <li>能力展示</li>
              </ul>
            </div>
          </div>

          <div class="border-t border-gray-100 pt-3">
            <h4 class="text-gray-700 font-medium mb-2">联系我们</h4>
            <div class="text-gray-500 text-xs space-y-1">
              <p><Phone class="inline-block mr-2" :size="14" />400-123-4567</p>
              <p><Mail class="inline-block mr-2" :size="14" />contact@textilebounty.com</p>
            </div>
          </div>
        </div>
      </aside>

      <!-- 右侧：任务列表 -->
      <div class="flex-1 min-w-0 flex flex-col gap-5">
        <!-- 加载状态 -->
        <div v-if="loading" class="bg-white rounded-lg p-16 shadow-sm text-center">
          <Loader2 class="text-gray-300 mb-4 animate-spin mx-auto" :size="40" />
          <p class="text-gray-500">加载中...</p>
        </div>

        <!-- 错误状态 -->
        <div v-else-if="error" class="bg-white rounded-lg p-16 shadow-sm text-center">
          <Inbox class="text-gray-300 mb-6 mx-auto" :size="60" />
          <h3 class="text-lg font-medium text-gray-600 mb-2">加载失败</h3>
          <p class="text-sm text-gray-400">请稍后再试</p>
        </div>

        <!-- 空状态 -->
        <div v-else-if="displayTasks.length === 0" class="bg-white rounded-lg p-16 shadow-sm text-center">
          <Inbox class="text-gray-300 mb-6 mx-auto" :size="60" />
          <h3 class="text-lg font-medium text-gray-600 mb-2">暂无悬赏任务</h3>
          <p class="text-sm text-gray-400">
            {{ isLoggedIn && searchKeyword ? '未找到匹配的结果，请尝试其他关键词' : '请稍后再来查看' }}
          </p>
        </div>

        <!-- 任务列表 -->
        <BountyCard
          v-else
          v-for="(task, index) in displayTasks"
          :key="task.id || index"
          :title="task.title"
          :publishTime="task.publishTime"
          :tags="task.tags"
          :description="task.description"
          :deadline="task.deadline"
          @click="openDrawer(task)"
        />
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
        v-if="drawerVisible && selectedTask"
        class="fixed top-18 bottom-0 right-0 w-[600px] bg-white shadow-2xl z-[101] flex flex-col rounded-tl-lg"
      >
        <!-- 抽屉头部 -->
        <div class="p-6 border-b border-gray-100">
          <div class="flex items-start justify-between">
            <div class="flex-1 pr-4">
              <h2 class="text-xl font-bold text-gray-900">{{ selectedTask.title }}</h2>
              <p class="text-gray-400 text-sm mt-1">{{ selectedTask.publishTime }}</p>
            </div>
            <button
              @click="closeDrawer"
              class="p-2 hover:bg-gray-100 rounded-full transition-colors shrink-0"
            >
              <X class="text-gray-500" :size="16" />
            </button>
          </div>
        </div>

        <!-- 抽屉内容 -->
        <div class="flex-1 overflow-y-auto p-6">
          <!-- 标签 -->
          <div class="flex flex-wrap gap-2 mb-6">
            <span
              v-for="(tag, index) in selectedTask.tags"
              :key="index"
              class="bg-blue-50 text-blue-600 text-xs px-3 py-1.5 rounded-full"
            >
              {{ tag }}
            </span>
          </div>

          <!-- 规格信息 -->
          <div class="mb-6">
            <h4 class="text-sm font-semibold text-gray-700 mb-3">规格信息</h4>

            <!-- 梭织规格 -->
            <div v-if="selectedTask.bountyType === 'woven' && selectedTask.wovenSpec" class="bg-gray-50 rounded-lg p-4 space-y-3">
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <span class="text-xs text-gray-500">产品名称</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.title }}</p>
                </div>
                <div v-if="selectedTask.productCode">
                  <span class="text-xs text-gray-500">产品编码</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.productCode }}</p>
                </div>
                <div v-if="selectedTask.wovenSpec.composition">
                  <span class="text-xs text-gray-500">面料成分</span>
                  <p class="text-sm font-medium text-gray-800">{{ formatComposition(selectedTask.wovenSpec.composition) }}</p>
                </div>
                <div v-if="selectedTask.wovenSpec.fabricWeight">
                  <span class="text-xs text-gray-500">成品克重</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.wovenSpec.fabricWeight }} g/m²</p>
                </div>
                <div v-if="selectedTask.wovenSpec.fabricWidth">
                  <span class="text-xs text-gray-500">成品幅宽</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.wovenSpec.fabricWidth }} cm</p>
                </div>
                <div v-if="selectedTask.wovenSpec.warpDensity">
                  <span class="text-xs text-gray-500">成品经密</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.wovenSpec.warpDensity }} 根/英寸</p>
                </div>
                <div v-if="selectedTask.wovenSpec.weftDensity">
                  <span class="text-xs text-gray-500">成品纬密</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.wovenSpec.weftDensity }} 根/英寸</p>
                </div>
                <div v-if="selectedTask.wovenSpec.warpMaterial">
                  <span class="text-xs text-gray-500">经向原料</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.wovenSpec.warpMaterial }}</p>
                </div>
                <div v-if="selectedTask.wovenSpec.weftMaterial">
                  <span class="text-xs text-gray-500">纬向原料</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.wovenSpec.weftMaterial }}</p>
                </div>
                <div v-if="selectedTask.wovenSpec.quantityMeters">
                  <span class="text-xs text-gray-500">需求数量</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.wovenSpec.quantityMeters }} 米</p>
                </div>
              </div>
            </div>

            <!-- 针织规格 -->
            <div v-else-if="selectedTask.bountyType === 'knitted' && selectedTask.knittedSpec" class="bg-gray-50 rounded-lg p-4 space-y-3">
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <span class="text-xs text-gray-500">产品名称</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.title }}</p>
                </div>
                <div v-if="selectedTask.productCode">
                  <span class="text-xs text-gray-500">产品编码</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.productCode }}</p>
                </div>
                <div v-if="selectedTask.knittedSpec.composition">
                  <span class="text-xs text-gray-500">面料成分</span>
                  <p class="text-sm font-medium text-gray-800">{{ formatComposition(selectedTask.knittedSpec.composition) }}</p>
                </div>
                <div v-if="selectedTask.knittedSpec.fabricWeight">
                  <span class="text-xs text-gray-500">成品克重</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.knittedSpec.fabricWeight }} g/m²</p>
                </div>
                <div v-if="selectedTask.knittedSpec.fabricWidth">
                  <span class="text-xs text-gray-500">成品幅宽</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.knittedSpec.fabricWidth }} cm</p>
                </div>
                <div v-if="selectedTask.knittedSpec.machineType">
                  <span class="text-xs text-gray-500">机型</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.knittedSpec.machineType }}</p>
                </div>
                <div v-if="selectedTask.knittedSpec.quantityKg">
                  <span class="text-xs text-gray-500">需求数量</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.knittedSpec.quantityKg }} kg</p>
                </div>
              </div>
              <!-- 原料明细 -->
              <div v-if="selectedTask.knittedSpec.materials && selectedTask.knittedSpec.materials.length > 0" class="pt-2 border-t border-gray-200">
                <span class="text-xs text-gray-500 block mb-2">原料明细</span>
                <div class="flex flex-wrap gap-2">
                  <span
                    v-for="(material, idx) in selectedTask.knittedSpec.materials"
                    :key="idx"
                    class="bg-blue-50 text-blue-700 text-xs px-2 py-1 rounded"
                  >
                    {{ material.name }} {{ (material.percentage * 100) % 1 === 0 ? (material.percentage * 100).toFixed(0) : (material.percentage * 100).toFixed(1) }}%
                  </span>
                </div>
              </div>
            </div>

            <!-- 无规格信息 -->
            <div v-else class="bg-gray-50 rounded-lg p-4">
              <div class="grid grid-cols-2 gap-4 mb-3">
                <div>
                  <span class="text-xs text-gray-500">产品名称</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.title }}</p>
                </div>
                <div v-if="selectedTask.productCode">
                  <span class="text-xs text-gray-500">产品编码</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.productCode }}</p>
                </div>
              </div>
              <p class="text-sm text-gray-500">暂无其他规格信息</p>
            </div>
          </div>

          <!-- 日期信息 -->
          <div class="grid grid-cols-2 gap-4">
            <div class="bg-orange-50 rounded-lg p-4">
              <div class="flex items-center gap-2 text-orange-600">
                <Calendar :size="14" />
                <span class="text-sm font-medium">投标截止</span>
              </div>
              <p class="text-orange-700 font-semibold mt-1">{{ selectedTask.deadline }}</p>
            </div>
            <div v-if="selectedTask.expectedDeliveryDate" class="bg-green-50 rounded-lg p-4">
              <div class="flex items-center gap-2 text-green-600">
                <Truck :size="14" />
                <span class="text-sm font-medium">期望交付</span>
              </div>
              <p class="text-green-700 font-semibold mt-1">{{ formatDate(selectedTask.expectedDeliveryDate) }}</p>
            </div>
          </div>
        </div>

        <!-- 抽屉底部 -->
        <div class="p-6 border-t border-gray-100">
          <el-button
            type="primary"
            @click="handleBid"
            class="!w-full !py-5 !text-base"
          >
            立即投标
          </el-button>
        </div>
      </div>
    </Transition>

    <!-- 投标 Modal 遮罩 -->
    <Transition name="fade">
      <div
        v-if="bidModalVisible"
        class="fixed inset-0 bg-black/50 z-[102] flex items-center justify-center"
        @click.self="closeBidModal"
      >
        <!-- Modal 内容 -->
        <div class="bg-white rounded-lg w-[450px] shadow-xl">
          <!-- Modal 头部 -->
          <div class="flex items-center justify-between p-5 border-b border-gray-100">
            <h3 class="text-lg font-semibold text-gray-800">提交投标</h3>
            <button
              @click="closeBidModal"
              class="p-1.5 hover:bg-gray-100 rounded-full transition-colors"
            >
              <X class="text-gray-500" :size="16" />
            </button>
          </div>

          <!-- Modal 内容 -->
          <div class="p-5">
            <!-- 悬赏信息 -->
            <div class="bg-gray-50 rounded-lg p-4 mb-5">
              <p class="text-sm text-gray-600 mb-1">投标悬赏：</p>
              <p class="font-medium text-gray-800">{{ selectedTask?.title }}</p>
              <p class="text-xs text-blue-600 mt-1">
                <Tag class="inline-block mr-1" :size="12" />
                {{ selectedTask?.bountyType === 'woven' ? '梭织' : '针织' }}
              </p>
            </div>

            <!-- 投标金额 -->
            <div class="mb-4">
              <label class="block text-sm font-medium text-gray-700 mb-2">
                投标金额 <span class="text-red-500">*</span>
              </label>
              <el-input
                v-model="bidAmount"
                type="number"
                placeholder="请输入投标金额"
              >
                <template #prefix>¥</template>
              </el-input>
            </div>

            <!-- 梭织规格字段 -->
            <template v-if="selectedTask?.bountyType === 'woven'">
              <div class="mb-4">
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  尺码（长度/米） <span class="text-red-500">*</span>
                </label>
                <el-input
                  v-model="wovenSizeLength"
                  type="number"
                  placeholder="请输入尺码长度"
                />
              </div>
              <div class="mb-4">
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  胚布类型 <span class="text-red-500">*</span>
                </label>
                <el-select v-model="wovenGreigeFabricType" placeholder="请选择胚布类型" class="w-full">
                  <el-option v-for="opt in greigeFabricTypeOptions" :key="opt" :label="opt" :value="opt" />
                </el-select>
              </div>
              <div class="mb-4">
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  胚布交期 <span class="text-red-500">*</span>
                </label>
                <el-date-picker
                  v-model="wovenGreigeDeliveryDate"
                  type="date"
                  placeholder="请选择日期"
                  value-format="YYYY-MM-DD"
                  class="!w-full"
                />
              </div>
              <div class="mb-5">
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  交货方式 <span class="text-red-500">*</span>
                </label>
                <el-select v-model="wovenDeliveryMethod" placeholder="请选择交货方式" class="w-full">
                  <el-option v-for="opt in deliveryMethodOptions" :key="opt" :label="opt" :value="opt" />
                </el-select>
              </div>
            </template>

            <!-- 针织规格字段 -->
            <template v-else-if="selectedTask?.bountyType === 'knitted'">
              <div class="mb-4">
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  尺码（重量/kg） <span class="text-red-500">*</span>
                </label>
                <el-input
                  v-model="knittedSizeWeight"
                  type="number"
                  placeholder="请输入尺码重量"
                />
              </div>
              <div class="mb-4">
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  胚布类型 <span class="text-red-500">*</span>
                </label>
                <el-select v-model="knittedGreigeFabricType" placeholder="请选择胚布类型" class="w-full">
                  <el-option v-for="opt in greigeFabricTypeOptions" :key="opt" :label="opt" :value="opt" />
                </el-select>
              </div>
              <div class="mb-4">
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  胚布交期 <span class="text-red-500">*</span>
                </label>
                <el-date-picker
                  v-model="knittedGreigeDeliveryDate"
                  type="date"
                  placeholder="请选择日期"
                  value-format="YYYY-MM-DD"
                  class="!w-full"
                />
              </div>
              <div class="mb-5">
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  交货方式 <span class="text-red-500">*</span>
                </label>
                <el-select v-model="knittedDeliveryMethod" placeholder="请选择交货方式" class="w-full">
                  <el-option v-for="opt in deliveryMethodOptions" :key="opt" :label="opt" :value="opt" />
                </el-select>
              </div>
            </template>
          </div>

          <!-- Modal 底部 -->
          <div class="flex gap-3 p-5 border-t border-gray-100">
            <el-button @click="closeBidModal" class="!flex-1 !py-5">
              取消
            </el-button>
            <el-button
              type="primary"
              @click="submitBid"
              :disabled="bidSubmitting"
              :loading="bidSubmitting"
              class="!flex-1 !py-5"
            >
              {{ bidSubmitting ? '提交中...' : '确认投标' }}
            </el-button>
          </div>
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
