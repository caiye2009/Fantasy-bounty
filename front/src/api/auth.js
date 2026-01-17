const API_BASE = '/api/v1'

// Token 管理
const TOKEN_KEY = 'token'

export const getToken = () => localStorage.getItem(TOKEN_KEY)

export const setToken = (token) => localStorage.setItem(TOKEN_KEY, token)

export const removeToken = () => localStorage.removeItem(TOKEN_KEY)

export const isAuthenticated = () => !!getToken()

/**
 * 用户登录
 * @param {string} username - 用户名
 * @param {string} password - 密码
 * @returns {Promise<{token: string}>}
 */
export const login = async (username, password) => {
  const response = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password })
  })

  const result = await response.json()

  if (!response.ok) {
    throw new Error(result.message || '登录失败')
  }

  // 存储 token
  setToken(result.token)
  return result
}

/**
 * 用户登出
 */
export const logout = () => {
  removeToken()
}
