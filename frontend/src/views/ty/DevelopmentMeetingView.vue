<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>发展大会</span>
          <el-button type="primary" @click="openCreateDialog">召开发展大会</el-button>
        </div>
      </template>

      <el-table :data="list" stripe v-loading="loading" class="rd-table" table-layout="auto">
        <el-table-column prop="biz_no" label="业务编号" min-width="200" />
        <el-table-column prop="student_name" label="申请人" min-width="120" />
        <el-table-column prop="meeting_at" label="会议时间" min-width="200">
          <template #default="{ row }">{{ formatDateTime(row.meeting_at) }}</template>
        </el-table-column>
        <el-table-column label="到会情况" min-width="140">
          <template #default="{ row }">
            {{ row.actual_count }} / {{ row.expected_count }}
          </template>
        </el-table-column>
        <el-table-column label="票数明细" min-width="280">
          <template #default="{ row }">
            赞成 {{ row.approve_count }} / 反对 {{ row.against_count }} / 弃权 {{ row.abstain_count }}
          </template>
        </el-table-column>
        <el-table-column prop="decision" label="决议" min-width="110">
          <template #default="{ row }">
            <el-tag :type="row.decision === 'pass' ? 'success' : 'danger'" size="small">
              {{ row.decision === 'pass' ? '通过' : '不通过' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" min-width="200">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
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

    <!-- 创建发展大会弹窗 -->
    <el-dialog v-model="createDialogVisible" title="召开发展大会" width="800px" destroy-on-close>
      <el-alert
        title="发展大会通过后，将自动为该同志创建团员花名册记录，进入预备团员阶段。"
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 20px"
      />

      <el-form ref="createFormRef" :model="createForm" :rules="createFormRules" label-width="130px">
        <el-form-item label="关联发展对象" prop="development_id">
          <el-select v-model="createForm.development_id" placeholder="请选择待发展对象" style="width: 100%" filterable>
            <el-option v-for="obj in devObjects" :key="obj.id" :label="`${obj.student_name}（${obj.biz_no}）`" :value="obj.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="会议时间" prop="meeting_at">
          <el-date-picker
            v-model="createForm.meeting_at"
            type="datetime"
            value-format="YYYY-MM-DD HH:mm:ss"
            placeholder="选择会议时间"
            style="width: 100%"
          />
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="应到人数" prop="expected_count">
              <el-input-number v-model="createForm.expected_count" :min="1" :max="999" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="实到人数" prop="actual_count">
              <el-input-number v-model="createForm.actual_count" :min="0" :max="999" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">投票统计</el-divider>

        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="赞成票数" prop="approve_count">
              <el-input-number v-model="createForm.approve_count" :min="0" :max="999" style="width: 180px" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="反对票数" prop="against_count">
              <el-input-number v-model="createForm.against_count" :min="0" :max="999" style="width: 180px" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="弃权票数" prop="abstain_count">
              <el-input-number v-model="createForm.abstain_count" :min="0" :max="999" style="width: 180px" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">决议信息</el-divider>

        <el-form-item label="决议结果" prop="decision">
          <el-radio-group v-model="createForm.decision">
            <el-radio value="pass">通过（接收为预备团员）</el-radio>
            <el-radio value="reject">不通过</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="入团志愿书" prop="volunteer_form_path">
          <div style="display: flex; align-items: center; gap: 8px; flex-wrap: wrap;">
            <el-upload
              action="#"
              :auto-upload="false"
              :on-change="handleVolunteerFormChange"
              :on-remove="handleVolunteerFormRemove"
              :file-list="volunteerFormFileList"
              accept=".jpg,.jpeg,.png,.pdf"
              :limit="1"
            >
              <el-button type="primary" size="small">选择文件</el-button>
            </el-upload>
            <span style="color: var(--sh-text-placeholder); font-size: 12px;">支持图片/PDF格式，不超过50MB</span>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :loading="createSaving">提交大会记录</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { tyDevelopmentMeetingApi, tyDevelopmentObjectApi } from '@/api/ty'
import { fileApi } from '@/api/file'
import { formatDateTime } from '@/utils/datetime'

// 列表数据
const list = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

// 发展对象下拉（S4 已通过的）
const devObjects = ref([])

// 创建弹窗
const createDialogVisible = ref(false)
const createSaving = ref(false)
const createFormRef = ref()
const createForm = ref({
  development_id: null,
  meeting_at: '',
  expected_count: null,
  actual_count: null,
  approve_count: 0,
  against_count: 0,
  abstain_count: 0,
  decision: '',
  volunteer_form_path: ''
})
const createFormRules = {
  development_id: [{ required: true, message: '请选择发展对象', trigger: 'change' }],
  meeting_at: [{ required: true, message: '请选择会议时间', trigger: 'change' }],
  expected_count: [{ required: true, message: '请输入应到人数', trigger: 'blur' }],
  actual_count: [{ required: true, message: '请输入实到人数', trigger: 'blur' }],
  decision: [{ required: true, message: '请选择决议结果', trigger: 'change' }],
  volunteer_form_path: [{ required: true, message: '请上传入团志愿书', trigger: 'change' }]
}

// 入团志愿书上传
const volunteerFormFileList = ref([])
const uploadingVolunteerForm = ref(false)

// 入团志愿书文件选择
async function handleVolunteerFormChange(file) {
  const isImage = file.raw.type.startsWith('image/')
  const isPdf = file.raw.type === 'application/pdf'
  if (!isImage && !isPdf) {
    ElMessage.error('仅支持图片（jpg/png）或 PDF 格式')
    volunteerFormFileList.value = []
    return false
  }
  if (file.raw.size > 50 * 1024 * 1024) {
    ElMessage.error('文件大小不能超过 50MB')
    volunteerFormFileList.value = []
    return false
  }
  uploadingVolunteerForm.value = true
  try {
    const formData = new FormData()
    formData.append('file', file.raw)
    formData.append('module', 'ty')
    formData.append('biz_type', 'volunteer_form')
    const res = await fileApi.upload(formData)
    const fileKey = res?.key || res?.data?.key
    if (fileKey) {
      createForm.value.volunteer_form_path = fileKey
      volunteerFormFileList.value = [{ name: file.name, url: URL.createObjectURL(file.raw) }]
      ElMessage.success('文件上传成功')
    } else {
      ElMessage.error('文件上传失败：未返回文件标识')
      volunteerFormFileList.value = []
    }
  } catch (e) {
    ElMessage.error('文件上传失败')
    volunteerFormFileList.value = []
  } finally {
    uploadingVolunteerForm.value = false
  }
}

// 入团志愿书文件移除
function handleVolunteerFormRemove() {
  createForm.value.volunteer_form_path = ''
  volunteerFormFileList.value = []
}

// 获取列表
async function fetchList() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    const data = await tyDevelopmentMeetingApi.list(params)
    list.value = data.items || []
    total.value = data.total || 0
  } catch (e) {
    console.error('获取发展大会列表失败', e)
  } finally {
    loading.value = false
  }
}

