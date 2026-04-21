<script setup lang="ts">
import { Loading, Plus } from '@element-plus/icons-vue'
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { UploadRequestOptions, UploadUserFile } from 'element-plus'
import { uploadReviewImage } from '../api/reviewService'
import { useAppStore } from '../stores/app'
import type { UploadedMediaItem } from '../types/review'

const props = withDefaults(
  defineProps<{
    modelValue: string
    limit?: number
  }>(),
  {
    limit: 6,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const store = useAppStore()
const internalItems = ref<UploadedMediaItem[]>([])
const fileList = ref<UploadUserFile[]>([])
const uploading = ref(false)

const tipText = computed(() => `最多上传 ${props.limit} 张，本地存储到 review-service/uploads/review`)

watch(
  () => props.modelValue,
  (value) => {
    internalItems.value = parseMedia(value)
    fileList.value = internalItems.value.map((item, index) => ({
      name: item.name || `image-${index + 1}`,
      url: toPreviewUrl(item.url),
      response: item,
    }))
  },
  { immediate: true },
)

async function handleHttpRequest(options: UploadRequestOptions) {
  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', options.file)
    const result = await uploadReviewImage(formData)
    const nextItems = [...internalItems.value, result]
    emitValue(nextItems)
    options.onSuccess?.(result)
    ElMessage.success('图片上传成功')
  } catch (error) {
    const message = error instanceof Error ? error.message : '图片上传失败'
    ElMessage.error(message)
    options.onError?.(
      Object.assign(new Error(message), {
        status: 500,
        method: 'POST',
        url: '/v1/upload/review-image',
      }) as never,
    )
  } finally {
    uploading.value = false
  }
}

function handleRemove(file: UploadUserFile) {
  const rawUrl = ((file.response as UploadedMediaItem | undefined)?.url || file.url || '').replace(store.reviewServiceBaseUrl, '')
  const nextItems = internalItems.value.filter((item) => item.url !== rawUrl)
  emitValue(nextItems)
}

function beforeUpload(rawFile: File) {
  const isImage = rawFile.type.startsWith('image/')
  if (!isImage) {
    ElMessage.warning('只能上传图片文件')
    return false
  }

  const isTooLarge = rawFile.size > 10 * 1024 * 1024
  if (isTooLarge) {
    ElMessage.warning('图片大小不能超过 10MB')
    return false
  }
  return true
}

function emitValue(items: UploadedMediaItem[]) {
  emit('update:modelValue', JSON.stringify(items))
}

function parseMedia(raw: string): UploadedMediaItem[] {
  if (!raw) {
    return []
  }
  try {
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed)) {
      return parsed.filter((item) => item && typeof item.url === 'string')
    }
  } catch {
    return []
  }
  return []
}

function toPreviewUrl(url: string) {
  if (!url) {
    return ''
  }
  if (url.startsWith('http://') || url.startsWith('https://')) {
    return url
  }
  return `${store.reviewServiceBaseUrl}${url}`
}
</script>

<template>
  <div class="uploader-panel">
    <el-upload
      v-model:file-list="fileList"
      action="#"
      list-type="picture-card"
      :limit="limit"
      :auto-upload="true"
      :http-request="handleHttpRequest"
      :before-upload="beforeUpload"
      :on-remove="handleRemove"
      :show-file-list="true"
    >
      <el-icon v-if="!uploading"><Plus /></el-icon>
      <el-icon v-else class="is-loading"><Loading /></el-icon>
    </el-upload>
    <p class="mini-note">{{ tipText }}</p>
  </div>
</template>
