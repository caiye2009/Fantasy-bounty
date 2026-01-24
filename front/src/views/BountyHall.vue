<script setup>
import { ref, onMounted, onUnmounted, inject, watch, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, ArrowUpNarrowWide, ArrowDownNarrowWide, Phone, Mail, Loader2, Inbox, X, Calendar, Truck, Tag } from 'lucide-vue-next'
import BountyCard from '@/components/BountyCard.vue'
import { placeBid } from '@/api/bid'
import { peekBountyList } from '@/api/bounty'
import { useSearch } from '@/composables/useSearch'

// 从父组件注入登录状态和登录modal控制
const isLoggedIn = inject('isLoggedIn')
const openLoginModal = inject('openLoginModal')

// 使用 ES 搜索
const {
  searchState,
  results: esResults,
  total: esTotal,
  totalPages,
  loading: esLoading,
  availableFilters,
  doSearch,
  setQuery,
  setFilter,
  clearFilters,
  setSort,
  toggleSortOrder: toggleEsSortOrder,
  setPage
} = useSearch()

// 需要登录才能执行的操作
const requireLogin = (callback) => {
  if (!isLoggedIn.value) {
    openLoginModal()
    return false
  }
  if (callback) callback()
  return true
}

// 排序选项（静态）
const sortOptions = [
  { label: '发布时间', value: 'created_at' },
  { label: '截止时间', value: 'bid_deadline' },
  { label: '克重', value: 'fabric_weight' }
]

// 从 ES 聚合结果获取动态筛选选项
const getFilterOptions = (field) => {
  const filter = availableFilters.value.find(f => f.field === field)
  if (!filter || !filter.buckets) {
    return [{ label: '全部', value: '' }]
  }
  return [
    { label: '全部', value: '' },
    ...filter.buckets.map(b => ({
      label: `${b.label} (${b.docCount})`,
      value: b.key
    }))
  ]
}

// 计算属性：动态筛选选项
const bountyTypeOptions = computed(() => getFilterOptions('bounty_type'))
const createdAtOptions = computed(() => getFilterOptions('created_at_ranges'))
const bidDeadlineOptions = computed(() => getFilterOptions('bid_deadline_ranges'))
const quantityMetersOptions = computed(() => getFilterOptions('quantity_meters_ranges'))
const quantityKgOptions = computed(() => getFilterOptions('quantity_kg_ranges'))

// 切换排序方向
const toggleSortOrder = () => {
  if (!requireLogin()) return
  toggleEsSortOrder()
  doSearch()
}

// 处理排序字段变化（下拉框）
const handleSortFieldChange = (value) => {
  if (!requireLogin()) return
  setSort(value)
  doSearch()
}

// 处理筛选器变化
const handleFilterChange = (field, value) => {
  if (!requireLogin()) return
  setFilter(field, value || null)
  doSearch()
}

// 清除所有筛选
const handleClearFilters = () => {
  if (!requireLogin()) return
  clearFilters()
  doSearch()
}

// 下拉框展开前检查登录
const onSelectVisibleChange = (visible) => {
  if (visible && !isLoggedIn.value) {
    openLoginModal()
  }
}

// 点击搜索框时检查登录
const onFilterClick = (event) => {
  if (!isLoggedIn.value) {
    event.preventDefault()
    event.stopPropagation()
    openLoginModal()
  }
}

// 任务列表（未登录时用 peek API）
const peekTasks = ref([])
const peekLoading = ref(false)

// 悬赏类型中英文对照
const bountyTypeMap = {
  woven: '梭织',
  knitted: '针织'
}

// 状态中英文对照
// const statusMap = {
//   open: '招标中',
//   in_progress: '进行中',
//   completed: '已完成',
//   closed: '已关闭'
// }

// 加载预览列表（未登录时使用）
const loadPeekBounties = async () => {
  peekLoading.value = true
  try {
    const result = await peekBountyList()
    peekTasks.value = result.data.map(item => transformBountyItem(item))
  } catch (error) {
    console.error('加载悬赏列表失败:', error)
  } finally {
    peekLoading.value = false
  }
}

