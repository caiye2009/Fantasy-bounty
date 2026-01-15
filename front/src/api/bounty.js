const API_BASE = '/api/v1'

/**
 * 获取悬赏列表
 * @param {number} page - 页码
 * @param {number} pageSize - 每页数量
 * @returns {Promise<{data: Array, total: number}>}
 */
export async function fetchBountyList(page = 1, pageSize = 10) {
  const response = await fetch(
    `${API_BASE}/bounties?page=${page}&page_size=${pageSize}`
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
 * 获取悬赏详情
 * @param {number} id - 悬赏ID
 * @returns {Promise<Object>}
 */
export async function fetchBountyDetail(id) {
  const response = await fetch(`${API_BASE}/bounties/${id}`)

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`)
  }

  const result = await response.json()
  return result.data
}
