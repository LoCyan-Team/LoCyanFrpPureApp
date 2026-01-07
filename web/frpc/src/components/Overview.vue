<template>
  <div class="status-container">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span class="title">隧道状态</span>
          <el-button
            type="primary"
            :icon="Refresh"
            @click="fetchData"
            :loading="loading"
            :disabled="loading"
          >
            立即刷新
          </el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="status"
        stripe
        style="width: 100%"
        :default-sort="{ prop: 'type', order: 'ascending' }"
      >
        <el-table-column
          prop="name"
          label="名称"
          sortable
          show-overflow-tooltip
        />
        <el-table-column
          prop="type"
          label="类型"
          width="100"
          sortable
          align="center"
        >
          <template #default="scope">
            <el-tag size="small">{{ scope.row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="local_addr"
          label="本地地址"
          width="180"
          sortable
        />
        <el-table-column
          prop="remote_addr"
          label="远程地址"
          width="180"
          sortable
        />
        <el-table-column prop="plugin" label="插件" width="120" />

        <el-table-column
          prop="status"
          label="状态"
          width="120"
          sortable
          align="center"
        >
          <template #default="scope">
            <el-tag :type="statusMap[scope.row.status] || 'info'" effect="dark">
              {{ scope.row.status === 'running' ? '运行中' : scope.row.status }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column
          prop="err"
          label="详情/错误信息"
          min-width="200"
          show-overflow-tooltip
        >
          <template #default="scope">
            <span :class="{ 'error-text': scope.row.err }">
              {{ scope.row.err || '正常' }}
            </span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'

// 状态对应的颜色
const statusMap: Record<string, string> = {
  running: 'success',
  closed: 'info',
  error: 'danger',
  waiting: 'warning',
}

const status = ref<any[]>([])
const loading = ref(false)

const fetchData = async () => {
  loading.value = true
  try {
    const response = await fetch('/api/status', { credentials: 'include' })
    if (!response.ok) throw new Error('网络请求失败')

    const json = await response.json()
    const tempList: any[] = []

    // 遍历处理数据
    for (let key in json) {
      if (Array.isArray(json[key])) {
        tempList.push(...json[key])
      }
    }
    status.value = tempList
  } catch (err) {
    ElMessage.warning(`获取状态信息失败: ${err}`)
    console.error(err)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.status-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title {
  font-size: 18px;
  font-weight: bold;
}

.error-text {
  color: var(--el-color-danger);
  font-size: 12px;
}

/* 让表格渲染更紧凑些 */
:deep(.el-table) {
  margin-top: 10px;
}
</style>
