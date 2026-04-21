<script setup lang="ts">
import { Connection, Monitor, OfficeBuilding, ShoppingBag, SwitchButton } from '@element-plus/icons-vue'
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { roleLabel, useAppStore, useRoleNavigation, type AppRole } from '../stores/app'

const store = useAppStore()
const { activeRole, goRole } = useRoleNavigation()
const router = useRouter()

const roleOptions: Array<{ role: AppRole; label: string; icon: unknown }> = [
  { role: 'consumer', label: '消费者端', icon: ShoppingBag },
  { role: 'merchant', label: '商家端', icon: OfficeBuilding },
  { role: 'operator', label: '运营端', icon: Monitor },
]

const roleTitle = computed(() => roleOptions.find((item) => item.role === activeRole.value)?.label ?? '评价系统')
const roleSubtitle = computed(() => {
  if (activeRole.value === 'consumer') {
    return '拼团商品详情、评价晒单、订单查询与编辑体验'
  }
  if (activeRole.value === 'merchant') {
    return '店铺评价运营、回复互动与申诉处理工作台'
  }
  return '评价审核、申诉处理与风控工作台'
})
const sessionSummary = computed(() => {
  if (!store.currentSession) {
    return '未登录'
  }
  return `${store.currentSession.displayName} / ${roleLabel(store.currentSession.role)} / ${store.currentSession.identity}`
})

async function handleLogout() {
  store.logout()
  await router.push('/login')
}
</script>

<template>
  <div class="app-shell">
    <section class="shell-banner">
      <div class="shell-banner__content">
        <div class="shell-badge">电商评价中心</div>
        <h1>商品评价系统前端原型</h1>
        <p>{{ roleTitle }}，{{ roleSubtitle }}</p>

        <div class="shell-role-switch">
          <el-button
            v-for="item in roleOptions"
            :key="item.role"
            :type="activeRole === item.role ? 'primary' : 'default'"
            round
            @click="goRole(item.role)"
          >
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </el-button>
        </div>

        <div class="shell-meta">
          <div class="shell-meta-card">
            <el-icon><Connection /></el-icon>
            <div>
              <strong>当前身份</strong>
              <span>{{ sessionSummary }}</span>
            </div>
          </div>
          <div class="shell-meta-card">
            <el-icon><SwitchButton /></el-icon>
            <div>
              <strong>快速切换</strong>
              <span>支持消费者 / 商家 / 运营三端模拟联调</span>
            </div>
          </div>
        </div>
      </div>

      <div class="shell-banner__actions">
        <el-button type="danger" plain round @click="handleLogout">退出登录</el-button>
      </div>
    </section>

    <el-card class="base-url-card commerce-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span>服务地址配置</span>
          <span class="card-hint">开发阶段可直接调整到各个后端 HTTP 端口</span>
        </div>
      </template>
      <div class="url-grid">
        <div class="service-field">
          <label>review-service</label>
          <el-input v-model="store.reviewServiceBaseUrl" placeholder="http://localhost:8000" />
        </div>
        <div class="service-field">
          <label>review-B</label>
          <el-input v-model="store.businessServiceBaseUrl" placeholder="http://localhost:8001" />
        </div>
        <div class="service-field">
          <label>review-O</label>
          <el-input v-model="store.operationServiceBaseUrl" placeholder="http://localhost:8002" />
        </div>
      </div>
    </el-card>

    <main class="shell-main">
      <slot />
    </main>
  </div>
</template>
