import { authFetch } from './request'

const API_BASE = '/api/v1'

/**
 * 统一搜索接口
 */
export const searchBounties = async (params) => {
  const response = await authFetch(`${API_BASE}/search`, {
    method: 'POST',
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
