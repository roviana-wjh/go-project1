<script setup lang="ts">
import { reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { roleLabel, useAppStore, type AppRole } from '../stores/app'

const router = useRouter()
const store = useAppStore()

const form = reactive({
  displayName: '演示用户',
  identity: '10001',
  role: 'consumer' as AppRole,
})

function applyRolePreset(role: AppRole) {
  form.role = role
  if (role === 'consumer') {
    form.displayName = '演示消费者'
    form.identity = '10001'
    return
  }
  if (role === 'merchant') {
    form.displayName = '演示商家'
    form.identity = '30001'
    return
  }
  form.displayName = '演示运营'
  form.identity = 'op-admin'
}

async function handleLogin() {
  if (!form.displayName.trim()) {
    ElMessage.warning('请输入展示名称')
    return
  }
  if (!form.identity.trim()) {
    ElMessage.warning('请输入身份标识')
    return
  }

  store.login({
    displayName: form.displayName.trim(),
    identity: form.identity.trim(),
    role: form.role,
  })
  ElMessage.success(`已进入${roleLabel(form.role)}`)
  await router.push(`/${form.role}`)
}
</script>

<template>
  <div class="login-page">
    <el-card class="login-card" shadow="hover">
      <template #header>
        <div>
          <h1 class="login-title">商品评价系统</h1>
          <p class="login-subtitle">这是一个纯前端模拟登录页，不依赖后端登录、注册或用户表。</p>
        </div>
      </template>

      <el-alert
        type="info"
        :closable="false"
        title="登录后只是在浏览器本地保存一个 mock 会话，用于区分消费者、商家、运营三种页面。"
        class="panel-alert"
      />

      <el-form label-width="96px">
        <el-form-item label="登录角色">
          <el-radio-group v-model="form.role" @change="applyRolePreset(form.role)">
            <el-radio value="consumer">消费者</el-radio>
            <el-radio value="merchant">商家</el-radio>
            <el-radio value="operator">运营</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="展示名称">
          <el-input v-model="form.displayName" placeholder="例如：演示消费者" />
        </el-form-item>

        <el-form-item label="身份标识">
          <el-input v-model="form.identity" placeholder="消费者填 userID，商家填 storeID，运营填 operator" />
        </el-form-item>

        <el-button type="primary" size="large" class="login-submit" @click="handleLogin">进入系统</el-button>
      </el-form>
    </el-card>
  </div>
</template>
