import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/home'
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/LoginPage.vue')
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/RegisterPage.vue')
  },
  {
    path: '/home',
    name: 'Home',
    component: () => import('@/views/HomePage.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/post/:id',
    name: 'PostDetail',
    component: () => import('@/views/PostDetailPage.vue')
  },
  {
    path: '/create-post',
    name: 'CreatePost',
    component: () => import('@/views/CreatePostPage.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/user/:id',
    name: 'UserProfile',
    component: () => import('@/views/UserProfilePage.vue')
  },
  {
    path: '/settings',
    name: 'UserSettings',
    component: () => import('@/views/UserSettingsPage.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/following',
    name: 'FollowingList',
    component: () => import('@/views/FollowingListPage.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/followers',
    name: 'FollowerList',
    component: () => import('@/views/FollowerListPage.vue'),
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    next('/login')
  } else {
    next()
  }
})

export default router
