import { reactive } from 'vue'
import { settingApi } from '../api'

const state = reactive({
  title: 'ACME Console',
  subtitle: '',
  loaded: false
})

let loadPromise = null

export const useSite = () => {
  const load = async () => {
    if (loadPromise) return loadPromise
    loadPromise = settingApi.getSite().then(res => {
      if (res.data?.title) state.title = res.data.title
      if (res.data?.subtitle) state.subtitle = res.data.subtitle
      state.loaded = true
      document.title = state.title
    }).catch(() => {
      // Fallback to defaults
      state.loaded = true
    })
    return loadPromise
  }

  const getTitle = () => state.title
  const getSubtitle = () => state.subtitle
  const isLoaded = () => state.loaded

  const update = (title, subtitle) => {
    if (title !== undefined) state.title = title
    if (subtitle !== undefined) state.subtitle = subtitle
    document.title = state.title
  }

  const reload = () => {
    loadPromise = null
    return load()
  }

  return { load, getTitle, getSubtitle, isLoaded, update, reload }
}
