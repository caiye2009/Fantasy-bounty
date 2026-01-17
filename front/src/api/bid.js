import { getToken } from './auth'

const API_BASE = '/api/v1'

// 通用请求头
const getHeaders = () => ({
  'Content-Type': 'application/json',
  'Authorization': `Bearer ${getToken()}`
})

/**
 * 获取竞标列表
 * @param {number} bountyId - 赏金ID
 * @param {number} page - 页码
 * @param {number} pageSize - 每页数量
 * @returns {Promise<{data: Array, total: number}>}
 */
export const fetchBidList = async (bountyId, page = 1, pageSize = 10) => {
  const response = await fetch(
    `${API_BASE}/bids?bounty_id=${bountyId}&page=${page}&page_size=${pageSize}`,
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
 * 获取我的投标列表
 * @param {string} status - 状态筛选: pending, accepted, rejected, completed
 * @param {number} page - 页码
 * @param {number} pageSize - 每页数量
 * @returns {Promise<{data: Array, total: number}>}
 */
export const fetchMyBids = async (status = '', page = 1, pageSize = 10) => {
  let url = `${API_BASE}/bids/my?page=${page}&page_size=${pageSize}`
  if (status) {
    url += `&status=${status}`
  }

  const response = await fetch(url, { headers: getHeaders() })

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
 * @param {Object} bidData - 投标数据
 * @param {number} bidData.bountyId - 悬赏ID
 * @param {number} bidData.bidPrice - 投标价格
 * @param {Object} [bidData.wovenSpec] - 梭织规格（梭织类型时必填）
 * @param {number} [bidData.wovenSpec.sizeLength] - 尺码（长度）
 * @param {string} [bidData.wovenSpec.greigeFabricType] - 胚布类型
 * @param {string} [bidData.wovenSpec.greigeDeliveryDate] - 胚布交期
 * @param {Object} [bidData.knittedSpec] - 针织规格（针织类型时必填）
 * @param {number} [bidData.knittedSpec.sizeWeight] - 尺码（重量/皮重）
 * @param {string} [bidData.knittedSpec.greigeFabricType] - 胚布类型
 * @param {string} [bidData.knittedSpec.greigeDeliveryDate] - 胚布交期
 * @returns {Promise<Object>}
 */
export const placeBid = async (bidData) => {
  const response = await fetch(`${API_BASE}/bids`, {
    method: 'POST',
    headers: getHeaders(),
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
 * @param {string} bidId - 竞标ID (UUID)
 * @returns {Promise<void>}
 */
export const deleteBid = async (bidId) => {
  const response = await fetch(`${API_BASE}/bids/${bidId}`, {
    method: 'DELETE',
    headers: getHeaders()
  })

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`)
  }
}
