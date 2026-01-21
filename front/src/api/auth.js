const API_BASE = '/api/v1'

// Token 管理
const TOKEN_KEY = 'token'
const PHONE_KEY = 'user_phone'

export const getToken = () => localStorage.getItem(TOKEN_KEY)

export const setToken = (token) => localStorage.setItem(TOKEN_KEY, token)

export const removeToken = () => localStorage.removeItem(TOKEN_KEY)

export const getPhone = () => localStorage.getItem(PHONE_KEY)

export const setPhone = (phone) => localStorage.setItem(PHONE_KEY, phone)

export const removePhone = () => localStorage.removeItem(PHONE_KEY)

export const isAuthenticated = () => !!getToken()

// 模拟验证码（开发阶段使用）
const MOCK_CODE = '123'
const USE_MOCK = true // 切换为 false 时启用真实手机号验证码接口

/**
 * 发送验证码
 * @param {string} phone - 手机号
 * @returns {Promise<{success: boolean}>}
 */
export const sendVerifyCode = async (phone) => {
  if (USE_MOCK) {
    // 模拟发送验证码，延迟500ms
    await new Promise(resolve => setTimeout(resolve, 500))
    console.log(`[Mock] 验证码已发送到 ${phone}，验证码为: ${MOCK_CODE}`)
    return { success: true }
  }

  // 真实接口（后续启用）
  const response = await fetch(`${API_BASE}/auth/send-code`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ phone })
  })

  const result = await response.json()

  if (!response.ok) {
    throw new Error(result.message || '验证码发送失败')
  }

  return result
}

/**
 * 手机号+验证码登录（注册与登录合并）
 * @param {string} phone - 手机号
 * @param {string} code - 验证码
 * @returns {Promise<{token: string, phone: string}>}
 */
export const loginWithCode = async (phone, code) => {
  if (USE_MOCK) {
    // 模拟验证码校验
    if (code !== MOCK_CODE) {
      throw new Error('验证码错误')
    }

    // 调用后端真实登录接口获取有效 JWT token（开发阶段用 admin/admin）
    const response = await fetch(`${API_BASE}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: 'admin', password: 'admin' })
    })

    const result = await response.json()

    if (!response.ok) {
      throw new Error(result.message || '登录失败')
    }

    // 存储真实 token 和手机号
    setToken(result.token)
    setPhone(phone)
    return { token: result.token, phone }
  }

  // 真实手机号验证码登录接口（后续启用）
  const response = await fetch(`${API_BASE}/auth/login-with-code`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ phone, code })
  })

  const result = await response.json()

  if (!response.ok) {
    throw new Error(result.message || '登录失败')
  }

  // 存储 token 和手机号
  setToken(result.token)
  setPhone(phone)
  return result
}

/**
 * 用户登出
 */
export const logout = () => {
  removeToken()
  removePhone()
}
