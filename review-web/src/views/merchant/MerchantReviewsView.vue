<script setup lang="ts">
import { computed, onMounted, reactive, ref, watchEffect } from 'vue'
import { ElMessage } from 'element-plus'
import AppShell from '../../layouts/AppShell.vue'
import LocalImageUploader from '../../components/LocalImageUploader.vue'
import { replyReview } from '../../api/businessService'
import { appealReview, listGoodsScoreRank, listReviewByStore } from '../../api/reviewService'
import { useAppStore } from '../../stores/app'
import type { GoodsScoreRankItem, ReplyReviewPayload, ReviewListItem } from '../../types/review'

const store = useAppStore()

const storeQuery = reactive({ storeID: 30001, page: 1, pageSize: 10 })
const storeList = ref<ReviewListItem[]>([])
const total = ref(0)
const loading = ref(false)
const selectedReview = ref<ReviewListItem | null>(null)
const rankLoading = ref(false)
const rankQuery = reactive({ page: 1, pageSize: 10 })
const rankList = ref<GoodsScoreRankItem[]>([])
const rankTotal = ref(0)

const replyForm = reactive<ReplyReviewPayload>({
  reviewID: '',
  storeID: 30001,
  content: '',
  picInfo: '[]',
  videoInfo: '[]',
  extJSON: '{}',
  ctrlJSON: '{}',
})

const appealForm = reactive({
  userID: 90001,
  reviewID: '',
  reason: '',
  picInfo: '[]',
})

watchEffect(() => {
  const session = store.currentSession
  if (!session || session.role !== 'merchant') {
    return
  }
  const identity = Number(session.identity) || 30001
  storeQuery.storeID = identity
  replyForm.storeID = identity
  appealForm.userID = identity
})

const merchantStats = computed(() => [
  { label: '店铺评价数', value: total.value || storeList.value.length, hint: '店铺全量评价入口' },
  { label: '待回复', value: storeList.value.filter((item) => item.hasReply === 0).length, hint: '需要尽快互动' },
  { label: '已审核通过', value: storeList.value.filter((item) => item.status === 20).length, hint: '可重点运营' },
  { label: '平均评分', value: storeList.value.length ? (storeList.value.reduce((sum, item) => sum + item.score, 0) / storeList.value.length).toFixed(1) : '0.0', hint: '店铺口碑概览' },
])

const quickReplyTemplates = [
  '感谢您的反馈，我们会持续优化服务，欢迎再次光临。',
  '非常感谢您的支持，祝您生活愉快，期待下次购买。',
  '收到建议了，我们已同步给运营团队并持续改进。',
]

async function handleLoadStoreReviews() {
  loading.value = true
  try {
    const result = await listReviewByStore(storeQuery.storeID, {
      page: storeQuery.page,
      pageSize: storeQuery.pageSize,
    })
    storeList.value = result.list ?? []
    total.value = result.total ?? 0
    const preferred = storeList.value.find((item) => item.hasReply === 0) ?? storeList.value[0]
    if (preferred) {
      useRowForReply(preferred)
    }
  } catch (error) {
    ElMessage.error((error as Error).message)
  } finally {
    loading.value = false
  }
}

async function handleLoadGoodsRank() {
  rankLoading.value = true
  try {
    const result = await listGoodsScoreRank({
      page: rankQuery.page,
      pageSize: rankQuery.pageSize,
    })
    rankList.value = result.list ?? []
    rankTotal.value = result.total ?? 0
  } catch (error) {
    ElMessage.error((error as Error).message)
  } finally {
    rankLoading.value = false
  }
}

function useRowForReply(row: ReviewListItem) {
  selectedReview.value = row
  replyForm.reviewID = String(row.reviewID)
  replyForm.storeID = storeQuery.storeID
  appealForm.reviewID = String(row.reviewID)
  if (!replyForm.content) {
    replyForm.content = `感谢您的评价（评分 ${row.score}/5），我们会持续优化商品与服务体验。`
  }
}

async function handleReplyReview() {
  if (!/^\d+$/.test(String(replyForm.reviewID)) || String(replyForm.reviewID) === '0') {
    ElMessage.warning('请先从左侧列表选择一条评价')
    return
  }
  if (!replyForm.content.trim()) {
    ElMessage.warning('请填写回复内容')
    return
  }
  loading.value = true
  try {
    const result = await replyReview(replyForm)
    ElMessage.success(`回复成功，replyID=${result.replyID}`)
    replyForm.content = ''
    replyForm.picInfo = '[]'
    replyForm.videoInfo = '[]'
    replyForm.extJSON = '{}'
    replyForm.ctrlJSON = '{}'
    await handleLoadStoreReviews()
    useNextPendingReview()
  } catch (error) {
    ElMessage.error((error as Error).message)
  } finally {
    loading.value = false
  }
}

function applyTemplate(content: string) {
  replyForm.content = content
}

