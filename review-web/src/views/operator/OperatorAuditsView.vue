<script setup lang="ts">
import { computed, onMounted, reactive, ref, watchEffect } from 'vue'
import { ElMessage } from 'element-plus'
import AppShell from '../../layouts/AppShell.vue'
import { auditAppeal, auditReview, listPendingAppeals, listPendingReviews } from '../../api/operationService'
import { useAppStore } from '../../stores/app'
import type { AppealListItem, ReviewListItem } from '../../types/review'

const store = useAppStore()

const loading = ref(false)

const reviewAuditForm = reactive({
  reviewID: '',
  result: 1,
  remark: '',
  operator: 'op-admin',
})

const appealAuditForm = reactive({
  appealID: '',
  result: 1,
  remark: '',
  operator: 'op-admin',
})

const pendingQuery = reactive({ page: 1, pageSize: 10 })
const pendingReviews = ref<ReviewListItem[]>([])
const pendingTotal = ref(0)
const pendingAppealQuery = reactive({ page: 1, pageSize: 10 })
const pendingAppeals = ref<AppealListItem[]>([])
const pendingAppealTotal = ref(0)

watchEffect(() => {
  const session = store.currentSession
  if (!session || session.role !== 'operator') {
    return
  }
  reviewAuditForm.operator = session.identity
  appealAuditForm.operator = session.identity
})

const operatorStats = computed(() => [
  { label: '待处理评价', value: pendingTotal.value || 0, hint: '自动拉取待审核队列' },
  { label: '待处理申诉', value: pendingAppealTotal.value || 0, hint: '自动拉取待审核申诉' },
  { label: '审核动作', value: reviewAuditForm.result === 1 ? '通过' : '驳回', hint: '当前评价审核结果' },
  { label: '操作人', value: reviewAuditForm.operator || '未设置', hint: '运营账号标识' },
])

function formatTime(timestamp: number | string) {
  const n = Number(timestamp)
  if (!n) {
    return '-'
  }
  return new Date(n).toLocaleString()
}

function pickPendingReview(row: ReviewListItem) {
  reviewAuditForm.reviewID = String(row.reviewID ?? '')
}

function pickPendingAppeal(row: AppealListItem) {
  appealAuditForm.appealID = String(row.appealID ?? '')
}

async function handleLoadPendingReviews() {
  loading.value = true
  try {
    const res = await listPendingReviews({
      page: pendingQuery.page,
      pageSize: pendingQuery.pageSize,
    })
    pendingReviews.value = res.list ?? []
    pendingTotal.value = res.total ?? 0
  } catch (error) {
    ElMessage.error((error as Error).message)
  } finally {
    loading.value = false
  }
}

async function handleLoadPendingAppeals() {
  loading.value = true
  try {
    const res = await listPendingAppeals({
      page: pendingAppealQuery.page,
      pageSize: pendingAppealQuery.pageSize,
    })
    pendingAppeals.value = res.list ?? []
    pendingAppealTotal.value = res.total ?? 0
  } catch (error) {
    ElMessage.error((error as Error).message)
  } finally {
    loading.value = false
  }
}

async function handleAuditReview() {
  if (!/^\d+$/.test(reviewAuditForm.reviewID) || reviewAuditForm.reviewID === '0') {
    ElMessage.warning('请先选择或输入正确的 Review ID')
    return
  }
  loading.value = true
  try {
    await auditReview(reviewAuditForm)
    ElMessage.success('评价审核已提交')
    await handleLoadPendingReviews()
  } catch (error) {
    ElMessage.error((error as Error).message)
  } finally {
    loading.value = false
  }
}

async function handleAuditAppeal() {
  if (!/^\d+$/.test(appealAuditForm.appealID) || appealAuditForm.appealID === '0') {
    ElMessage.warning('请输入正确的 Appeal ID')
    return
  }
  loading.value = true
  try {
    await auditAppeal(appealAuditForm)
    ElMessage.success('申诉审核已提交')
    await handleLoadPendingAppeals()
  } catch (error) {
    ElMessage.error((error as Error).message)
  } finally {
    loading.value = false
  }
}

function appealStatusLabel(status: number) {
  if (status === 10) return '待审核'
  if (status === 20) return '申诉通过'
  if (status === 30) return '申诉驳回'
  return `未知(${status})`
}

onMounted(() => {
  void Promise.all([handleLoadPendingReviews(), handleLoadPendingAppeals()])
})
</script>

