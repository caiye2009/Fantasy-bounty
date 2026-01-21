import { getToken } from './auth'

const API_BASE = '/api/v1'

// 通用请求头
const getHeaders = () => ({
  'Content-Type': 'application/json',
  'Authorization': `Bearer ${getToken()}`
})

/**
 * 统一搜索接口
 * @param {Object} params - 搜索参数
 * @param {string} params.index - 索引名称 (如 "bounty")
 * @param {string} params.query - 关键词
 * @param {Object} params.filters - 筛选条件
 * @param {Object} params.sort - 排序配置 { field, order }
 * @param {number} params.page - 页码
 * @param {number} params.size - 每页数量
 * @returns {Promise<Object>}
 */
export const searchBounties = async (params) => {
  const response = await fetch(`${API_BASE}/search`, {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify({
      index: params.index || 'bounty',
      query: params.query || '',
      filters: params.filters || {},
      sort: params.sort || null,
      page: params.page || 1,
      size: params.size || 20
    })
  })

  if (!response.ok) {
    throw new Error(`Search failed: ${response.status}`)
  }

  const result = await response.json()

  if (result.code !== 200) {
    throw new Error(result.message || 'Search failed')
  }

  return result.data
}
