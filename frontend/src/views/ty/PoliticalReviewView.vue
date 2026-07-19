<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>政审管理</span>
          <el-button type="primary" @click="openCreateDialog">新增政审记录</el-button>
        </div>
      </template>

      <!-- 按发展对象筛选 -->
      <div class="filter-bar">
        <el-select
          v-model="filterDevelopmentId"
          placeholder="选择发展对象"
          clearable
          filterable
          style="width: 260px"
          @change="fetchList"
        >
          <el-option v-for="obj in developmentObjects" :key="obj.id" :label="`${obj.student_name}（${obj.biz_no}）`" :value="obj.id" />
        </el-select>
      </div>

      <!-- 结论统计 -->
      <div class="stats-bar">
        <div class="stat-item stat-item--pass">
          <span class="stat-item__bar"></span>
          <div class="stat-item__main">
            <div class="stat-item__label">通过</div>
            <div class="stat-item__value">{{ statPass }}</div>
          </div>
        </div>
        <div class="stat-item stat-item--basic">
          <span class="stat-item__bar"></span>
          <div class="stat-item__main">
            <div class="stat-item__label">基本合格</div>
            <div class="stat-item__value">{{ statBasicPass }}</div>
          </div>
        </div>
        <div class="stat-item stat-item--fail">
          <span class="stat-item__bar"></span>
          <div class="stat-item__main">
            <div class="stat-item__label">不合格</div>
            <div class="stat-item__value">{{ statFail }}</div>
          </div>
        </div>
      </div>

      <el-table :data="list" stripe v-loading="loading" class="rd-table" table-layout="auto">
        <el-table-column prop="target_name" label="审查对象姓名" min-width="140" />
        <el-table-column prop="target_relation" label="与本人关系" min-width="160">
          <template #default="{ row }">
            {{ relationMap[row.target_relation] }}
          </template>
        </el-table-column>
        <el-table-column prop="method" label="审查方式" min-width="120">
          <template #default="{ row }">
            {{ row.method === 'letter' ? '函调' : '面谈' }}
          </template>
        </el-table-column>
        <el-table-column prop="conclusion" label="审查结论" min-width="130">
          <template #default="{ row }">
            <el-tag :type="conclusionTypeMap[row.conclusion]" size="small">
              {{ conclusionMap[row.conclusion] }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="document_path" label="材料路径" min-width="200" show-overflow-tooltip />
        <el-table-column prop="is_extend_3m" label="延长考察3月" min-width="140">
          <template #default="{ row }">
            <el-tag :type="row.is_extend_3m ? 'warning' : 'info'" size="small">
              {{ row.is_extend_3m ? '是' : '否' }}
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

    <!-- 新增政审记录弹窗 -->
    <el-dialog v-model="createDialogVisible" title="新增政审记录" width="580px" destroy-on-close>
      <el-form ref="createFormRef" :model="createForm" :rules="createFormRules" label-width="120px">
        <el-form-item label="发展对象" prop="development_id">
          <el-select v-model="createForm.development_id" placeholder="请选择发展对象" style="width: 100%" filterable>
            <el-option v-for="obj in developmentObjects" :key="obj.id" :label="`${obj.student_name}（${obj.biz_no}）`" :value="obj.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="审查对象关系" prop="target_relation">
          <el-radio-group v-model="createForm.target_relation">
            <el-radio value="self">本人</el-radio>
            <el-radio value="parent">父母/监护人</el-radio>
            <el-radio value="spouse">配偶</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="对象姓名" prop="target_name">
          <el-input v-model="createForm.target_name" placeholder="请输入审查对象姓名" />
        </el-form-item>
        <el-form-item label="审查方式" prop="method">
          <el-radio-group v-model="createForm.method">
            <el-radio value="letter">函调</el-radio>
            <el-radio value="interview">面谈</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="审查结论" prop="conclusion">
          <el-radio-group v-model="createForm.conclusion">
            <el-radio value="pass">通过</el-radio>
            <el-radio value="basic_pass">基本合格</el-radio>
            <el-radio value="fail">不合格</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="材料上传" prop="document_path">
          <div style="display: flex; align-items: center; gap: 8px; flex-wrap: wrap;">
            <el-upload
              action="#"
              :auto-upload="false"
              :on-change="handleDocumentChange"
              :on-remove="handleDocumentRemove"
              :file-list="documentFileList"
              accept=".jpg,.jpeg,.png,.pdf"
              :limit="1"
            >
              <el-button type="primary" size="small">选择文件</el-button>
            </el-upload>
            <span style="color: var(--sh-text-placeholder); font-size: 12px;">支持图片/PDF格式，不超过50MB</span>
          </div>
        </el-form-item>
        <el-form-item label="延长考察" prop="is_extend_3m">
          <el-switch v-model="createForm.is_extend_3m" :active-value="1" :inactive-value="0" />
          <span class="form-tip">如需延长3个月考察期请开启</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :loading="createSaving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { tyPoliticalReviewApi, tyDevelopmentObjectApi } from '@/api/ty'
import { fileApi } from '@/api/file'

// 映射
const relationMap = { self: '本人', parent: '父母/监护人', spouse: '配偶' }
const conclusionMap = { pass: '通过', basic_pass: '基本合格', fail: '不合格' }
const conclusionTypeMap = { pass: 'success', basic_pass: 'warning', fail: 'danger' }

// 列表数据
const list = ref([])
const loading = ref(false)
const filterDevelopmentId = ref(null)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

// 发展对象下拉
const developmentObjects = ref([])

// 统计
const statPass = computed(() => list.value.filter(r => r.conclusion === 'pass').length)
const statBasicPass = computed(() => list.value.filter(r => r.conclusion === 'basic_pass').length)
const statFail = computed(() => list.value.filter(r => r.conclusion === 'fail').length)
const passPct = computed(() => total.value ? Math.round((statPass.value / total.value) * 100) : 0)
const basicPct = computed(() => total.value ? Math.round((statBasicPass.value / total.value) * 100) : 0)
const failPct = computed(() => total.value ? Math.round((statFail.value / total.value) * 100) : 0)

// 创建弹窗
const createDialogVisible = ref(false)
const createSaving = ref(false)
const createFormRef = ref()
const createForm = ref({
  development_id: null,
  target_relation: 'self',
  target_name: '',
  method: 'letter',
  conclusion: 'pass',
  document_path: '',
  is_extend_3m: 0
})
const createFormRules = {
  development_id: [{ required: true, message: '请选择发展对象', trigger: 'change' }],
  target_relation: [{ required: true, message: '请选择关系', trigger: 'change' }],
  target_name: [{ required: true, message: '请输入对象姓名', trigger: 'blur' }],
  method: [{ required: true, message: '请选择审查方式', trigger: 'change' }],
  conclusion: [{ required: true, message: '请选择审查结论', trigger: 'change' }]
}

// 材料上传
const documentFileList = ref([])
const uploadingDocument = ref(false)

// 材料文件选择
async function handleDocumentChange(file) {
  const isImage = file.raw.type.startsWith('image/')
  const isPdf = file.raw.type === 'application/pdf'
  if (!isImage && !isPdf) {
    ElMessage.error('仅支持图片（jpg/png）或 PDF 格式')
    documentFileList.value = []
    return false
  }
  if (file.raw.size > 50 * 1024 * 1024) {
    ElMessage.error('文件大小不能超过 50MB')
    documentFileList.value = []
    return false
  }
  uploadingDocument.value = true
  try {
    const formData = new FormData()
    formData.append('file', file.raw)
    formData.append('module', 'ty')
    formData.append('biz_type', 'political_review')
    const res = await fileApi.upload(formData)
    const fileKey = res?.key || res?.data?.key
    if (fileKey) {
      createForm.value.document_path = fileKey
      documentFileList.value = [{ name: file.name, url: URL.createObjectURL(file.raw) }]
      ElMessage.success('文件上传成功')
    } else {
      ElMessage.error('文件上传失败：未返回文件标识')
      documentFileList.value = []
    }
  } catch (e) {
    ElMessage.error('文件上传失败')
    documentFileList.value = []
  } finally {
    uploadingDocument.value = false
  }
}

// 材料文件移除
function handleDocumentRemove() {
  createForm.value.document_path = ''
  documentFileList.value = []
}

// 获取列表
async function fetchList() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (filterDevelopmentId.value) params.development_id = filterDevelopmentId.value
    const data = await tyPoliticalReviewApi.list(params)
    list.value = data.items || []
    total.value = data.total || 0
  } catch (e) {
    console.error('获取政审记录列表失败', e)
  } finally {
    loading.value = false
  }
}

