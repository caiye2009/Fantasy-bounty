<script setup>
import { ref, onMounted, inject, watch } from 'vue'
import BountyCard from '@/components/BountyCard.vue'
import { fetchBountyList, peekBountyList } from '@/api/bounty'
import { placeBid } from '@/api/bid'

// 从父组件注入登录状态和登录modal控制
const isLoggedIn = inject('isLoggedIn')
const openLoginModal = inject('openLoginModal')

// 需要登录才能执行的操作
const requireLogin = (callback) => {
  if (!isLoggedIn.value) {
    openLoginModal()
    return false
  }
  if (callback) callback()
  return true
}

// 搜索关键词
const searchKeyword = ref('')

// 筛选条件 - 下拉框选项
const fabricTypeOptions = [
  { label: '全部', value: '' },
  { label: '棉布', value: 'cotton' },
  { label: '涤纶', value: 'polyester' },
  { label: '尼龙', value: 'nylon' },
  { label: '丝绸', value: 'silk' },
  { label: '麻布', value: 'linen' }
]
const selectedFabricType = ref('')

const priceRangeOptions = [
  { label: '全部', value: '' },
  { label: '1万以下', value: '0-10000' },
  { label: '1-5万', value: '10000-50000' },
  { label: '5-10万', value: '50000-100000' },
  { label: '10万以上', value: '100000+' }
]
const selectedPriceRange = ref('')

const craftOptions = [
  { label: '全部', value: '' },
  { label: '染色', value: 'dye' },
  { label: '印花', value: 'print' },
  { label: '提花', value: 'jacquard' },
  { label: '刺绣', value: 'embroidery' },
  { label: '涂层', value: 'coating' },
  { label: '复合', value: 'composite' }
]
const selectedCraft = ref('')

// 排序选项
const sortOptions = [
  { label: '默认', value: '' },
  { label: '金额', value: 'price' },
  { label: '发布时间', value: 'publish' },
  { label: '截止时间', value: 'deadline' },
  { label: '米数', value: 'quantity' }
]
const selectedSort = ref('')
const sortAsc = ref(false)

const toggleSortOrder = () => {
  // 排序需要登录
  if (!requireLogin()) return
  sortAsc.value = !sortAsc.value
}

// 点击搜索框或筛选器时检查登录（未登录则阻止操作并弹出登录框）
const onFilterClick = (event) => {
  if (!isLoggedIn.value) {
    event.preventDefault()
    event.target.blur()
    openLoginModal()
  }
}

// 任务列表
const tasks = ref([])
const loading = ref(false)

// 悬赏类型中英文对照
const bountyTypeMap = {
  woven: '梭织',
  knitted: '针织'
}

// 状态中英文对照
const statusMap = {
  open: '招标中',
  in_progress: '进行中',
  completed: '已完成',
  closed: '已关闭'
}

// 生成悬赏描述
const generateDescription = (item) => {
  const parts = []
  if (item.bountyType === 'woven' && item.wovenSpec) {
    const spec = item.wovenSpec
    if (spec.composition) parts.push(`成分: ${spec.composition}`)
    if (spec.fabricWeight) parts.push(`克重: ${spec.fabricWeight}g/m²`)
    if (spec.fabricWidth) parts.push(`幅宽: ${spec.fabricWidth}cm`)
    if (spec.quantityMeters) parts.push(`需求: ${spec.quantityMeters}米`)
  } else if (item.bountyType === 'knitted' && item.knittedSpec) {
    const spec = item.knittedSpec
    if (spec.composition) parts.push(`成分: ${spec.composition}`)
    if (spec.fabricWeight) parts.push(`克重: ${spec.fabricWeight}g/m²`)
    if (spec.fabricWidth) parts.push(`幅宽: ${spec.fabricWidth}cm`)
    if (spec.quantityKg) parts.push(`需求: ${spec.quantityKg}kg`)
  }
  return parts.length > 0 ? parts.join(' | ') : '暂无详细规格'
}