function useNextPendingReview() {
  if (!storeList.value.length) return
  const currentIdx = storeList.value.findIndex((item) => String(item.reviewID) === String(replyForm.reviewID))
  const nextPending = storeList.value.slice(Math.max(currentIdx + 1, 0)).find((item) => item.hasReply === 0)
    ?? storeList.value.find((item) => item.hasReply === 0)
  if (nextPending) {
    useRowForReply(nextPending)
  }
}

async function handleAppealReview() {
  if (!/^\d+$/.test(String(appealForm.reviewID)) || String(appealForm.reviewID) === '0') {
    ElMessage.warning('请先选择要申诉的评价')
    return
  }
  loading.value = true
  try {
    const result = await appealReview(appealForm)
    ElMessage.success(`申诉提交成功，appealID=${result.appealID}`)
  } catch (error) {
    ElMessage.error((error as Error).message)
  } finally {
    loading.value = false
  }
}

function statusLabel(status: number) {
  if (status === 10) return '待审核'
  if (status === 20) return '审核通过'
  if (status === 30) return '审核驳回'
  if (status === 40) return '隐藏'
  return `未知(${status})`
}

function replyLabel(hasReply: number) {
  return hasReply === 1 ? '已回复' : '待回复'
}

onMounted(() => {
  void handleLoadStoreReviews()
  void handleLoadGoodsRank()
})
</script>