// 转换悬赏数据格式（统一处理 ES 返回和 API 返回的数据）
const transformBountyItem = (item) => {
  // ES 返回的字段是 snake_case，API 返回的是 camelCase
  const id = item.id
  const productName = item.product_name || item.productName
  const productCode = item.product_code || item.productCode
  const createdAt = item.created_at || item.createdAt
  const bidDeadline = item.bid_deadline || item.bidDeadline
  const bountyType = item.bounty_type || item.bountyType
  const status = item.status
  const sampleType = item.sample_type || item.sampleType
  const expectedDeliveryDate = item.expected_delivery_date || item.expectedDeliveryDate

  // 重建 spec 对象（ES 返回的是扁平化的字段）
  let wovenSpec = item.wovenSpec
  let knittedSpec = item.knittedSpec

  // 如果没有嵌套对象，从扁平字段构建
  if (!wovenSpec && !knittedSpec) {
    if (bountyType === 'woven') {
      wovenSpec = {
        composition: item.composition,
        fabricWeight: item.fabric_weight,
        fabricWidth: item.fabric_width,
        warpDensity: item.warp_density,
        weftDensity: item.weft_density,
        warpMaterial: item.warp_material,
        weftMaterial: item.weft_material,
        quantityMeters: item.quantity_meters
      }
    } else if (bountyType === 'knitted') {
      knittedSpec = {
        composition: item.composition,
        fabricWeight: item.fabric_weight,
        fabricWidth: item.fabric_width,
        machineType: item.machine_type,
        materials: item.materials,
        quantityKg: item.quantity_kg
      }
    }
  }

  return {
    id,
    title: productName,
    productCode,
    publishTime: formatDateTime(createdAt),
    tags: generateTagsFromData({ bountyType, sampleType, status }),
    description: generateDescriptionFromData({ bountyType, wovenSpec, knittedSpec }),
    deadline: formatDate(bidDeadline),
    bountyType,
    status,
    wovenSpec,
    knittedSpec,
    expectedDeliveryDate
  }
}

// 从数据生成标签
const generateTagsFromData = (data) => {
  const tags = []
  if (data.bountyType) tags.push(bountyTypeMap[data.bountyType] || data.bountyType)
  if (data.sampleType) tags.push(data.sampleType)
  // if (data.status) tags.push(statusMap[data.status] || data.status)
  return tags
}

// 从数据生成描述
const generateDescriptionFromData = (data) => {
  const parts = []
  if (data.bountyType === 'woven' && data.wovenSpec) {
    const spec = data.wovenSpec
    if (spec.composition) {
      // 处理对象或字符串格式的成分
      const compStr = typeof spec.composition === 'object'
        ? Object.entries(spec.composition).map(([n, p]) => `${n} ${(p * 100).toFixed(0)}%`).join(' / ')
        : spec.composition
      parts.push(`成分: ${compStr}`)
    }
    if (spec.fabricWeight) parts.push(`克重: ${spec.fabricWeight}g/m²`)
    if (spec.fabricWidth) parts.push(`幅宽: ${spec.fabricWidth}cm`)
    if (spec.quantityMeters) parts.push(`需求: ${spec.quantityMeters}米`)
  } else if (data.bountyType === 'knitted' && data.knittedSpec) {
    const spec = data.knittedSpec
    if (spec.composition) {
      const compStr = typeof spec.composition === 'object'
        ? Object.entries(spec.composition).map(([n, p]) => `${n} ${(p * 100).toFixed(0)}%`).join(' / ')
        : spec.composition
      parts.push(`成分: ${compStr}`)
    }
    if (spec.fabricWeight) parts.push(`克重: ${spec.fabricWeight}g/m²`)
    if (spec.fabricWidth) parts.push(`幅宽: ${spec.fabricWidth}cm`)
    if (spec.quantityKg) parts.push(`需求: ${spec.quantityKg}kg`)
  }
  return parts.length > 0 ? parts.join(' | ') : '暂无详细规格'
}

// 计算当前显示的任务列表
const displayTasks = ref([])

// 监听登录状态变化
watch(isLoggedIn, (loggedIn) => {
  if (loggedIn) {
    // 登录后执行 ES 搜索
    doSearch()
  } else {
    // 登出后显示 peek 预览列表
    loadPeekBounties()
  }
}, { immediate: false })