// 生成标签
const generateTags = (item) => {
  const tags = []
  if (item.bountyType) tags.push(bountyTypeMap[item.bountyType] || item.bountyType)
  if (item.sampleType) tags.push(item.sampleType)
  if (item.status) tags.push(statusMap[item.status] || item.status)
  return tags
}

// 加载悬赏列表
const loadBounties = async () => {
  loading.value = true
  try {
    // 根据登录状态调用不同的API
    let result
    if (isLoggedIn.value) {
      // 已登录：调用完整列表API
      result = await fetchBountyList(1, 50)
    } else {
      // 未登录：调用peek API，只获取第一页10条
      result = await peekBountyList()
    }
    tasks.value = result.data.map(item => ({
      id: item.id,
      title: item.productName,
      publishTime: formatDateTime(item.createdAt),
      tags: generateTags(item),
      description: generateDescription(item),
      deadline: formatDate(item.bidDeadline),
      bountyType: item.bountyType,
      status: item.status,
      wovenSpec: item.wovenSpec,
      knittedSpec: item.knittedSpec,
      expectedDeliveryDate: item.expectedDeliveryDate
    }))
  } catch (error) {
    console.error('加载悬赏列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 监听登录状态变化，重新加载列表
watch(isLoggedIn, () => {
  loadBounties()
})

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

onMounted(() => {
  loadBounties()
})

const handleSearch = () => {
  // TODO: 接入远程搜索API
  console.log('搜索:', searchKeyword.value, selectedFabricType.value, selectedPriceRange.value, selectedCraft.value)
}

// 抽屉状态
const drawerVisible = ref(false)
const selectedTask = ref(null)

const openDrawer = (task) => {
  // 查看详情需要登录
  if (!requireLogin()) return
  selectedTask.value = task
  drawerVisible.value = true
}

const closeDrawer = () => {
  drawerVisible.value = false
  selectedTask.value = null
}

// 投标 Modal 状态
const bidModalVisible = ref(false)
const bidAmount = ref('')
const bidSubmitting = ref(false)

// 梭织投标字段
const wovenSizeLength = ref('')
const wovenGreigeFabricType = ref('')
const wovenGreigeDeliveryDate = ref('')

// 针织投标字段
const knittedSizeWeight = ref('')
const knittedGreigeFabricType = ref('')
const knittedGreigeDeliveryDate = ref('')

const openBidModal = () => {
  bidAmount.value = ''
  // 重置梭织字段
  wovenSizeLength.value = ''
  wovenGreigeFabricType.value = ''
  wovenGreigeDeliveryDate.value = ''
  // 重置针织字段
  knittedSizeWeight.value = ''
  knittedGreigeFabricType.value = ''
  knittedGreigeDeliveryDate.value = ''
  bidModalVisible.value = true
}

const closeBidModal = () => {
  bidModalVisible.value = false
  bidAmount.value = ''
}

const submitBid = async () => {
  if (!bidAmount.value || parseFloat(bidAmount.value) <= 0) {
    alert('请输入有效的投标金额')
    return
  }

  const bountyType = selectedTask.value.bountyType

  // 验证必填字段
  if (bountyType === 'woven') {
    if (!wovenSizeLength.value || !wovenGreigeFabricType.value || !wovenGreigeDeliveryDate.value) {
      alert('请填写完整的梭织规格信息')
      return
    }
  } else if (bountyType === 'knitted') {
    if (!knittedSizeWeight.value || !knittedGreigeFabricType.value || !knittedGreigeDeliveryDate.value) {
      alert('请填写完整的针织规格信息')
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
        greigeDeliveryDate: wovenGreigeDeliveryDate.value
      }
    } else if (bountyType === 'knitted') {
      bidData.knittedSpec = {
        sizeWeight: parseFloat(knittedSizeWeight.value),
        greigeFabricType: knittedGreigeFabricType.value,
        greigeDeliveryDate: knittedGreigeDeliveryDate.value
      }
    }

    await placeBid(bidData)
    alert('投标成功！')
    closeBidModal()
    closeDrawer()
  } catch (error) {
    console.error('投标失败:', error)
    alert('投标失败，请重试')
  } finally {
    bidSubmitting.value = false
  }
}

const handleBid = () => {
  openBidModal()
}
</script>

<template>
  <div class="w-[80%] mx-auto py-6">
    <!-- 左右布局容器 -->
    <div class="flex gap-6">
      <!-- 左侧：搜索和筛选区域 -->
      <aside class="w-[30%] shrink-0 sticky top-24 h-[calc(100vh-7rem)] flex flex-col justify-between">
        <!-- 筛选框 -->
        <div class="bg-white rounded-lg p-8 shadow-sm">
          <!-- 搜索框 -->
          <div class="relative mb-6">
            <input
              v-model="searchKeyword"
              type="text"
              placeholder="搜索悬赏任务..."
              class="w-full px-4 py-2.5 pl-10 border border-gray-200 rounded-lg text-sm focus:outline-none focus:border-blue-500"
              @focus="onFilterClick"
              @keyup.enter="handleSearch"
            >
            <i class="fas fa-search absolute left-4 top-1/2 -translate-y-1/2 text-gray-400 text-sm"></i>
          </div>

          <!-- 筛选条件 -->
          <div class="flex flex-col gap-5">
            <!-- 布料类型 -->
            <div class="flex items-center justify-between gap-3">
              <span class="text-sm text-gray-600 whitespace-nowrap">布料类型</span>
              <select
                v-model="selectedFabricType"
                @mousedown="onFilterClick"
                class="px-3 py-1.5 border border-gray-200 rounded text-sm text-gray-600 focus:outline-none focus:border-blue-500 flex-1"
              >
                <option v-for="opt in fabricTypeOptions" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </option>
              </select>
            </div>

            <!-- 悬赏金额 -->
            <div class="flex items-center justify-between gap-3">
              <span class="text-sm text-gray-600 whitespace-nowrap">悬赏金额</span>
              <select
                v-model="selectedPriceRange"
                @mousedown="onFilterClick"
                class="px-3 py-1.5 border border-gray-200 rounded text-sm text-gray-600 focus:outline-none focus:border-blue-500 flex-1"
              >
                <option v-for="opt in priceRangeOptions" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </option>
              </select>
            </div>

            <!-- 面料工艺 -->
            <div class="flex items-center justify-between gap-3">
              <span class="text-sm text-gray-600 whitespace-nowrap">面料工艺</span>
              <select
                v-model="selectedCraft"
                @mousedown="onFilterClick"
                class="px-3 py-1.5 border border-gray-200 rounded text-sm text-gray-600 focus:outline-none focus:border-blue-500 flex-1"
              >
                <option v-for="opt in craftOptions" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </option>
              </select>
            </div>

            <!-- 排序方式 -->
            <div class="flex items-center gap-2">
              <span class="text-sm text-gray-600 whitespace-nowrap">排序方式</span>
              <select
                v-model="selectedSort"
                @mousedown="onFilterClick"
                class="px-3 py-1.5 border border-gray-200 rounded text-sm text-gray-600 focus:outline-none focus:border-blue-500 flex-1"
              >
                <option v-for="opt in sortOptions" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </option>
              </select>
              <button
                @click="toggleSortOrder"
                class="p-1.5 border border-gray-200 rounded hover:bg-gray-50 transition-colors"
                :class="{ 'text-blue-500 border-blue-300': selectedSort }"
                :disabled="!selectedSort"
              >
                <i :class="sortAsc ? 'fas fa-sort-amount-up' : 'fas fa-sort-amount-down'" class="text-sm"></i>
              </button>
            </div>
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
              <p><i class="fas fa-phone-alt mr-2"></i>400-123-4567</p>
              <p><i class="fas fa-envelope mr-2"></i>contact@textilebounty.com</p>
            </div>
          </div>
        </div>
      </aside>

      <!-- 右侧：任务列表 -->
      <div class="flex-1 min-w-0 flex flex-col gap-5">
        <!-- 加载状态 -->
        <div v-if="loading" class="bg-white rounded-lg p-16 shadow-sm text-center">
          <i class="fas fa-spinner fa-spin text-4xl text-gray-300 mb-4"></i>
          <p class="text-gray-500">加载中...</p>
        </div>

        <!-- 空状态 -->
        <div v-else-if="tasks.length === 0" class="bg-white rounded-lg p-16 shadow-sm text-center">
          <i class="fas fa-inbox text-6xl text-gray-300 mb-6"></i>
          <h3 class="text-lg font-medium text-gray-600 mb-2">暂无悬赏任务</h3>
          <p class="text-sm text-gray-400">请稍后再来查看</p>
        </div>

        <!-- 任务列表 -->
        <BountyCard
          v-else
          v-for="(task, index) in tasks"
          :key="index"
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
        class="fixed inset-0 bg-black/30 z-40"
        @click="closeDrawer"
      ></div>
    </Transition>

    <!-- 抽屉 -->
    <Transition name="slide">
      <div
        v-if="drawerVisible && selectedTask"
        class="fixed top-18 bottom-0 right-0 w-[600px] bg-white shadow-2xl z-50 flex flex-col rounded-tl-lg"
      >
        <!-- 抽屉头部 -->
        <div class="flex items-center justify-between p-6 border-b border-gray-100">
          <h2 class="text-lg font-semibold text-gray-800">悬赏详情</h2>
          <button
            @click="closeDrawer"
            class="p-2 hover:bg-gray-100 rounded-full transition-colors"
          >
            <i class="fas fa-times text-gray-500"></i>
          </button>
        </div>

        <!-- 抽屉内容 -->
        <div class="flex-1 overflow-y-auto p-6">
          <!-- 标题 -->
          <h3 class="text-xl font-bold text-gray-900 mb-3">{{ selectedTask.title }}</h3>

          <!-- 发布时间 -->
          <p class="text-gray-500 text-sm mb-4">
            <i class="fas fa-clock mr-2"></i>发布时间：{{ selectedTask.publishTime }}
          </p>

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
                <div v-if="selectedTask.wovenSpec.composition">
                  <span class="text-xs text-gray-500">成分</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.wovenSpec.composition }}</p>
                </div>
                <div v-if="selectedTask.wovenSpec.yarnCount">
                  <span class="text-xs text-gray-500">纱支</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.wovenSpec.yarnCount }}</p>
                </div>
                <div v-if="selectedTask.wovenSpec.density">
                  <span class="text-xs text-gray-500">密度</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.wovenSpec.density }}</p>
                </div>
                <div v-if="selectedTask.wovenSpec.weavePattern">
                  <span class="text-xs text-gray-500">组织结构</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.wovenSpec.weavePattern }}</p>
                </div>
                <div v-if="selectedTask.wovenSpec.fabricWeight">
                  <span class="text-xs text-gray-500">克重</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.wovenSpec.fabricWeight }} g/m²</p>
                </div>
                <div v-if="selectedTask.wovenSpec.fabricWidth">
                  <span class="text-xs text-gray-500">幅宽</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.wovenSpec.fabricWidth }} cm</p>
                </div>
                <div v-if="selectedTask.wovenSpec.quantityMeters">
                  <span class="text-xs text-gray-500">需求数量</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.wovenSpec.quantityMeters }} 米</p>
                </div>
                <div v-if="selectedTask.wovenSpec.colorRequirement">
                  <span class="text-xs text-gray-500">颜色要求</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.wovenSpec.colorRequirement }}</p>
                </div>
              </div>
              <div v-if="selectedTask.wovenSpec.specialProcess" class="pt-2 border-t border-gray-200">
                <span class="text-xs text-gray-500">特殊工艺</span>
                <p class="text-sm font-medium text-gray-800">{{ selectedTask.wovenSpec.specialProcess }}</p>
              </div>
            </div>

            <!-- 针织规格 -->
            <div v-else-if="selectedTask.bountyType === 'knitted' && selectedTask.knittedSpec" class="bg-gray-50 rounded-lg p-4 space-y-3">
              <div class="grid grid-cols-2 gap-4">
                <div v-if="selectedTask.knittedSpec.composition">
                  <span class="text-xs text-gray-500">成分</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.knittedSpec.composition }}</p>
                </div>
                <div v-if="selectedTask.knittedSpec.yarnCount">
                  <span class="text-xs text-gray-500">纱支</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.knittedSpec.yarnCount }}</p>
                </div>
                <div v-if="selectedTask.knittedSpec.knitPattern">
                  <span class="text-xs text-gray-500">针织类型</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.knittedSpec.knitPattern }}</p>
                </div>
                <div v-if="selectedTask.knittedSpec.gauge">
                  <span class="text-xs text-gray-500">针数规格</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.knittedSpec.gauge }}</p>
                </div>
                <div v-if="selectedTask.knittedSpec.fabricWeight">
                  <span class="text-xs text-gray-500">克重</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.knittedSpec.fabricWeight }} g/m²</p>
                </div>
                <div v-if="selectedTask.knittedSpec.fabricWidth">
                  <span class="text-xs text-gray-500">幅宽</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.knittedSpec.fabricWidth }} cm</p>
                </div>
                <div v-if="selectedTask.knittedSpec.quantityKg">
                  <span class="text-xs text-gray-500">需求数量</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.knittedSpec.quantityKg }} kg</p>
                </div>
                <div v-if="selectedTask.knittedSpec.colorRequirement">
                  <span class="text-xs text-gray-500">颜色要求</span>
                  <p class="text-sm font-medium text-gray-800">{{ selectedTask.knittedSpec.colorRequirement }}</p>
                </div>
              </div>
              <div v-if="selectedTask.knittedSpec.specialProcess" class="pt-2 border-t border-gray-200">
                <span class="text-xs text-gray-500">特殊工艺</span>
                <p class="text-sm font-medium text-gray-800">{{ selectedTask.knittedSpec.specialProcess }}</p>
              </div>
            </div>

            <!-- 无规格信息 -->
            <div v-else class="bg-gray-50 rounded-lg p-4">
              <p class="text-sm text-gray-500">暂无详细规格信息</p>
            </div>
          </div>

          <!-- 日期信息 -->
          <div class="grid grid-cols-2 gap-4">
            <div class="bg-orange-50 rounded-lg p-4">
              <div class="flex items-center gap-2 text-orange-600">
                <i class="fas fa-calendar-alt"></i>
                <span class="text-sm font-medium">投标截止</span>
              </div>
              <p class="text-orange-700 font-semibold mt-1">{{ selectedTask.deadline }}</p>
            </div>
            <div v-if="selectedTask.expectedDeliveryDate" class="bg-green-50 rounded-lg p-4">
              <div class="flex items-center gap-2 text-green-600">
                <i class="fas fa-truck"></i>
                <span class="text-sm font-medium">期望交付</span>
              </div>
              <p class="text-green-700 font-semibold mt-1">{{ formatDate(selectedTask.expectedDeliveryDate) }}</p>
            </div>
          </div>
        </div>

        <!-- 抽屉底部 -->
        <div class="p-6 border-t border-gray-100">
          <button
            @click="handleBid"
            class="w-full py-3 bg-blue-500 text-white rounded-lg font-medium hover:bg-blue-600 transition-colors"
          >
            立即投标
          </button>
        </div>
      </div>
    </Transition>

    <!-- 投标 Modal 遮罩 -->
    <Transition name="fade">
      <div
        v-if="bidModalVisible"
        class="fixed inset-0 bg-black/50 z-[60] flex items-center justify-center"
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
              <i class="fas fa-times text-gray-500"></i>
            </button>
          </div>

          <!-- Modal 内容 -->
          <div class="p-5">
            <!-- 悬赏信息 -->
            <div class="bg-gray-50 rounded-lg p-4 mb-5">
              <p class="text-sm text-gray-600 mb-1">投标悬赏：</p>
              <p class="font-medium text-gray-800">{{ selectedTask?.title }}</p>
              <p class="text-xs text-blue-600 mt-1">
                <i class="fas fa-tag mr-1"></i>
                {{ selectedTask?.bountyType === 'woven' ? '梭织' : '针织' }}
              </p>
            </div>

            <!-- 投标金额 -->
            <div class="mb-4">
              <label class="block text-sm font-medium text-gray-700 mb-2">
                投标金额 <span class="text-red-500">*</span>
              </label>
              <div class="relative">
                <span class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500">¥</span>
                <input
                  v-model="bidAmount"
                  type="number"
                  min="0"
                  step="0.01"
                  placeholder="请输入投标金额"
                  class="w-full pl-8 pr-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:border-blue-500"
                >
              </div>
            </div>

            <!-- 梭织规格字段 -->
            <template v-if="selectedTask?.bountyType === 'woven'">
              <div class="mb-4">
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  尺码（长度/米） <span class="text-red-500">*</span>
                </label>
                <input
                  v-model="wovenSizeLength"
                  type="number"
                  min="0"
                  step="0.01"
                  placeholder="请输入尺码长度"
                  class="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:border-blue-500"
                >
              </div>
              <div class="mb-4">
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  胚布类型 <span class="text-red-500">*</span>
                </label>
                <input
                  v-model="wovenGreigeFabricType"
                  type="text"
                  placeholder="请输入胚布类型"
                  class="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:border-blue-500"
                >
              </div>
              <div class="mb-5">
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  胚布交期 <span class="text-red-500">*</span>
                </label>
                <input
                  v-model="wovenGreigeDeliveryDate"
                  type="date"
                  class="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:border-blue-500"
                >
              </div>
            </template>

            <!-- 针织规格字段 -->
            <template v-else-if="selectedTask?.bountyType === 'knitted'">
              <div class="mb-4">
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  尺码（重量/kg） <span class="text-red-500">*</span>
                </label>
                <input
                  v-model="knittedSizeWeight"
                  type="number"
                  min="0"
                  step="0.01"
                  placeholder="请输入尺码重量"
                  class="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:border-blue-500"
                >
              </div>
              <div class="mb-4">
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  胚布类型 <span class="text-red-500">*</span>
                </label>
                <input
                  v-model="knittedGreigeFabricType"
                  type="text"
                  placeholder="请输入胚布类型"
                  class="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:border-blue-500"
                >
              </div>
              <div class="mb-5">
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  胚布交期 <span class="text-red-500">*</span>
                </label>
                <input
                  v-model="knittedGreigeDeliveryDate"
                  type="date"
                  class="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:border-blue-500"
                >
              </div>
            </template>
          </div>

          <!-- Modal 底部 -->
          <div class="flex gap-3 p-5 border-t border-gray-100">
            <button
              @click="closeBidModal"
              class="flex-1 py-2.5 border border-gray-200 text-gray-600 rounded-lg font-medium hover:bg-gray-50 transition-colors"
            >
              取消
            </button>
            <button
              @click="submitBid"
              :disabled="bidSubmitting"
              class="flex-1 py-2.5 bg-blue-500 text-white rounded-lg font-medium hover:bg-blue-600 transition-colors disabled:bg-blue-300"
            >
              {{ bidSubmitting ? '提交中...' : '确认投标' }}
            </button>
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
