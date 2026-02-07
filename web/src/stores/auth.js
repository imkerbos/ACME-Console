import { reactive } from 'vue'
import { authApi } from '../api'

const state = reactive({
  token: localStorage.getItem('token') || null,
  user: JSON.parse(localStorage.getItem('user') || 'null')
})

export const useAuth = () => {
  const isAuthenticated = () => !!state.token

  const isAdmin = () => state.user?.role === 'admin'

  const login = async (username, password) => {
    const response = await authApi.login(username, password)
    const { token, user } = response.data

    state.token = token
    state.user = user

    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(user))

    return user
  }

  const logout = () => {
    state.token = null
    state.user = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  const getUser = () => state.user

  const setUser = (user) => {
    state.user = user
    localStorage.setItem('user', JSON.stringify(user))
  }

  const getToken = () => state.token

  const getRole = () => state.user?.role || 'user'

  return {
    isAuthenticated,
    isAdmin,
    login,
    logout,
    getUser,
    setUser,
    getToken,
    getRole
  }
}
