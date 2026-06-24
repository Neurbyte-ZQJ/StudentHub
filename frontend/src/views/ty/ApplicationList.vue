<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>入团申请</span>
          <div class="header-actions">
            <el-button
              v-if="authStore.user?.student_id"
              @click="goMyDevelopment"
            >
              查看我的团员发展
            </el-button>
            <el-button type="primary" @click="goCreate">新增申请</el-button>
          </div>
        </div>
      </template>

      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-select v-model="filterStatus" placeholder="状态筛选" clearable style="width: 140px" @change="fetchList">
          <el-option label="草稿" value="S0" />
          <el-option label="待审" value="S1" />
          <el-option label="审批中" value="S2" />
          <el-option label="通过" value="S3" />
          <el-option label="驳回" value="S4" />
        </el-select>
      </div>

      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="biz_no" label="编号" width="150" />
        <el-table-column prop="student_name" label="申请人" width="100" />
        <el-table-column prop="student_no" label="学号" width="120" />
        <el-table-column prop="branch_name" label="团支部" min-width="140" />
        <el-table-column prop="college_name" label="院系" min-width="120" />
        <el-table-column prop="apply_date" label="申请日期" width="110" />
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusType[row.status]" size="small">
              {{ statusMap[row.status] || row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="goDetail(row.id)">查看</el-button>
            <el-button v-if="row.status === 'S0'" link type="primary" size="small" @click="goEdit(row.id)">编辑</el-button>
            <el-button v-if="row.status === 'S0'" link type="success" size="small" @click="handleSubmit(row.id)">提交</el-button>
            <el-button v-if="row.status === 'S1'" link type="warning" size="small" @click="handleWithdraw(row.id)">撤回</el-button>
            <el-popconfirm v-if="row.status === 'S0' || row.status === 'S4'" title="确认删除此申请？" @confirm="handleDelete(row.id)">
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
import { tyApplicationApi } from '@/api/ty'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

// 状态映射
const statusMap = { S0: '草稿', S1: '待审', S2: '审批中', S3: '通过', S4: '驳回' }
const statusType = { S0: 'info', S1: 'warning', S2: '', S3: 'success', S4: 'danger' }

// 列表数据
const list = ref([])
const loading = ref(false)
const filterStatus = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

// 获取列表
async function fetchList() {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value
    }
    if (filterStatus.value) params.status = filterStatus.value
    // 绑定当前登录用户的学生ID，确保只查"我的"入团申请
    if (authStore.user?.student_id) params.student_id = authStore.user.student_id
    const data = await tyApplicationApi.list(params)
    list.value = data.items || []
    total.value = data.total || 0
  } catch (e) {
    console.error('获取入团申请列表失败', e)
  } finally {
    loading.value = false
  }
}

// 跳转新增
function goCreate() {
  router.push('/ty/application/new')
}

// 跳转我的团员发展（学生视角）
function goMyDevelopment() {
  router.push('/mine/ty-development')
}

// 跳转编辑
function goEdit(id) {
  router.push(`/ty/application/${id}/edit`)
}

// 跳转详情
function goDetail(id) {
  router.push(`/ty/application/${id}`)
}

// 提交申请
async function handleSubmit(id) {
  try {
    await ElMessageBox.confirm('确认提交此申请？提交后将进入审批流程。', '提交确认')
    await tyApplicationApi.submit(id)
    ElMessage.success('提交成功')
    fetchList()
  } catch (e) {
    if (e !== 'cancel') {
      // 错误已由 http 拦截器处理
    }
  }
}

// 撤回申请
async function handleWithdraw(id) {
  try {
    const { value } = await ElMessageBox.prompt('请输入撤回原因', '撤回申请', {
      confirmButtonText: '确认撤回',
      cancelButtonText: '取消',
      inputPlaceholder: '请说明撤回原因'
    })
    await tyApplicationApi.withdraw(id, value || '')
    ElMessage.success('已撤回')
    fetchList()
  } catch (e) {
    if (e !== 'cancel') {
      // 错误已由 http 拦截器处理
    }
  }
}

// 删除申请
async function handleDelete(id) {
  try {
    await tyApplicationApi.delete(id)
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
/* .card-header / .filter-bar / .pagination-wrap 已在全局定义 */
.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}
</style>
