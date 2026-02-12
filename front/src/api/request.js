import { getToken, setToken, removeToken, removePhone, removeUsername } from './token'

const API_BASE = '/api/v1'

let isRefreshing = false
let pendingRequests = []

// 刷新 token
const refreshToken = async () => {
  const token = getToken()
  if (!token) return false

  try {
    const response = await fetch(`${API_BASE}/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    })

    if (!response.ok) return false

    const result = await response.json()
    if (result.token) {
      setToken(result.token)
      return true
    }
    return false
  } catch {
    return false
  }
}

// 强制登出
const forceLogout = () => {
  removeToken()
  removePhone()
  removeUsername()
  window.location.reload()
}

/**
 * 带自动 token 刷新的 fetch 封装
 * 遇到 401 自动尝试 refresh，成功后重试原请求
 */
export const authFetch = async (url, options = {}) => {
  const token = getToken()

  // 设置默认 headers
  const headers = { ...options.headers }

  // 只在没有显式设置时添加 Content-Type（FormData 不应设置）
  if (!options.body || !(options.body instanceof FormData)) {
    if (!headers['Content-Type']) {
      headers['Content-Type'] = 'application/json'
    }
  }
  // 如果显式传 undefined 表示不要设置
  if (headers['Content-Type'] === undefined) {
    delete headers['Content-Type']
  }

  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const response = await fetch(url, { ...options, headers })

  // 非 401 直接返回
  if (response.status !== 401) {
    return response
  }

  // 401：尝试刷新 token
  if (!isRefreshing) {
    isRefreshing = true

    const success = await refreshToken()
    isRefreshing = false

    if (success) {
      // 刷新成功，重试所有等待中的请求
      pendingRequests.forEach(cb => cb())
      pendingRequests = []

      // 重试当前请求
      const newToken = getToken()
      headers['Authorization'] = `Bearer ${newToken}`
      return fetch(url, { ...options, headers })
    } else {
      // 刷新失败，清除登录状态
      pendingRequests = []
      forceLogout()
      return response
    }
  }

  // 已经在刷新中，排队等待
  return new Promise((resolve) => {
    pendingRequests.push(() => {
      const newToken = getToken()
      headers['Authorization'] = `Bearer ${newToken}`
      resolve(fetch(url, { ...options, headers }))
    })
  })
}