// 监听 ES 搜索结果变化，转换数据格式
watch(esResults, (results) => {
  if (isLoggedIn.value && results) {
    displayTasks.value = results.map(item => transformBountyItem(item))
  }
}, { deep: true })

// 监听预览数据变化
watch(peekTasks, (tasks) => {
  if (!isLoggedIn.value) {
    displayTasks.value = tasks
  }
}, { deep: true })

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

// 格式化面料成分
// 处理对象格式 {"棉": 0.6, "涤纶": 0.4} -> "棉 60% / 涤纶 40%"
// 如果已经是字符串则直接返回
const formatComposition = (composition) => {
  if (!composition) return ''
  if (typeof composition === 'string') return composition
  if (typeof composition === 'object') {
    return Object.entries(composition)
      .map(([name, pct]) => {
        const percent = pct * 100
        return `${name} ${percent % 1 === 0 ? percent.toFixed(0) : percent.toFixed(1)}%`
      })
      .join(' / ')
  }
  return ''
}

// ESC 键关闭弹窗/抽屉（Modal 优先于抽屉）
const handleKeydown = (e) => {
  if (e.key === 'Escape') {
    if (bidModalVisible.value) {
      closeBidModal()
    } else if (drawerVisible.value) {
      closeDrawer()
    }
  }
}

onMounted(() => {
  if (isLoggedIn.value) {
    doSearch()
  } else {
    // 未登录时显示 peek 预览列表
    loadPeekBounties()
  }
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})

// 搜索输入防抖
let searchTimeout = null
const handleSearchInput = (event) => {
  if (!requireLogin()) return
  const value = event.target.value
  setQuery(value)

  // 防抖搜索
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    doSearch()
  }, 300)
}

// 按回车立即搜索
const handleSearch = () => {
  if (!requireLogin()) return
  if (searchTimeout) clearTimeout(searchTimeout)
  doSearch()
}

// 抽屉状态
const drawerVisible = ref(false)
const selectedTask = ref(null)

const openDrawer = (task) => {
  // 查看详情需要登录
  if (!requireLogin()) return
  selectedTask.value = task
  drawerVisible.value = true
  // 禁用页面滚动
  document.body.style.overflow = 'hidden'
}

const closeDrawer = () => {
  drawerVisible.value = false
  selectedTask.value = null
  // 恢复页面滚动
  document.body.style.overflow = ''
}

