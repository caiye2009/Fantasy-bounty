import { authFetch } from './request'
import {
  getToken, setToken, removeToken,
  getPhone, setPhone, removePhone,
  getUsername, setUsername, removeUsername,
  isAuthenticated,
} from './token'

// 重新导出 token 操作，保持外部引用不变
export {
  getToken, setToken, removeToken,
  getPhone, setPhone, removePhone,
  getUsername, setUsername, removeUsername,
  isAuthenticated,
}

const API_BASE = '/api/v1'

// 模拟验证码（开发阶段使用）
const MOCK_CODE = '123'
const USE_MOCK = false // 切换为 false 时启用真实手机号验证码接口

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

    // 存储 token 和手机号
    setToken('mock-token')
    setPhone(phone)
    return { token: 'mock-token', phone }
  }

  // 真实手机号验证码登录接口
  const response = await fetch(`${API_BASE}/auth/verify-code`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ phone, code })
  })

  const result = await response.json()

  if (!response.ok) {
    throw new Error(result.message || '登录失败')
  }

  // 存储 token、手机号、用户名
  setToken(result.token)
  setPhone(phone)
  if (result.username) setUsername(result.username)
  return result
}

/**
 * 用户登出
 */
export const logout = () => {
  removeToken()
  removePhone()
  removeUsername()
}

// ========== 企业认证相关 API ==========

/**
 * 获取我的企业认证状态
 * @returns {Promise<{hasVerifiedCompany: boolean, company?: object, pendingApplication?: object, latestRejected?: object}>}
 */
export const getMyCompanyStatus = async () => {
  const token = getToken()
  if (!token) {
    throw new Error('未登录')
  }

  const response = await authFetch(`${API_BASE}/companies/my`, {
    method: 'GET',
  })

  const result = await response.json()

  if (!response.ok) {
    throw new Error(result.message || '获取企业认证状态失败')
  }

  return result.data
}

/**
 * 上传营业执照图片进行OCR识别
 * @param {File} file - 营业执照图片文件
 * @returns {Promise<{companyName: string, businessLicenseNo: string, legalPerson: string, ...}>}
 */
export const recognizeLicense = async (file) => {
  const token = getToken()
  if (!token) {
    throw new Error('未登录')
  }

  const formData = new FormData()
  formData.append('license', file)

  const response = await authFetch(`${API_BASE}/companies/recognize`, {
    method: 'POST',
    body: formData
  })

  const result = await response.json()

  if (!response.ok) {
    throw new Error(result.message || 'OCR识别失败')
  }

  return result
}

/**
 * 提交企业认证申请（图片已在OCR识别阶段上传）
 * @param {object} data - { name, businessLicenseNo, imagePath }
 * @returns {Promise<object>}
 */
export const applyCompany = async (data) => {
  const token = getToken()
  if (!token) {
    throw new Error('未登录')
  }

  const response = await authFetch(`${API_BASE}/companies/apply`, {
    method: 'POST',
    body: JSON.stringify(data)
  })

  const result = await response.json()

  if (!response.ok) {
    throw new Error(result.message || '提交认证申请失败')
  }

  return result
}
