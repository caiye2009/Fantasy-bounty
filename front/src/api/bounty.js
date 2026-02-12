import { authFetch } from './request'

const API_BASE = '/api/v1'

/**
 * 获取悬赏列表（需要登录）
 */
export const fetchBountyList = async (params = {}) => {
  const query = new URLSearchParams()
  if (params.keyword) query.set('keyword', params.keyword)
  if (params.beginDate) query.set('begin_date', params.beginDate)
  if (params.endDate) query.set('end_date', params.endDate)
  if (params.includeEnd) query.set('include_end', params.includeEnd)

  const qs = query.toString()
  const response = await authFetch(
    `${API_BASE}/internal/bounties${qs ? '?' + qs : ''}`
  )

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`)
  }

  const result = await response.json()

  // 适配内部系统响应格式
  if (result.isSucceed && result.data) {
    return {
      data: Array.isArray(result.data) ? result.data : (result.data.list || []),
      total: result.data.total || (Array.isArray(result.data) ? result.data.length : 0)
    }
  }

  return {
    data: result.data || [],
    total: result.total || 0
  }
}

/**
 * 预览悬赏列表（公开接口，无需登录）
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
 */
export const fetchBountyDetail = async (id) => {
  const response = await authFetch(
    `${API_BASE}/internal/bounties/${id}`
  )

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`)
  }

  const result = await response.json()

  if (result.isSucceed && result.data) {
    return result.data
  }

  return result.data
}