// 投标 Modal 状态
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
  // 重置梭织字段
  wovenSizeLength.value = ''
  wovenGreigeFabricType.value = ''
  wovenGreigeDeliveryDate.value = ''
  wovenDeliveryMethod.value = ''
  // 重置针织字段
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

  // 验证必填字段
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
  } catch (error) {
    console.error('投标失败:', error)
    ElMessage.error('投标失败，请重试')
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
      <aside class="w-[30%] shrink-0 sticky top-24 h-[calc(100vh-7rem)] flex flex-col justify-between overflow-y-auto">
        <!-- 搜索和排序 -->
        <div class="bg-white rounded-lg p-6 shadow-sm mb-4">
          <!-- 搜索框 -->
          <div class="relative mb-4">
            <input
              :value="searchState.query"
              type="text"
              placeholder="搜索悬赏任务..."
              class="w-full px-4 py-2.5 pl-10 border border-gray-200 rounded-lg text-sm focus:outline-none focus:border-blue-500"
              @mousedown="onFilterClick"
              @input="handleSearchInput"
              @keyup.enter="handleSearch"
            >
            <Search class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400" :size="14" />
          </div>

          <!-- 筛选器 -->
          <div class="mt-4 space-y-3">
            <!-- 悬赏类型 -->
            <div class="flex items-center gap-2">
              <span class="text-sm text-gray-600 whitespace-nowrap w-16">类型</span>
              <el-select
                :model-value="searchState.filters.bounty_type || ''"
                @change="handleFilterChange('bounty_type', $event)"
                @visible-change="onSelectVisibleChange"
                placeholder="全部"
                class="flex-1"
              >
                <el-option v-for="opt in bountyTypeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
              </el-select>
            </div>

            <!-- 发布时间 -->
            <div v-if="createdAtOptions.length > 1" class="flex items-center gap-2">
              <span class="text-sm text-gray-600 whitespace-nowrap w-16">发布</span>
              <el-select
                :model-value="searchState.filters.created_at || ''"
                @change="handleFilterChange('created_at', $event)"
                @visible-change="onSelectVisibleChange"
                placeholder="全部"
                class="flex-1"
              >
                <el-option v-for="opt in createdAtOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
              </el-select>
            </div>

            <!-- 截止接单 -->
            <div v-if="bidDeadlineOptions.length > 1" class="flex items-center gap-2">
              <span class="text-sm text-gray-600 whitespace-nowrap w-16">截止</span>
              <el-select
                :model-value="searchState.filters.bid_deadline || ''"
                @change="handleFilterChange('bid_deadline', $event)"
                @visible-change="onSelectVisibleChange"
                placeholder="全部"
                class="flex-1"
              >
                <el-option v-for="opt in bidDeadlineOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
              </el-select>
            </div>

            <!-- 需求量(米) - 梭织 -->
            <div v-if="quantityMetersOptions.length > 1" class="flex items-center gap-2">
              <span class="text-sm text-gray-600 whitespace-nowrap w-16">需求(米)</span>
              <el-select
                :model-value="searchState.filters.quantity_meters || ''"
                @change="handleFilterChange('quantity_meters', $event)"
                @visible-change="onSelectVisibleChange"
                placeholder="全部"
                class="flex-1"
              >
                <el-option v-for="opt in quantityMetersOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
              </el-select>
            </div>

            <!-- 需求量(kg) - 针织 -->
            <div v-if="quantityKgOptions.length > 1" class="flex items-center gap-2">
              <span class="text-sm text-gray-600 whitespace-nowrap w-16">需求(kg)</span>
              <el-select
                :model-value="searchState.filters.quantity_kg || ''"
                @change="handleFilterChange('quantity_kg', $event)"
                @visible-change="onSelectVisibleChange"
                placeholder="全部"
                class="flex-1"
              >
                <el-option v-for="opt in quantityKgOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
              </el-select>
            </div>
          </div>

           <!-- 排序方式 -->
          <div class="flex items-center gap-2">
            <span class="text-sm text-gray-600 whitespace-nowrap">排序</span>
            <el-select
              :model-value="searchState.sort.field"
              @change="handleSortFieldChange"
              @visible-change="onSelectVisibleChange"
              class="flex-1"
            >
              <el-option v-for="opt in sortOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
            </el-select>
            <el-button
              @click="toggleSortOrder"
              :title="searchState.sort.order === 'asc' ? '升序' : '降序'"
            >
              <ArrowUpNarrowWide v-if="searchState.sort.order === 'asc'" :size="14" />
              <ArrowDownNarrowWide v-else :size="14" />
            </el-button>
          </div>

          <!-- 搜索统计 -->
          <div v-if="isLoggedIn" class="mt-4 text-xs text-gray-400">
            共 {{ esTotal }} 条结果
            <span v-if="Object.keys(searchState.filters).length > 0" class="ml-2">
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
        <div v-if="esLoading || peekLoading" class="bg-white rounded-lg p-16 shadow-sm text-center">
          <Loader2 class="text-gray-300 mb-4 animate-spin mx-auto" :size="40" />
          <p class="text-gray-500">加载中...</p>
        </div>

        <!-- 空状态 -->
        <div v-else-if="displayTasks.length === 0" class="bg-white rounded-lg p-16 shadow-sm text-center">
          <Inbox class="text-gray-300 mb-6 mx-auto" :size="60" />
          <h3 class="text-lg font-medium text-gray-600 mb-2">暂无悬赏任务</h3>
          <p class="text-sm text-gray-400">
            {{ isLoggedIn && searchState.query ? '未找到匹配的结果，请尝试其他关键词' : '请稍后再来查看' }}
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

        <!-- 分页（登录后显示） -->
        <div v-if="isLoggedIn && totalPages > 1" class="flex justify-center py-4">
          <el-pagination
            :current-page="searchState.page"
            :page-count="totalPages"
            layout="prev, pager, next"
            @current-change="(page) => { setPage(page); doSearch() }"
          />
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