// 获取发展对象列表（用于筛选和新建）
async function fetchDevelopmentObjects() {
  try {
    const data = await tyDevelopmentObjectApi.list({ page_size: 200 })
    developmentObjects.value = data.items || []
  } catch (e) {
    console.error('获取发展对象列表失败', e)
  }
}

function openCreateDialog() {
  createForm.value = {
    development_id: null,
    target_relation: 'self',
    target_name: '',
    method: 'letter',
    conclusion: 'pass',
    document_path: '',
    is_extend_3m: 0
  }
  documentFileList.value = []
  // 如果已选了筛选，自动填充
  if (filterDevelopmentId.value) {
    createForm.value.development_id = filterDevelopmentId.value
  }
  createDialogVisible.value = true
}

async function handleCreate() {
  try { await createFormRef.value.validate() } catch { return }
  if (uploadingDocument.value) { ElMessage.warning('文件正在上传中，请稍候'); return }
  createSaving.value = true
  try {
    await tyPoliticalReviewApi.create(createForm.value)
    ElMessage.success('政审记录已保存')
    createDialogVisible.value = false
    fetchList()
  } catch (e) {} finally { createSaving.value = false }
}

onMounted(() => {
  fetchList()
  fetchDevelopmentObjects()
})
</script>

<style scoped>
/* .card-header / .filter-bar / .pagination-wrap / .form-tip 已在全局定义 */

