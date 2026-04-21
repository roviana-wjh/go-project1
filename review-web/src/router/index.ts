import { createRouter, createWebHistory } from 'vue-router'
import { loadMockSession } from '../stores/app'

const routes = [
  {
    path: '/',
    redirect: '/login',
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('../views/LoginView.vue'),
  },
  {
    path: '/consumer',
    name: 'consumer',
    component: () => import('../views/consumer/ConsumerReviewsView.vue'),
  },
  {
    path: '/merchant',
    name: 'merchant',
    component: () => import('../views/merchant/MerchantReviewsView.vue'),
  },
  {
    path: '/operator',
    name: 'operator',
    component: () => import('../views/operator/OperatorAuditsView.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const session = loadMockSession()

  if (to.path === '/login') {
    return session ? `/${session.role}` : true
  }

  if (!session) {
    return '/login'
  }

  const allowedPath = `/${session.role}`
  if (!to.path.startsWith(allowedPath)) {
    return allowedPath
  }

  return true
})

export default router
