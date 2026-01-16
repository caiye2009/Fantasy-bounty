import { createRouter, createWebHistory } from 'vue-router'
import BountyHome from '@/views/BountyHome.vue'

const routes = [
  {
    path: '/',
    component: BountyHome
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
