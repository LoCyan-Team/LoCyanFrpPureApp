<template>
  <el-container class="layout-container">
    <el-header class="custom-header">
      <div class="header-content">
        <div class="brand">
          <el-icon size="22"><Platform /></el-icon>
          <span class="brand-title">LoCyanFrp 客户端网页面板</span>
        </div>
        <div class="spacer"></div>
        <div class="actions">
          <el-switch
            v-model="isDark"
            inline-prompt
            active-text="Dark"
            inactive-text="Light"
            @change="toggleDark"
          />
        </div>
      </div>
    </el-header>

    <el-container class="main-body">
      <el-aside width="220px" class="custom-aside">
        <el-menu
          :default-active="activePath"
          class="side-menu"
          router
          @select="handleSelect"
        >
          <el-menu-item index="/">
            <el-icon><Monitor /></el-icon>
            <span>概览</span>
          </el-menu-item>
          <el-menu-item index="/configure">
            <el-icon><Setting /></el-icon>
            <span>客户端配置</span>
          </el-menu-item>
          <el-menu-item index="help">
            <el-icon><QuestionFilled /></el-icon>
            <span>帮助文档</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <el-main class="custom-main">
        <router-view></router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useDark, useToggle } from '@vueuse/core'
import {
  Platform,
  Monitor,
  Setting,
  QuestionFilled,
} from '@element-plus/icons-vue'

// 直接使用 VueUse 控制深色模式
const isDark = useDark()
const toggleDark = useToggle(isDark)

const route = useRoute()
const activePath = computed(() => route.path)

const handleSelect = (key: string) => {
  if (key === 'help') {
    window.open('https://docs.locyanfrp.cn/frp/dashboard/client', '_blank')
  }
}
</script>

<style scoped>
/* 撑满全屏，移除多余的外边距 */
.layout-container {
  height: 100vh;
  border: none !important; /* 强制移除可能存在的外边框 */
}

/* Header 样式：跟随主题背景，仅保留下边框 */
.custom-header {
  display: flex;
  align-items: center;
  background-color: var(--el-bg-color);
  border-bottom: 1px solid var(--el-border-color-light);
  color: var(--el-text-color-primary); /* 自动适配深浅色文字 */
  padding: 0 20px;
}

.header-content {
  display: flex;
  align-items: center;
  width: 100%;
  height: 100%;
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.brand-title {
  font-size: 18px;
  font-weight: 600;
}

.spacer {
  flex-grow: 1;
}

/* 侧边栏：移除右侧多余边框 */
.custom-aside {
  background-color: var(--el-bg-color);
  border-right: 1px solid var(--el-border-color-light);
}

.side-menu {
  border-right: none !important; /* 移除 el-menu 自带的右边框 */
}

/* 主内容区背景色稍微区分开，更有层次感 */
.custom-main {
  background-color: var(--el-bg-color-page);
  padding: 20px;
}

/* 全局消除 body 默认间距 */
:global(body) {
  margin: 0;
  padding: 0;
}
</style>
