<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>发展对象管理</span>
          <el-button type="primary" @click="openCreateDialog">提交发展对象申请</el-button>
        </div>
      </template>

      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-select v-model="filterStatus" placeholder="状态筛选" clearable style="width: 160px" @change="fetchList">
          <el-option label="待提交" value="S0" />
          <el-option label="公示中" value="S1" />
          <el-option label="待审批" value="S2" />
          <el-option label="政审中" value="S3" />
          <el-option label="已通过" value="S4" />
        </el-select>
      </div>

      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="biz_no" label="业务编号" width="170" />
        <el-table-column prop="student_name" label="申请人" width="100" />
        <el-table-column prop="branch_name" label="团支部" min-width="140" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTypeMap[row.status]" size="small">
              {{ statusTextMap[row.status] || row.status_text || row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" min-width="180">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.status === 'S0'" link type="primary" size="small" @click="handleSubmit(row.id)">提交申请</el-button>
            <el-button v-if="row.status === 'S1'" link type="warning" size="small" @click="handlePublicize(row.id)">公示</el-button>
            <el-button v-if="row.status === 'S2'" link type="success" size="small" @click="openApproveDialog(row)">审批</el-button>
            <el-button link type="primary" size="small" @click="showDetail(row)">查看详情</el-button>
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

    <!-- 创建发展对象申请弹窗 -->
    <el-dialog v-model="createDialogVisible" title="提交发展对象申请" width="650px" destroy-on-close>
      <el-form ref="createFormRef" :model="createForm" :rules="createFormRules" label-width="120px">
        <el-form-item label="关联申请" prop="application_id">
          <el-select v-model="createForm.application_id" placeholder="请选择已通过的入团申请" style="width: 100%" filterable>
            <el-option v-for="app in passedApps" :key="app.id" :label="`${app.student_name}（${app.biz_no}）`" :value="app.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="团课证书编号" prop="course_cert_no">
          <el-select v-model="createForm.course_cert_no" placeholder="请选择该学生的团课结业证书" style="width: 100%" filterable clearable @change="onCourseCertChange">
            <el-option v-for="c in studentCourseOptions" :key="c.certificate_no" :label="`${c.course_name} · ${c.certificate_no} · 成绩${c.score ?? '—'}`" :value="c.certificate_no" />
          </el-select>
        </el-form-item>
        <el-form-item label="培养联系人意见" prop="mentor_opinion">
          <el-input v-model="createForm.mentor_opinion" type="textarea" :rows="3" placeholder="培养联系人对该同志的评价意见" />
        </el-form-item>
        <el-form-item label="辅导员意见" prop="counselor_opinion">
          <el-input v-model="createForm.counselor_opinion" type="textarea" :rows="3" placeholder="辅导员的推荐意见" />
        </el-form-item>
        <el-form-item label="自传材料" prop="autobiography_path">
          <div style="display: flex; align-items: center; gap: 8px; flex-wrap: wrap;">
            <el-upload
              action="#"
              :auto-upload="false"
              :on-change="handleAutobiographyChange"
              :on-remove="handleAutobiographyRemove"
              :file-list="autobiographyFileList"
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
        <el-button type="primary" @click="handleCreate" :loading="createSaving">提交申请</el-button>
      </template>
    </el-dialog>

    <!-- 审批弹窗 -->
    <el-dialog v-model="approveDialogVisible" title="审批发展对象" width="500px" destroy-on-close>
      <el-form ref="approveFormRef" :model="approveForm" :rules="approveFormRules" label-width="100px">
        <el-form-item label="审批结果" prop="decision">
          <el-radio-group v-model="approveForm.decision">
            <el-radio value="pass">通过</el-radio>
            <el-radio value="reject">驳回</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="审批意见" prop="opinion">
          <el-input v-model="approveForm.opinion" type="textarea" :rows="4" placeholder="请输入审批意见" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="approveDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleApprove" :loading="approveSaving">确认审批</el-button>
      </template>
    </el-dialog>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="发展对象详情" width="800px" destroy-on-close>
      <el-descriptions :column="2" border v-if="currentDetail">
        <el-descriptions-item label="业务编号">{{ currentDetail.biz_no }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTypeMap[currentDetail.status]" size="small">
            {{ statusTextMap[currentDetail.status] || currentDetail.status_text }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="申请人">{{ currentDetail.student_name }}</el-descriptions-item>
        <el-descriptions-item label="团支部">{{ currentDetail.branch_name }}</el-descriptions-item>
        <el-descriptions-item label="团课证书编号">{{ currentDetail.course_cert_no }}</el-descriptions-item>
        <el-descriptions-item label="培养联系人意见">{{ currentDetail.mentor_opinion }}</el-descriptions-item>
        <el-descriptions-item label="辅导员意见">{{ currentDetail.counselor_opinion }}</el-descriptions-item>
        <el-descriptions-item label="公示开始">{{ currentDetail.public_start || '—' }}</el-descriptions-item>
        <el-descriptions-item label="公示结束">{{ currentDetail.public_end || '—' }}</el-descriptions-item>
        <el-descriptions-item label="自传材料">{{ currentDetail.autobiography_path || '—' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDateTime(currentDetail.created_at) }}</el-descriptions-item>
      </el-descriptions>

      <el-divider content-position="left">政审记录</el-divider>
      <el-table :data="politicalReviews" stripe size="small" v-loading="reviewLoading">
        <el-table-column prop="target_name" label="审查对象" width="120" />
        <el-table-column prop="target_relation" label="关系" width="100">
          <template #default="{ row }">
            {{ relationMap[row.target_relation] }}
          </template>
        </el-table-column>
        <el-table-column prop="method" label="方式" width="100">
          <template #default="{ row }">
            {{ row.method === 'letter' ? '函调' : '面谈' }}
          </template>
        </el-table-column>
        <el-table-column prop="conclusion" label="结论" width="100">
          <template #default="{ row }">
            <el-tag :type="conclusionTypeMap[row.conclusion]" size="small">
              {{ conclusionMap[row.conclusion] }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="is_extend_3m" label="延长考察" width="90">
          <template #default="{ row }">
            {{ row.is_extend_3m ? '是' : '否' }}
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { tyDevelopmentObjectApi, tyPoliticalReviewApi, tyApplicationApi, tyCourseRecordApi } from '@/api/ty'
import { fileApi } from '@/api/file'
import { formatDateTime } from '@/utils/datetime'

// 状态映射
const statusTextMap = { S0: '待提交', S1: '公示中', S2: '待审批', S3: '政审中', S4: '已通过' }
const statusTypeMap = { S0: 'info', S1: 'warning', S2: '', S3: 'warning', S4: 'success' }

// 政审结论映射
const relationMap = { self: '本人', parent: '父母/监护人', spouse: '配偶' }
const conclusionMap = { pass: '通过', basic_pass: '基本合格', fail: '不合格' }
const conclusionTypeMap = { pass: 'success', basic_pass: 'warning', fail: 'danger' }

// 列表数据
const list = ref([])
const loading = ref(false)
const filterStatus = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

// 详情相关
const detailVisible = ref(false)
const currentDetail = ref(null)
const politicalReviews = ref([])
const reviewLoading = ref(false)

// 创建弹窗
const createDialogVisible = ref(false)
const createSaving = ref(false)
const createFormRef = ref()
const createForm = ref({ application_id: null, course_cert_no: '', mentor_opinion: '', counselor_opinion: '', autobiography_path: '' })
const createFormRules = {
  application_id: [{ required: true, message: '请选择入团申请', trigger: 'change' }],
  course_cert_no: [{ required: true, message: '请输入团课证书编号', trigger: 'blur' }]
}
const passedApps = ref([])

// 自传材料上传
const autobiographyFileList = ref([])
const uploadingAutobiography = ref(false)

// 自传材料文件选择
async function handleAutobiographyChange(file) {
  const isImage = file.raw.type.startsWith('image/')
  const isPdf = file.raw.type === 'application/pdf'
  if (!isImage && !isPdf) {
    ElMessage.error('仅支持图片（jpg/png）或 PDF 格式')
    autobiographyFileList.value = []
    return false
  }
  if (file.raw.size > 50 * 1024 * 1024) {
    ElMessage.error('文件大小不能超过 50MB')
    autobiographyFileList.value = []
    return false
  }
  // 上传文件
  uploadingAutobiography.value = true
  try {
    const formData = new FormData()
    formData.append('file', file.raw)
    formData.append('module', 'ty')
    formData.append('biz_type', 'autobiography')
    const res = await fileApi.upload(formData)
    const fileKey = res?.key || res?.data?.key
    if (fileKey) {
      createForm.value.autobiography_path = fileKey
      autobiographyFileList.value = [{ name: file.name, url: URL.createObjectURL(file.raw) }]
      ElMessage.success('文件上传成功')
    } else {
      ElMessage.error('文件上传失败：未返回文件标识')
      autobiographyFileList.value = []
    }
  } catch (e) {
    ElMessage.error('文件上传失败')
    autobiographyFileList.value = []
  } finally {
    uploadingAutobiography.value = false
  }
}

// 自传材料文件移除
function handleAutobiographyRemove() {
  createForm.value.autobiography_path = ''
  autobiographyFileList.value = []
}

// 学生的团课证书选项
const studentCourseOptions = ref([])

// 选择关联申请后，加载该学生的团课证书编号列表
watch(() => createForm.value.application_id, async (appId) => {
  studentCourseOptions.value = []
  createForm.value.course_cert_no = ''
  if (!appId) return
  const app = passedApps.value.find(a => a.id === appId)
  if (!app || !app.student_id) return
  try {
    const data = await tyCourseRecordApi.list({ student_id: app.student_id, page_size: 100 })
    const courses = (data.items || []).filter(c => c.certificate_no)
    studentCourseOptions.value = courses
  } catch (e) {
    console.error('加载团课证书列表失败', e)
  }
})

function onCourseCertChange() {
  // 选择证书后无需额外操作
}

// 审批弹窗
const approveDialogVisible = ref(false)
const approveSaving = ref(false)
const approveFormRef = ref()
const approveForm = ref({ decision: '', opinion: '' })
const approveFormRules = {
  decision: [{ required: true, message: '请选择审批结果', trigger: 'change' }],
  opinion: [{ required: true, message: '请输入审批意见', trigger: 'blur' }]
}
const approveTargetId = ref(null)

// 获取列表
async function fetchList() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (filterStatus.value) params.status = filterStatus.value
    const data = await tyDevelopmentObjectApi.list(params)
    list.value = data.items || []
    total.value = data.total || 0
  } catch (e) {
    console.error('获取发展对象列表失败', e)
  } finally {
    loading.value = false
  }
}

// 获取已通过申请（用于下拉）
async function fetchPassedApps() {
  try {
    const data = await tyApplicationApi.list({ status: 'S3', page_size: 200 })
    passedApps.value = data.items || []
  } catch (e) {
    console.error('获取已通过申请失败', e)
  }
}

function openCreateDialog() {
  createForm.value = { application_id: null, course_cert_no: '', mentor_opinion: '', counselor_opinion: '', autobiography_path: '' }
  createDialogVisible.value = true
}

async function handleCreate() {
  try { await createFormRef.value.validate() } catch { return }
  if (uploadingAutobiography.value) { ElMessage.warning('文件正在上传中，请稍候'); return }
  createSaving.value = true
  try {
    await tyDevelopmentObjectApi.create(createForm.value)
    ElMessage.success('发展对象申请已提交')
    createDialogVisible.value = false
    fetchList()
  } catch (e) {} finally { createSaving.value = false }
}

async function handleSubmit(id) {
  try {
    await ElMessageBox.confirm('确认提交此发展对象申请？', '提交确认')
    // 提交即进入公示阶段
    await tyDevelopmentObjectApi.publicize(id)
    ElMessage.success('已提交并进入公示阶段')
    fetchList()
  } catch (e) { if (e !== 'cancel') {} }
}

async function handlePublicize(id) {
  try {
    await ElMessageBox.confirm('确认开始公示？', '公示确认')
    await tyDevelopmentObjectApi.publicize(id)
    ElMessage.success('公示已开启')
    fetchList()
  } catch (e) { if (e !== 'cancel') {} }
}

function openApproveDialog(row) {
  approveTargetId.value = row.id
  approveForm.value = { decision: '', opinion: '' }
  approveDialogVisible.value = true
}

async function handleApprove() {
  try { await approveFormRef.value.validate() } catch { return }
  approveSaving.value = true
  try {
    await tyDevelopmentObjectApi.approve(approveTargetId.value, approveForm.value)
    ElMessage.success('审批完成')
    approveDialogVisible.value = false
    fetchList()
  } catch (e) {} finally { approveSaving.value = false }
}

async function showDetail(row) {
  currentDetail.value = row
  detailVisible.value = true
  reviewLoading.value = true
  try {
    const data = await tyPoliticalReviewApi.list({ development_id: row.id })
    politicalReviews.value = data.items || []
  } catch (e) {
    console.error('获取政审记录失败', e)
  } finally {
    reviewLoading.value = false
  }
}

onMounted(() => {
  fetchList()
  fetchPassedApps()
})
</script>

<style scoped>
/* .card-header / .filter-bar / .pagination-wrap 已在全局定义 */
</style>
