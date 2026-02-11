import { getToken } from './auth'

const API_BASE = '/api/v1'

// 通用请求头
const getHeaders = () => ({
  'Content-Type': 'application/json',
  'Authorization': `Bearer ${getToken()}`
})

/**
 * 获取悬赏列表（需要登录）
 * @param {Object} params - 查询参数
 * @param {string} params.keyword - 产品名称搜索（模糊）
 * @param {string} params.beginDate - 发布开始时间
 * @param {string} params.endDate - 发布结束时间
 * @param {string} params.includeEnd - 是否包含已截止 '1'包含 '0'不包含
 * @returns {Promise<{data: Array, total: number}>}
 */
export const fetchBountyList = async (params = {}) => {
  const query = new URLSearchParams()
  if (params.keyword) query.set('keyword', params.keyword)
  if (params.beginDate) query.set('begin_date', params.beginDate)
  if (params.endDate) query.set('end_date', params.endDate)
  if (params.includeEnd) query.set('include_end', params.includeEnd)

  const qs = query.toString()
  const response = await fetch(
    `${API_BASE}/internal/bounties${qs ? '?' + qs : ''}`,
    { headers: getHeaders() }
  )

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`)
  }

  const result = await response.json()

  // 适配内部系统响应格式
  if (result.isSucceed && result.data) {
    // 内部系统返回的数据格式
    return {
      data: Array.isArray(result.data) ? result.data : (result.data.list || []),
      total: result.data.total || (Array.isArray(result.data) ? result.data.length : 0)
    }
  }

  // 原有格式兼容
  return {
    data: result.data || [],
    total: result.total || 0
  }
}

/**
 * 预览悬赏列表（公开接口，无需登录）
 * 仅返回第一页10条数据，不支持筛选
 * @returns {Promise<{data: Array, total: number}>}
 */
export const peekBountyList = async () => {
  const response = await fetch(`${API_BASE}/bounties/peek`, {
    headers: { 'Content-Type': 'application/json' }
  })

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`)
  }

  const result = await response.json()
  return {
    data: result.data || [],
    total: result.total || 0
  }
}

/**
 * 获取悬赏详情
 * @param {number} id - 悬赏ID
 * @returns {Promise<Object>}
 */
export const fetchBountyDetail = async (id) => {
  const response = await fetch(
    `${API_BASE}/internal/bounties/${id}`,
    { headers: getHeaders() }
  )

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`)
  }

  const result = await response.json()

  // 适配内部系统响应格式
  if (result.isSucceed && result.data) {
    return result.data
  }

  return result.data
}
