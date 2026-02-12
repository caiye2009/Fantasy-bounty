import { authFetch } from './request'

const API_BASE = '/api/v1'

/**
 * 获取竞标列表
 */
export const fetchBidList = async (bountyId, page = 1, pageSize = 10) => {
  const response = await authFetch(
    `${API_BASE}/supplier/bids?bounty_id=${bountyId}&page=${page}&page_size=${pageSize}`
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
 * 获取我的投标列表
 */
export const fetchMyBids = async (status = '', page = 1, pageSize = 10) => {
  let url = `${API_BASE}/bids/my?page=${page}&page_size=${pageSize}`
  if (status) {
    url += `&status=${status}`
  }

  const response = await authFetch(url)

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
 * 创建竞标
 */
export const placeBid = async (bidData) => {
  const response = await authFetch(`${API_BASE}/bids`, {
    method: 'POST',
    body: JSON.stringify(bidData)
  })

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`)
  }

  const result = await response.json()
  return result.data
}

/**
 * 删除竞标
 */
export const deleteBid = async (bidId) => {
  const response = await authFetch(`${API_BASE}/bids/${bidId}`, {
    method: 'DELETE'
  })

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`)
  }
}
