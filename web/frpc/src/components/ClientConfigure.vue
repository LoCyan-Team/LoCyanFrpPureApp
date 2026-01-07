<template>
  <div class="config-container">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <span class="title">配置文件编辑</span>
          </div>
          <div class="header-right">
            <el-button
              :icon="Refresh"
              @click="fetchData"
              :loading="loading"
              :disabled="loading"
            >
              刷新
            </el-button>
            <el-button
              type="primary"
              :icon="Upload"
              @click="uploadConfig"
              :loading="uploading"
              :disabled="uploading"
            >
              保存并热重载
            </el-button>
          </div>
        </div>
      </template>

      <div class="config-alerts">
        <el-alert
            title="警告：修改配置并保存后，Frp 客户端将立即执行热重载。"
            type="warning"
            :closable="false"
            show-icon
        />
      </div>

      <div class="editor-wrapper">
        <el-input
          v-model="textarea"
          type="textarea"
          :autosize="{ minRows: 15, maxRows: 30 }"
          placeholder="请输入配置文件内容"
          spellcheck="false"
          class="custom-textarea"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Upload } from '@element-plus/icons-vue'

const textarea = ref('')
const loading = ref(false)
const uploading = ref(false)

const fetchData = async () => {
  loading.value = true
  try {
    const res = await fetch('/api/config', { credentials: 'include' })
    textarea.value = await res.text()
  } catch (err) {
    ElMessage.warning('获取配置失败')
  } finally {
    loading.value = false
  }
}

const uploadConfig = () => {
  if (!textarea.value.trim()) {
    ElMessage.warning('配置不能为空')
    return
  }
  ElMessageBox.confirm('确定要上传并重载吗？', '确认', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消',
  })
    .then(async () => {
      uploading.value = true
      try {
        await fetch('/api/config', {
          method: 'PUT',
          body: textarea.value,
          credentials: 'include',
        })
        await fetch('/api/reload', { credentials: 'include' })
        ElMessage.success('配置已更新')
      } catch (err) {
        ElMessage.error('操作失败')
      } finally {
        uploading.value = false
      }
    })
    .catch(() => {})
}

onMounted(fetchData)
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.title {
  font-size: 18px;
  font-weight: bold;
}

/* 编辑器外层容器 */
.editor-wrapper {
  border-radius: var(--el-border-radius-base);
  /* 这里的边框也会随主题变色 */
  border: 1px solid var(--el-border-color-lighter);
  overflow: hidden;
}

/* 深度选择器：强制覆盖 el-textarea 内部样式 */
:deep(.custom-textarea .el-textarea__inner) {
  font-family: 'Fira Code', 'Cascadia Code', Consolas, Monaco, monospace;

  /* 关键：直接绑定 Element Plus 的输入框动态变量 */
  background-color: var(
    --el-input-bg-color,
    var(--el-fill-color-blank)
  ) !important;
  color: var(--el-input-text-color, var(--el-text-color-primary)) !important;

  /* 移除边框和阴影，让它看起来更像一个纯粹的代码区域 */
  border: none;
  box-shadow: none !important;

  line-height: 1.6;
  padding: 16px;

  /* 增加平滑过渡效果 */
  transition:
    background-color var(--el-transition-duration),
    color var(--el-transition-duration);
}

/* 聚焦时的状态控制 */
:deep(.custom-textarea .el-textarea__inner:focus) {
  /* 聚焦时稍微改变背景色，增加反馈感 */
  background-color: var(--el-fill-color-light) !important;
}

.config-alerts {
  margin-bottom: 10px;
}
</style>
