// Token 存储（纯 localStorage 操作，无依赖）
const TOKEN_KEY = 'token'
const PHONE_KEY = 'user_phone'
const USERNAME_KEY = 'user_username'

export const getToken = () => localStorage.getItem(TOKEN_KEY)
export const setToken = (token) => localStorage.setItem(TOKEN_KEY, token)
export const removeToken = () => localStorage.removeItem(TOKEN_KEY)

export const getPhone = () => localStorage.getItem(PHONE_KEY)
export const setPhone = (phone) => localStorage.setItem(PHONE_KEY, phone)
export const removePhone = () => localStorage.removeItem(PHONE_KEY)

export const getUsername = () => localStorage.getItem(USERNAME_KEY)
export const setUsername = (username) => localStorage.setItem(USERNAME_KEY, username)
export const removeUsername = () => localStorage.removeItem(USERNAME_KEY)

export const isAuthenticated = () => !!getToken()
