<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>社团管理</span>
          <el-button type="primary" @click="goCreate">新建社团</el-button>
        </div>
      </template>

      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-select v-model="filterStatus" placeholder="状态筛选" clearable style="width: 140px" @change="fetchList">
          <el-option label="筹备中" value="preparing" />
          <el-option label="试运行" value="trial" />
          <el-option label="注册成立" value="registered" />
          <el-option label="评估整顿" value="rectifying" />
          <el-option label="注销" value="cancelled" />
        </el-select>
        <el-input v-model="keyword" placeholder="搜索社团名称" clearable style="width: 200px; margin-left: 12px" @clear="fetchList" @keyup.enter="fetchList" />
      </div>

      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="biz_no" label="编号" width="150" />
        <el-table-column prop="name" label="社团名称" min-width="160" />
        <el-table-column prop="college_name" label="所属院系" min-width="120" />
        <el-table-column prop="tutor_name" label="指导教师" width="100" />
        <el-table-column prop="president_name" label="社长" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType[row.status]" size="small">
              {{ row.status_text || row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="goDetail(row.id)">查看</el-button>
            <el-button v-if="row.status !== 'cancelled'" link type="primary" size="small" @click="goEdit(row.id)">编辑</el-button>
            <el-popconfirm v-if="row.status === 'preparing' || row.status === 'cancelled'" title="确认删除此社团？" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button link type="danger" size="small">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="fetchList"
          @current-change="fetchList"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { stAssociationApi } from '@/api/st'

const router = useRouter()

const statusType = { preparing: 'info', trial: 'warning', registered: 'success', rectifying: 'danger', cancelled: 'info' }

const list = ref([])
const loading = ref(false)
const filterStatus = ref('')
const keyword = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

async function fetchList() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (filterStatus.value) params.status = filterStatus.value
    if (keyword.value) params.keyword = keyword.value
    const data = await stAssociationApi.list(params)
    list.value = data.items || []
    total.value = data.total || 0
  } catch (e) {
    console.error('获取社团列表失败', e)
  } finally {
    loading.value = false
  }
}

function goCreate() {
  router.push('/st/association/new')
}

function goEdit(id) {
  router.push(`/st/association/${id}/edit`)
}

function goDetail(id) {
  router.push(`/st/association/${id}`)
}

async function handleDelete(id) {
  try {
    await stAssociationApi.delete(id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e) {
    // 错误已由 http 拦截器处理
  }
}

onMounted(() => {
  fetchList()
})
</script>

<style scoped>
/* .card-header, .filter-bar, .pagination-wrap 已在 App.vue 全局定义 */
</style>