/* 强制去掉 el-card 的所有阴影（包括 hover 时的） */
:deep(.el-card),
:deep(.el-card.is-always-shadow),
:deep(.el-card.is-hover-shadow):hover,
:deep(.el-card__header),
:deep(.el-card__body) {
  box-shadow: none !important;
}

/* 紧贴上下间距 */
:deep(.el-card__body) {
  padding: 16px 20px !important;
}
:deep(.filter-bar) {
  margin-bottom: 0 !important;
}
.stats-bar {
  display: flex;
  gap: 12px;
  margin-top: 0 !important;
  margin-bottom: 0 !important;
  background: transparent;
  border: 0;
  box-shadow: none;
}
.stat-item {
  flex: 1;
  position: relative;
  display: flex;
  align-items: stretch;
  border: 1px solid var(--sh-border-lighter);
  border-radius: var(--sh-radius-md);
  overflow: hidden;
  min-height: 48px;
  box-shadow: none;
}
.stat-item__bar {
  width: 3px;
  flex-shrink: 0;
}
.stat-item__main {
  flex: 1;
  padding: 8px 14px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 2px;
}
.stat-item__label {
  font-size: 13px;
  color: var(--sh-text-secondary, #606266);
  letter-spacing: 0.3px;
}
.stat-item__value {
  font-size: 22px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  line-height: 1.1;
}
.stat-item--pass { background: rgba(103, 194, 58, 0.08); }
.stat-item--pass .stat-item__bar { background: #67c23a; }
.stat-item--pass .stat-item__value { color: #67c23a; }
.stat-item--basic { background: rgba(230, 163, 60, 0.08); }
.stat-item--basic .stat-item__bar { background: #e6a23c; }
.stat-item--basic .stat-item__value { color: #e6a23c; }
.stat-item--fail { background: rgba(245, 108, 108, 0.08); }
.stat-item--fail .stat-item__bar { background: #f56c6c; }
.stat-item--fail .stat-item__value { color: #f56c6c; }

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
</style>
