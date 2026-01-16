const API_BASE = '/api/v1'

/**
 * 获取竞标列表
 * @param {number} bountyId - 赏金ID
 * @param {number} page - 页码
 * @param {number} pageSize - 每页数量
 * @returns {Promise<{data: Array, total: number}>}
 */
export const fetchBidList = async (bountyId, page = 1, pageSize = 10) => {
  const response = await fetch(
    `${API_BASE}/bids?bounty_id=${bountyId}&page=${page}&page_size=${pageSize}`
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
 * 创建竞标
 * @param {number} bountyId - 赏金ID
 * @param {number} price - 竞标价格
 * @returns {Promise<Object>}
 */
export const placeBid = async (bountyId, price) => {
  const response = await fetch(`${API_BASE}/bids`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      bounty_id: bountyId,
      price: price
    })
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
    method: 'DELETE'
  })

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`)
  }
}