<template>
  <AppShell>
    <div class="page-stack">
      <div class="hero-grid">
        <section class="promo-card">
          <div class="shell-badge">运营审核中台</div>
          <h2>像电商平台风控后台一样处理评价与申诉</h2>
          <p>先用卡片化工作台承接审核动作，后续如果后端补列表接口，可以直接升级成完整审核看板。</p>
          <div class="promo-tags">
            <span>评价审核</span>
            <span>申诉审核</span>
            <span>运营备注</span>
            <span>空白证据图占位</span>
          </div>
        </section>

        <el-card class="commerce-card hero-side-card" shadow="never">
          <div class="section-header">
            <div>
              <h3 class="section-title">审核总览</h3>
              <p class="section-subtitle">适合后续接入待审核列表和风控标签</p>
            </div>
            <el-tag type="danger" round>运营工作台</el-tag>
          </div>
          <div class="stats-grid">
            <div v-for="item in operatorStats" :key="item.label" class="stats-card">
              <span class="label">{{ item.label }}</span>
              <span class="value">{{ item.value }}</span>
              <span class="hint">{{ item.hint }}</span>
            </div>
          </div>
        </el-card>
      </div>

      <div class="page-grid two-columns">
        <el-card class="commerce-card" shadow="never">
          <div class="section-header">
            <div>
              <h3 class="section-title">审核评价</h3>
              <p class="section-subtitle">自动拉取待审核队列，点选后可直接审核</p>
            </div>
            <el-tag type="info" round>Review Audit</el-tag>
          </div>
          <div class="table-toolbar" style="margin-bottom: 12px">
            <el-input-number v-model="pendingQuery.page" :min="1" />
            <el-input-number v-model="pendingQuery.pageSize" :min="1" :max="100" />
            <el-button type="primary" :loading="loading" @click="handleLoadPendingReviews">刷新待审核列表</el-button>
          </div>
          <div class="placeholder-image" style="min-height: 140px; margin-bottom: 16px; align-items: flex-start; justify-content: flex-start">
            <el-table :data="pendingReviews" size="small" stripe style="width: 100%">
              <el-table-column prop="reviewID" label="ReviewID" min-width="100" />
              <el-table-column prop="userID" label="用户" min-width="90" />
              <el-table-column prop="orderID" label="订单" min-width="90" />
              <el-table-column prop="score" label="评分" min-width="70" />
              <el-table-column label="创建时间" min-width="170">
                <template #default="{ row }">
                  {{ formatTime(row.createAt) }}
                </template>
              </el-table-column>
              <el-table-column label="操作" min-width="110" fixed="right">
                <template #default="{ row }">
                  <el-button link type="primary" @click="pickPendingReview(row)">选中审核</el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
          <div class="dense-form">
            <div>
              <label class="mini-note">Review ID</label>
              <el-input v-model="reviewAuditForm.reviewID" placeholder="请输入 reviewID" />
            </div>
            <div>
              <label class="mini-note">审核结果</label>
              <el-radio-group v-model="reviewAuditForm.result">
                <el-radio :value="1">通过</el-radio>
                <el-radio :value="2">驳回</el-radio>
              </el-radio-group>
            </div>
            <div>
              <label class="mini-note">审核备注</label>
              <el-input v-model="reviewAuditForm.remark" type="textarea" :rows="4" placeholder="填写审核意见..." />
            </div>
            <div>
              <label class="mini-note">操作人</label>
              <el-input v-model="reviewAuditForm.operator" />
            </div>
            <el-button type="primary" :loading="loading" @click="handleAuditReview">提交评价审核</el-button>
          </div>
        </el-card>

        <el-card class="commerce-card" shadow="never">
          <div class="section-header">
            <div>
              <h3 class="section-title">审核申诉</h3>
              <p class="section-subtitle">当前先按 appealID 工作台式处理，待后端补列表后可无缝扩展</p>
            </div>
            <el-tag type="warning" round>Appeal Audit</el-tag>
          </div>
          <div class="table-toolbar" style="margin-bottom: 12px">
            <el-input-number v-model="pendingAppealQuery.page" :min="1" />
            <el-input-number v-model="pendingAppealQuery.pageSize" :min="1" :max="100" />
            <el-button type="primary" :loading="loading" @click="handleLoadPendingAppeals">刷新待审核申诉</el-button>
          </div>
          <div class="placeholder-image" style="min-height: 140px; margin-bottom: 16px; align-items: flex-start; justify-content: flex-start">
            <el-table :data="pendingAppeals" size="small" stripe style="width: 100%">
              <el-table-column prop="appealID" label="AppealID" min-width="120" />
              <el-table-column prop="reviewID" label="ReviewID" min-width="120" />
              <el-table-column prop="storeID" label="店铺" min-width="90" />
              <el-table-column label="状态" min-width="100">
                <template #default="{ row }">
                  {{ appealStatusLabel(row.status) }}
                </template>
              </el-table-column>
              <el-table-column prop="reason" label="申诉原因" min-width="180" show-overflow-tooltip />
              <el-table-column label="创建时间" min-width="170">
                <template #default="{ row }">
                  {{ formatTime(row.createAt) }}
                </template>
              </el-table-column>
              <el-table-column label="操作" min-width="120" fixed="right">
                <template #default="{ row }">
                  <el-button link type="primary" @click="pickPendingAppeal(row)">选中审核</el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
          <div class="dense-form">
            <div>
              <label class="mini-note">Appeal ID</label>
              <el-input v-model="appealAuditForm.appealID" placeholder="请输入 appealID" />
            </div>
            <div>
              <label class="mini-note">审核结果</label>
              <el-radio-group v-model="appealAuditForm.result">
                <el-radio :value="1">通过</el-radio>
                <el-radio :value="2">驳回</el-radio>
              </el-radio-group>
            </div>
            <div>
              <label class="mini-note">审核备注</label>
              <el-input v-model="appealAuditForm.remark" type="textarea" :rows="4" placeholder="填写申诉处理意见..." />
            </div>
            <div>
              <label class="mini-note">操作人</label>
              <el-input v-model="appealAuditForm.operator" />
            </div>
            <el-button type="primary" :loading="loading" @click="handleAuditAppeal">提交申诉审核</el-button>
          </div>
        </el-card>
      </div>
    </div>
  </AppShell>
</template>
