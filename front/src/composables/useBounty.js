import { ref, computed } from 'vue'
import { fetchBountyList } from '../api/bounty'

// 状态
const bounties = ref([])
const loading = ref(false)
const error = ref(null)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 计算属性
const totalPages = computed(() => Math.ceil(total.value / pageSize.value))
const isEmpty = computed(() => bounties.value.length === 0 && !loading.value)

/**
 * 加载悬赏列表
 */
async function loadBounties(page = 1) {
  loading.value = true
  error.value = null

  try {
    const result = await fetchBountyList(page, pageSize.value)
    bounties.value = result.data
    total.value = result.total
    currentPage.value = page
  } catch (e) {
    error.value = e.message
    bounties.value = []
  } finally {
    loading.value = false
  }
}

/**
 * 刷新列表
 */
function refresh() {
  return loadBounties(currentPage.value)
}

/**
 * 切换页码
 */
function goToPage(page) {
  if (page >= 1 && page <= totalPages.value) {
    return loadBounties(page)
  }
}

export function useBounty() {
  return {
    // 状态
    bounties,
    loading,
    error,
    currentPage,
    pageSize,
    total,
    // 计算属性
    totalPages,
    isEmpty,
    // 方法
    loadBounties,
    refresh,
    goToPage
  }
}