<template>
  <AppShell>
    <div class="page-stack">
      <div class="hero-grid">
        <section class="promo-card">
          <div class="shell-badge">店铺口碑运营</div>
          <h2>把评价区做成你的第二张商品详情页</h2>
          <p>商家端更像电商后台里的评价运营中心，支持查看店铺评价、快速回复和发起申诉。</p>
          <div class="promo-tags">
            <span>高分评价沉淀</span>
            <span>一键带入回复</span>
            <span>空白主图占位</span>
            <span>申诉流程预演</span>
          </div>
        </section>

        <el-card class="commerce-card hero-side-card" shadow="never">
          <div class="section-header">
            <div>
              <h3 class="section-title">店铺门面卡</h3>
              <p class="section-subtitle">模拟商详页侧边的店铺信息区域</p>
            </div>
            <el-tag type="success" round>店铺 {{ storeQuery.storeID }}</el-tag>
          </div>
          <div class="product-panel">
            <div class="placeholder-image">
              <span>店铺主图占位</span>
            </div>
            <div class="product-info">
              <h3>官方旗舰店体验区</h3>
              <div class="feature-list">
                <span>48 小时发货</span>
                <span>客服秒回</span>
                <span>品质保障</span>
                <span>退货包运费</span>
              </div>
              <p class="section-subtitle">这里的图文先用占位块，后续可替换成店铺招牌或商品封面。</p>
            </div>
          </div>
        </el-card>
      </div>

      <div class="stats-grid">
        <div v-for="item in merchantStats" :key="item.label" class="stats-card">
          <span class="label">{{ item.label }}</span>
          <span class="value">{{ item.value }}</span>
          <span class="hint">{{ item.hint }}</span>
        </div>
      </div>

      <div class="panel-grid">
        <el-card class="commerce-card table-card" shadow="never">
          <div class="section-header">
            <div>
              <h3 class="section-title">店铺评价看板</h3>
              <p class="section-subtitle">优先查看差评、未回复评价和可运营内容</p>
            </div>
            <el-tag type="warning" round>店铺评价</el-tag>
          </div>
          <div class="table-toolbar">
            <el-input-number v-model="storeQuery.storeID" :min="1" />
            <el-input-number v-model="storeQuery.page" :min="1" />
            <el-input-number v-model="storeQuery.pageSize" :min="1" :max="100" />
            <el-button type="primary" :loading="loading" @click="handleLoadStoreReviews">刷新看板</el-button>
          </div>
          <el-table :data="storeList" stripe>
            <el-table-column prop="reviewID" label="评价单" min-width="120" />
            <el-table-column prop="userID" label="用户" min-width="120" />
            <el-table-column prop="score" label="评分" min-width="80" />
            <el-table-column label="状态" min-width="110">
              <template #default="{ row }">{{ statusLabel(row.status) }}</template>
            </el-table-column>
            <el-table-column label="回复状态" min-width="100">
              <template #default="{ row }">
                <el-tag :type="row.hasReply === 1 ? 'success' : 'warning'" round>{{ replyLabel(row.hasReply) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="content" label="评价内容" min-width="260" show-overflow-tooltip />
            <el-table-column label="操作" min-width="140">
              <template #default="{ row }">
                <el-button link type="primary" @click="useRowForReply(row)">带入回复</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div class="table-footer">共 {{ total }} 条</div>
        </el-card>

        <el-card class="commerce-card table-card" shadow="never">
          <div class="section-header">
            <div>
              <h3 class="section-title">商品评分排行</h3>
              <p class="section-subtitle">基于 Redis ZSet 的实时口碑榜（仅统计审核通过评价）</p>
            </div>
            <el-tag type="danger" round>评分榜</el-tag>
          </div>
          <div class="table-toolbar">
            <el-input-number v-model="rankQuery.page" :min="1" />
            <el-input-number v-model="rankQuery.pageSize" :min="1" :max="100" />
            <el-button type="primary" :loading="rankLoading" @click="handleLoadGoodsRank">刷新排行</el-button>
          </div>
          <el-table :data="rankList" stripe>
            <el-table-column label="排名" min-width="80">
              <template #default="{ $index }">{{ (rankQuery.page - 1) * rankQuery.pageSize + $index + 1 }}</template>
            </el-table-column>
            <el-table-column prop="spuID" label="商品 SPU" min-width="160" />
            <el-table-column label="平均分" min-width="120">
              <template #default="{ row }">{{ Number(row.avgScore).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column prop="reviewCount" label="评价数" min-width="120" />
          </el-table>
          <div class="table-footer">共 {{ rankTotal }} 个商品上榜</div>
        </el-card>

        <el-card class="commerce-card" shadow="never">
          <div class="section-header">
            <div>
              <h3 class="section-title">回复评价</h3>
              <p class="section-subtitle">回复接口优先走 review-B，适合作为商家 BFF 入口</p>
            </div>
            <el-tag type="info" round>商家互动</el-tag>
          </div>
          <div class="placeholder-image" style="min-height: 140px; margin-bottom: 16px">
            <span>商品缩略图占位</span>
          </div>
          <div class="dense-form">
            <div class="inline-actions">
              <el-tag v-if="selectedReview" type="info" round>
                当前目标：#{{ selectedReview.reviewID }}（{{ replyLabel(selectedReview.hasReply) }}）
              </el-tag>
              <el-button link type="primary" @click="useNextPendingReview">下一条待回复</el-button>
            </div>
            <div class="dense-form-grid">
              <div>
                <label class="mini-note">Review ID</label>
                <el-input v-model="replyForm.reviewID" placeholder="请从左侧点“带入回复”或手动输入" />
              </div>
              <div>
                <label class="mini-note">Store ID</label>
                <el-input-number v-model="replyForm.storeID" :min="1" />
              </div>
            </div>
            <div>
              <label class="mini-note">回复内容</label>
              <el-input v-model="replyForm.content" type="textarea" :rows="4" placeholder="感谢支持，欢迎继续晒单..." />
            </div>
            <div class="inline-actions">
              <el-button
                v-for="text in quickReplyTemplates"
                :key="text"
                size="small"
                plain
                @click="applyTemplate(text)"
              >
                一键模板
              </el-button>
            </div>
            <div class="dense-form-grid">
              <div>
                <label class="mini-note">回复图片</label>
                <LocalImageUploader v-model="replyForm.picInfo" />
              </div>
              <div>
                <label class="mini-note">视频信息</label>
                <el-input v-model="replyForm.videoInfo" type="textarea" :rows="2" />
              </div>
            </div>
            <div class="dense-form-grid">
              <div>
                <label class="mini-note">extJSON</label>
                <el-input v-model="replyForm.extJSON" type="textarea" :rows="2" />
              </div>
              <div>
                <label class="mini-note">ctrlJSON</label>
                <el-input v-model="replyForm.ctrlJSON" type="textarea" :rows="2" />
              </div>
            </div>
            <el-button type="primary" :loading="loading" @click="handleReplyReview">发送回复</el-button>
          </div>
        </el-card>

        <el-card class="commerce-card" shadow="never">
          <div class="section-header">
            <div>
              <h3 class="section-title">发起申诉</h3>
              <p class="section-subtitle">用于处理异常评价、恶意内容或误判审核</p>
            </div>
            <el-tag type="danger" round>申诉流程</el-tag>
          </div>
          <div class="placeholder-image" style="min-height: 140px; margin-bottom: 16px">
            <span>申诉凭证图占位</span>
          </div>
          <div class="dense-form">
            <div class="dense-form-grid">
              <div>
                <label class="mini-note">商家身份</label>
                <el-input-number v-model="appealForm.userID" :min="1" />
              </div>
              <div>
                <label class="mini-note">Review ID</label>
                <el-input v-model="appealForm.reviewID" placeholder="请从左侧点“带入回复”或手动输入" />
              </div>
            </div>
            <div>
              <label class="mini-note">申诉原因</label>
              <el-input v-model="appealForm.reason" type="textarea" :rows="4" placeholder="请描述异常评价的原因..." />
            </div>
            <div>
              <label class="mini-note">申诉凭证图片</label>
              <LocalImageUploader v-model="appealForm.picInfo" :limit="3" />
            </div>
            <el-button type="primary" :loading="loading" @click="handleAppealReview">提交申诉</el-button>
          </div>
        </el-card>
      </div>
    </div>
  </AppShell>
</template>
