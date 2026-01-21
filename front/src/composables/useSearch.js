import { ref, reactive, computed } from 'vue'
import { searchBounties } from '@/api/search'

// 最小搜索字符数
const MIN_QUERY_LENGTH = 2

/**
 * 搜索状态管理 Composable
 * 统一管理 query / filters / sort，任意变化都通过 POST 发送
 */
export function useSearch() {
  // 搜索状态
  const searchState = reactive({
    query: '',
    filters: {},
    sort: {
      field: 'created_at',
      order: 'desc'
    },
    page: 1,
    size: 20
  })

  // 输入中的 query（用于显示，不一定触发搜索）
  const inputQuery = ref('')

  // 搜索结果
  const results = ref([])
  const total = ref(0)
  const totalPages = ref(0)
  const loading = ref(false)
  const error = ref(null)

  // 动态筛选项（从 ES 聚合返回）
  const availableFilters = ref([])

  // 计算属性
  const hasResults = computed(() => results.value.length > 0)
  const isEmpty = computed(() => !loading.value && results.value.length === 0)

  // 是否可以搜索（query 为空或 >= MIN_QUERY_LENGTH）
  const canSearch = computed(() => {
    const q = inputQuery.value.trim()
    return q.length === 0 || q.length >= MIN_QUERY_LENGTH
  })

  /**
   * 执行搜索
   * query/filters/sort 始终绑定在一起发送
   */
  async function doSearch() {
    // 检查 query 长度限制
    const trimmedQuery = inputQuery.value.trim()
    if (trimmedQuery.length > 0 && trimmedQuery.length < MIN_QUERY_LENGTH) {
      // query 不满足最小长度要求，不发送请求
      return
    }

    // 更新 searchState.query（只有满足条件时才更新）
    searchState.query = trimmedQuery

    loading.value = true
    error.value = null

    try {
      // query / filters / sort 始终一起发送
      const response = await searchBounties({
        index: 'bounty',
        query: searchState.query,
        filters: searchState.filters,
        sort: searchState.sort.field ? searchState.sort : null,
        page: searchState.page,
        size: searchState.size
      })

      results.value = response.hits || []
      total.value = response.total || 0
      totalPages.value = response.totalPages || 0
      availableFilters.value = response.filters || []
    } catch (e) {
      error.value = e.message
      results.value = []
      total.value = 0
      availableFilters.value = []
    } finally {
      loading.value = false
    }
  }

  /**
   * 更新输入的关键词（不立即触发搜索）
   */
  function setInputQuery(query) {
    inputQuery.value = query
    searchState.page = 1 // 重置页码
  }

  /**
   * 更新单个筛选条件
   */
  function setFilter(field, value) {
    if (value === null || value === undefined || value === '' ||
        (Array.isArray(value) && value.length === 0)) {
      delete searchState.filters[field]
    } else {
      searchState.filters[field] = value
    }
    searchState.page = 1 // 重置页码
  }

  /**
   * 批量设置筛选条件
   */
  function setFilters(filters) {
    searchState.filters = { ...filters }
    searchState.page = 1
  }

  /**
   * 清除所有筛选条件
   */
  function clearFilters() {
    searchState.filters = {}
    searchState.page = 1
  }

  /**
   * 设置排序
   */
  function setSort(field, order = 'desc') {
    searchState.sort = { field, order }
    searchState.page = 1
  }

  /**
   * 切换排序方向
   */
  function toggleSortOrder() {
    searchState.sort.order = searchState.sort.order === 'asc' ? 'desc' : 'asc'
  }

  /**
   * 设置页码
   */
  function setPage(page) {
    if (page >= 1 && (totalPages.value === 0 || page <= totalPages.value)) {
      searchState.page = page
    }
  }

  /**
   * 设置每页数量
   */
  function setPageSize(size) {
    searchState.size = size
    searchState.page = 1
  }

  /**
   * 重置所有搜索状态
   */
  function resetSearch() {
    inputQuery.value = ''
    searchState.query = ''
    searchState.filters = {}
    searchState.sort = { field: 'created_at', order: 'desc' }
    searchState.page = 1
    searchState.size = 20
  }

  return {
    // 状态
    searchState,
    inputQuery,
    results,
    total,
    totalPages,
    loading,
    error,
    availableFilters,

    // 常量
    MIN_QUERY_LENGTH,

    // 计算属性
    hasResults,
    isEmpty,

    // 方法
    doSearch,
    setQuery: setInputQuery,
    setFilter,
    setFilters,
    clearFilters,
    setSort,
    toggleSortOrder,
    setPage,
    setPageSize,
    resetSearch
  }
}
