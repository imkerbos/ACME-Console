import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor - add token
api.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => Promise.reject(error)
)

// Response interceptor
api.interceptors.response.use(
  response => response.data,
  error => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    const message = error.response?.data?.message || error.message || 'Request failed'
    return Promise.reject(new Error(message))
  }
)

// Auth API
export const authApi = {
  login(username, password) {
    return api.post('/auth/login', { username, password })
  },

  getCurrentUser() {
    return api.get('/auth/me')
  },

  changePassword(oldPassword, newPassword) {
    return api.post('/auth/change-password', {
      old_password: oldPassword,
      new_password: newPassword
    })
  },

  updateProfile(data) {
    return api.put('/auth/profile', data)
  }
}

// User API (admin only)
export const userApi = {
  list(params = {}) {
    return api.get('/admin/users', { params })
  },

  get(id) {
    return api.get(`/admin/users/${id}`)
  },

  create(data) {
    return api.post('/admin/users', data)
  },

  update(id, data) {
    return api.put(`/admin/users/${id}`, data)
  },

  delete(id) {
    return api.delete(`/admin/users/${id}`)
  },

  resetPassword(id, password) {
    return api.post(`/admin/users/${id}/reset-password`, { password })
  }
}

// Certificate API
export const certificateApi = {
  list(params = {}) {
    return api.get('/certificates', { params })
  },

  get(id) {
    return api.get(`/certificates/${id}`)
  },

  create(data) {
    return api.post('/certificates', data)
  },

  delete(id) {
    return api.delete(`/certificates/${id}`)
  },

  verify(id) {
    return api.post(`/certificates/${id}/verify`, {}, { timeout: 120000 })
  },

  preVerify(id) {
    return api.post(`/certificates/${id}/pre-verify`)
  },

  download(id, format = 'zip', password = '') {
    const params = { format }
    if (password) params.password = password

    // Create a new axios instance without response interceptor for blob downloads
    const downloadApi = axios.create({
      baseURL: '/api/v1',
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json'
      }
    })

    // Add auth token
    downloadApi.interceptors.request.use(config => {
      const token = localStorage.getItem('token')
      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }
      return config
    })

    return downloadApi.get(`/certificates/${id}/download`, {
      params,
      responseType: 'blob'
    })
  },

  getChallenges(id) {
    return api.get(`/certificates/${id}/challenges`)
  },

  exportChallenges(id) {
    return api.get(`/certificates/${id}/challenges/export`, {
      responseType: 'text'
    })
  }
}

// Workspace API
export const workspaceApi = {
  list() {
    return api.get('/workspaces')
  },

  get(id) {
    return api.get(`/workspaces/${id}`)
  },

  create(data) {
    return api.post('/workspaces', data)
  },

  update(id, data) {
    return api.put(`/workspaces/${id}`, data)
  },

  delete(id) {
    return api.delete(`/workspaces/${id}`)
  },

  listMembers(id) {
    return api.get(`/workspaces/${id}/members`)
  },

  addMember(id, data) {
    return api.post(`/workspaces/${id}/members`, data)
  },

  updateMember(id, userId, data) {
    return api.put(`/workspaces/${id}/members/${userId}`, data)
  },

  removeMember(id, userId) {
    return api.delete(`/workspaces/${id}/members/${userId}`)
  }
}

// Notification API
export const notificationApi = {
  list(params = {}) {
    return api.get('/notifications', { params })
  },

  get(id) {
    return api.get(`/notifications/${id}`)
  },

  create(data) {
    return api.post('/notifications', data)
  },

  update(id, data) {
    return api.put(`/notifications/${id}`, data)
  },

  delete(id) {
    return api.delete(`/notifications/${id}`)
  },

  test(id) {
    return api.post(`/notifications/${id}/test`)
  },

  getLogs(certId, limit = 50) {
    return api.get(`/certificates/${certId}/notification-logs`, { params: { limit } })
  }
}

export default api
