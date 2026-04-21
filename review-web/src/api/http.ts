import axios from 'axios'
import { useAppStore } from '../stores/app'

type ServiceName = 'review' | 'business' | 'operation'

function createClient(service: ServiceName) {
  const store = useAppStore()

  return axios.create({
    baseURL: store.getServiceBaseUrl(service),
    timeout: 15000,
  })
}

function resolveErrorMessage(error: unknown) {
  if (axios.isAxiosError(error)) {
    const detail = error.response?.data
    if (typeof detail?.message === 'string' && detail.message) {
      return detail.message
    }
    if (typeof detail?.error === 'string' && detail.error) {
      return detail.error
    }
    if (error.message) {
      return error.message
    }
  }

  if (error instanceof Error) {
    return error.message
  }

  return '请求失败，请稍后重试'
}

export async function request<T>(service: ServiceName, config: Parameters<ReturnType<typeof createClient>['request']>[0]) {
  try {
    const client = createClient(service)
    const nextConfig = { ...config }
    if (!(typeof FormData !== 'undefined' && nextConfig.data instanceof FormData)) {
      nextConfig.headers = {
        'Content-Type': 'application/json',
        ...(nextConfig.headers ?? {}),
      }
    }
    const response = await client.request<T>(nextConfig)
    return response.data
  } catch (error) {
    throw new Error(resolveErrorMessage(error))
  }
}