// 获取可发展对象（S4 状态）
async function fetchDevObjects() {
  try {
    const data = await tyDevelopmentObjectApi.list({ status: 'S4', page_size: 200 })
    devObjects.value = data.items || []
  } catch (e) {
    console.error('获取发展对象列表失败', e)
  }
}

function openCreateDialog() {
  createForm.value = {
    development_id: null,
    meeting_at: '',
    expected_count: null,
    actual_count: null,
    approve_count: 0,
    against_count: 0,
    abstain_count: 0,
    decision: '',
    volunteer_form_path: ''
  }
  volunteerFormFileList.value = []
  createDialogVisible.value = true
}

async function handleCreate() {
  try { await createFormRef.value.validate() } catch { return }
  if (uploadingVolunteerForm.value) { ElMessage.warning('文件正在上传中，请稍候'); return }

  // 校验：决议为通过时给出联动提示
  if (createForm.value.decision === 'pass') {
    try {
      await ElMessageBox.confirm(
        '决议为「通过」后，系统将自动为该同志创建团员花名册记录，进入预备团员阶段。是否确认？',
        '通过确认',
        { type: 'warning' }
      )
    } catch {
      return // 用户取消
    }
  }

  createSaving.value = true
  try {
    await tyDevelopmentMeetingApi.create(createForm.value)
    ElMessage.success('发展大会记录已提交')
    createDialogVisible.value = false
    fetchList()
  } catch (e) {} finally { createSaving.value = false }
}

onMounted(() => {
  fetchList()
  fetchDevObjects()
})
</script>

<style scoped>
/* .card-header / .pagination-wrap 已在全局定义 */

/* 让 el-table 整体铺满容器宽度，避免表头背景色在右侧断开 */
.rd-table,
.rd-table :deep(.el-table__inner-wrapper),
.rd-table :deep(.el-table__header-wrapper),
.rd-table :deep(.el-table__body-wrapper) {
  width: 100% !important;
}
.rd-table :deep(.el-table__header-wrapper table),
.rd-table :deep(.el-table__body-wrapper table) {
  width: 100% !important;
}

/* 修复 el-input-number 按钮样式 - 确保加减按钮颜色正常 */
:deep(.el-input-number .el-input-number__decrease),
:deep(.el-input-number .el-input-number__increase) {
  color: #606266 !important;
}
:deep(.el-input-number .el-input-number__decrease:hover),
:deep(.el-input-number .el-input-number__increase:hover) {
  color: #409eff !important;
}
:deep(.el-input-number .el-input-number__decrease.is-disabled),
:deep(.el-input-number .el-input-number__increase.is-disabled) {
  color: #c0c4cc !important;
  cursor: not-allowed !important;
}
/* 修复 el-input-number 内部输入区域被全局样式挤压的问题 */
:deep(.el-input-number .el-input) {
  flex: 1 !important;
  min-width: 40px !important;
}
:deep(.el-input-number .el-input__wrapper) {
  min-width: 40px !important;
  width: 100% !important;
}
</style>
