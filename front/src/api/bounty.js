import { getToken } from './auth'

const API_BASE = '/api/v1'

// 通用请求头
const getHeaders = () => ({
  'Content-Type': 'application/json',
  'Authorization': `Bearer ${getToken()}`
})

/**
 * 获取悬赏列表（需要登录）
 * @param {number} page - 页码
 * @param {number} pageSize - 每页数量
 * @returns {Promise<{data: Array, total: number}>}
 */
export const fetchBountyList = async (page = 1, pageSize = 10) => {
  const response = await fetch(
    `${API_BASE}/bounties?page=${page}&page_size=${pageSize}`,
    { headers: getHeaders() }
  )

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
    `${API_BASE}/bounties/${id}`,
    { headers: getHeaders() }
  )

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`)
  }

  const result = await response.json()
  return result.data
}
