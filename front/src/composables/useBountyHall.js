import { ref, watch } from 'vue'
import { fetchBountyList, peekBountyList } from '@/api/bounty'
import { formatDateTime, formatDate, bountyTypeMap } from '@/utils/format'

// ========== 数据转换 ==========

const generateTagsFromData = (data) => {
  const tags = []
  if (data.bountyType) tags.push(bountyTypeMap[data.bountyType] || data.bountyType)
  if (data.sampleType) tags.push(data.sampleType)
  return tags
}

const generateDescriptionFromData = (data) => {
  const parts = []
  const formatComp = (composition) => {
    if (typeof composition === 'object') {
      return Object.entries(composition).map(([n, p]) => `${n} ${(p * 100).toFixed(0)}%`).join(' / ')
    }
    return composition
  }

  if (data.bountyType === 'woven' && data.wovenSpec) {
    const spec = data.wovenSpec
    if (spec.composition) parts.push(`成分: ${formatComp(spec.composition)}`)
    if (spec.fabricWeight) parts.push(`克重: ${spec.fabricWeight}g/m²`)
    if (spec.fabricWidth) parts.push(`幅宽: ${spec.fabricWidth}cm`)
    if (spec.quantityMeters) parts.push(`需求: ${spec.quantityMeters}米`)
  } else if (data.bountyType === 'knitted' && data.knittedSpec) {
    const spec = data.knittedSpec
    if (spec.composition) parts.push(`成分: ${formatComp(spec.composition)}`)
    if (spec.fabricWeight) parts.push(`克重: ${spec.fabricWeight}g/m²`)
    if (spec.fabricWidth) parts.push(`幅宽: ${spec.fabricWidth}cm`)
    if (spec.quantityKg) parts.push(`需求: ${spec.quantityKg}kg`)
  }
  return parts.length > 0 ? parts.join(' | ') : '暂无详细规格'
}

const transformBountyItem = (item) => {
  const id = item.id
  const productName = item.product_name || item.productName
  const productCode = item.product_code || item.productCode
  const createdAt = item.created_at || item.createdAt
  const bidDeadline = item.bid_deadline || item.bidDeadline
  const bountyType = item.bounty_type || item.bountyType
  const status = item.status
  const sampleType = item.sample_type || item.sampleType
  const expectedDeliveryDate = item.expected_delivery_date || item.expectedDeliveryDate

  let wovenSpec = item.wovenSpec
  let knittedSpec = item.knittedSpec

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

// ========== Composable ==========

export function useBountyHall(isLoggedIn) {
  // 筛选状态
  const searchKeyword = ref('')
  const filterBeginDate = ref('')
  const filterEndDate = ref('')
  const filterIncludeEnd = ref('0')

  // 列表数据
  const displayTasks = ref([])
  const loading = ref(false)
  const total = ref(0)
  const error = ref(null)

  // 加载悬赏列表（登录后）
  const loadBountyList = async () => {
    loading.value = true
    error.value = null
    try {
      const result = await fetchBountyList({
        keyword: searchKeyword.value,
        beginDate: filterBeginDate.value,
        endDate: filterEndDate.value,
        includeEnd: filterIncludeEnd.value,
      })
      displayTasks.value = result.data.map(transformBountyItem)
      total.value = result.total
    } catch (e) {
      console.error('加载悬赏列表失败:', e)
      error.value = e.message
      displayTasks.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }

  // 加载预览列表（未登录）
  const loadPeekBounties = async () => {
    loading.value = true
    error.value = null
    try {
      const result = await peekBountyList()
      displayTasks.value = result.data.map(transformBountyItem)
    } catch (e) {
      console.error('加载预览列表失败:', e)
      displayTasks.value = []
    } finally {
      loading.value = false
    }
  }

  // 清除筛选
  const clearFilters = () => {
    searchKeyword.value = ''
    filterBeginDate.value = ''
    filterEndDate.value = ''
    filterIncludeEnd.value = '0'
    if (isLoggedIn.value) loadBountyList()
  }

  // 是否有筛选条件
  const hasFilters = () => {
    return searchKeyword.value || filterBeginDate.value || filterEndDate.value || filterIncludeEnd.value === '1'
  }

  // 监听登录状态变化
  watch(isLoggedIn, (loggedIn) => {
    if (loggedIn) {
      clearFilters()
    } else {
      loadPeekBounties()
    }
  }, { immediate: false })

  // 初始加载
  const init = () => {
    if (isLoggedIn.value) {
      loadBountyList()
    } else {
      loadPeekBounties()
    }
  }

  return {
    // 筛选状态
    searchKeyword,
    filterBeginDate,
    filterEndDate,
    filterIncludeEnd,
    // 列表数据
    displayTasks,
    loading,
    total,
    error,
    // 方法
    loadBountyList,
    loadPeekBounties,
    clearFilters,
    hasFilters,
    init,
  }
}
