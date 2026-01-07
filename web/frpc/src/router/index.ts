import { createRouter, createWebHashHistory } from 'vue-router'
import Overview from '../components/Overview.vue'
import ClientConfigure from '../components/ClientConfigure.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      name: '概览',
      component: Overview,
    },
    {
      path: '/configure',
      name: '客户端配置',
      component: ClientConfigure,
    },
  ],
})

export default router
