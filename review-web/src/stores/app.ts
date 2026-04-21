import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { useRoute, useRouter } from 'vue-router'

export type AppRole = 'consumer' | 'merchant' | 'operator'
export interface MockSession {
  role: AppRole
  displayName: string
  identity: string
}

const SESSION_STORAGE_KEY = 'review-web.mock-session'

function canUseStorage() {
  return typeof window !== 'undefined' && typeof window.localStorage !== 'undefined'
}

export function loadMockSession(): MockSession | null {
  if (!canUseStorage()) {
    return null
  }

  const raw = window.localStorage.getItem(SESSION_STORAGE_KEY)
  if (!raw) {
    return null
  }

  try {
    return JSON.parse(raw) as MockSession
  } catch {
    return null
  }
}

function persistMockSession(session: MockSession | null) {
  if (!canUseStorage()) {
    return
  }

  if (session) {
    window.localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(session))
    return
  }

  window.localStorage.removeItem(SESSION_STORAGE_KEY)
}

export const useAppStore = defineStore('app', () => {
  const reviewServiceBaseUrl = ref('http://localhost:8000')
  const businessServiceBaseUrl = ref('http://localhost:8010')
  const operationServiceBaseUrl = ref('http://localhost:8020')
  const currentSession = ref<MockSession | null>(loadMockSession())

  const serviceMap = computed(() => ({
    review: reviewServiceBaseUrl.value,
    business: businessServiceBaseUrl.value,
    operation: operationServiceBaseUrl.value,
  }))
  const isLoggedIn = computed(() => currentSession.value !== null)

  function getServiceBaseUrl(service: 'review' | 'business' | 'operation') {
    return serviceMap.value[service]
  }

  function login(session: MockSession) {
    currentSession.value = session
    persistMockSession(session)
  }

  function logout() {
    currentSession.value = null
    persistMockSession(null)
  }

  return {
    reviewServiceBaseUrl,
    businessServiceBaseUrl,
    operationServiceBaseUrl,
    currentSession,
    isLoggedIn,
    serviceMap,
    getServiceBaseUrl,
    login,
    logout,
  }
})

export function useRoleNavigation() {
  const route = useRoute()
  const router = useRouter()

  const activeRole = computed<AppRole>(() => {
    if (route.path.startsWith('/merchant')) {
      return 'merchant'
    }
    if (route.path.startsWith('/operator')) {
      return 'operator'
    }
    return 'consumer'
  })

  function goRole(role: AppRole) {
    void router.push(`/${role}`)
  }

  return {
    activeRole,
    goRole,
  }
}

export function roleLabel(role: AppRole) {
  if (role === 'consumer') return '消费者端'
  if (role === 'merchant') return '商家端'
  return '运营端'
}
