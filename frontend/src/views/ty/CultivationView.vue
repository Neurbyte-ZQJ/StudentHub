<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>培养记录</span>
        </div>
      </template>

      <el-tabs v-model="activeTab" type="border-card">
        <!-- 培养联系人 Tab -->
        <el-tab-pane label="培养联系人" name="link">
          <div class="tab-header">
            <h4>当前培养联系人</h4>
            <el-button type="primary" size="small" @click="openLinkDialog">分配联系人</el-button>
          </div>

          <el-table :data="linkList" stripe v-loading="linkLoading" class="rd-table" table-layout="auto">
            <el-table-column prop="mentor_name" label="联系人姓名" min-width="120" />
            <el-table-column prop="mentor_type" label="类型" min-width="100">
              <template #default="{ row }">
                {{ row.mentor_type === 'league_member' ? '团员' : '党员' }}
              </template>
            </el-table-column>
            <el-table-column prop="start_at" label="开始时间" min-width="180">
              <template #default="{ row }">{{ formatDateTime(row.start_at) }}</template>
            </el-table-column>
            <el-table-column prop="end_at" label="结束时间" min-width="180">
              <template #default="{ row }">
                {{ row.end_at ? formatDateTime(row.end_at) : '—' }}
              </template>
            </el-table-column>
            <el-table-column prop="is_active" label="状态" min-width="100">
              <template #default="{ row }">
                <el-tag :type="row.is_active ? 'success' : 'info'" size="small">
                  {{ row.is_active ? '在任' : '已结束' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" min-width="100">
              <template #default="{ row }">
                <el-button v-if="row.is_active" link type="warning" size="small" @click="handleEndLink(row.id)">结束</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 培养记录 Tab -->
        <el-tab-pane label="培养记录" name="record">
          <div class="tab-header">
            <h4>月度/季度培养记录</h4>
            <el-button type="primary" size="small" @click="openRecordDialog">新增记录</el-button>
          </div>

          <el-table :data="recordList" stripe v-loading="recordLoading" class="rd-table" table-layout="auto">
            <el-table-column prop="biz_no" label="编号" min-width="180" />
            <el-table-column prop="student_no" label="学号" min-width="140">
              <template #default="{ row }">
                {{ row.student_no || '—' }}
              </template>
            </el-table-column>
            <el-table-column prop="student_name" label="姓名" min-width="120">
              <template #default="{ row }">
                {{ row.student_name || '—' }}
              </template>
            </el-table-column>
            <el-table-column prop="record_year" label="年份" min-width="80" />
            <el-table-column prop="record_month" label="月份" min-width="80" />
            <el-table-column prop="record_type" label="记录类型" min-width="100">
              <template #default="{ row }">
                {{ row.record_type === 'monthly' ? '月度' : '季度' }}
              </template>
            </el-table-column>
            <el-table-column prop="summary" label="培养总结" min-width="240" show-overflow-tooltip />
            <el-table-column prop="performance_score" label="综合评分" min-width="100" />
            <el-table-column prop="recorded_by_name" label="记录人" min-width="100" />
            <el-table-column prop="created_at" label="创建时间" min-width="180">
              <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="100" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="showRecordDetail(row)">查看</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 团课记录 Tab -->
        <el-tab-pane label="团课记录" name="course">
          <div class="tab-header">
            <h4>团课学习记录</h4>
            <el-button type="primary" size="small" @click="openCourseDialog">新增团课</el-button>
          </div>

          <el-table :data="courseList" stripe v-loading="courseLoading" class="rd-table" table-layout="auto">
            <el-table-column prop="student_no" label="学号" min-width="120" />
            <el-table-column prop="student_name" label="姓名" min-width="100" />
            <el-table-column prop="course_name" label="课程名称" min-width="180" />
            <el-table-column prop="semester" label="学期" min-width="180" />
            <el-table-column prop="study_at" label="学习时间" min-width="160">
              <template #default="{ row }">{{ row.study_at ? formatDate(row.study_at) : '—' }}</template>
            </el-table-column>
            <el-table-column prop="score" label="成绩" min-width="80">
              <template #default="{ row }">
                {{ row.score ?? '—' }}
              </template>
            </el-table-column>
            <el-table-column prop="certificate_no" label="结业证书编号" min-width="180" />
            <el-table-column prop="is_pass" label="状态" min-width="100">
              <template #default="{ row }">
                <el-tag :type="row.is_pass ? 'success' : 'info'" size="small">
                  {{ row.is_pass ? '已结业' : '未结业' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" min-width="120">
              <template #default="{ row }">
                <el-button v-if="!row.is_pass && canMarkPass" link type="success" size="small" @click="handlePassCourse(row.id)">标记结业</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 思想汇报 Tab -->
        <el-tab-pane label="思想汇报" name="thought">
          <div class="tab-header">
            <h4>季度思想汇报</h4>
            <el-button type="primary" size="small" @click="openThoughtDialog">新增汇报</el-button>
          </div>

          <el-table :data="thoughtList" stripe v-loading="thoughtLoading">
            <el-table-column prop="biz_no" label="编号" width="170" />
            <el-table-column prop="student_no" label="学号" width="130" />
            <el-table-column prop="student_name" label="汇报人" min-width="100" />
            <el-table-column prop="title" label="标题" min-width="180" show-overflow-tooltip />
            <el-table-column prop="quarter" label="季度" width="100" />
            <el-table-column prop="ai_similarity" label="AI相似度" width="100">
              <template #default="{ row }">
                {{ row.ai_similarity != null ? (row.ai_similarity * 100).toFixed(1) + '%' : '—' }}
              </template>
            </el-table-column>
            <el-table-column prop="is_qualified" label="是否合格" width="90">
              <template #default="{ row }">
                <el-tag :type="row.is_qualified ? 'success' : 'danger'" size="small">
                  {{ row.is_qualified ? '合格' : '不合格' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="提交时间" min-width="180">
              <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="100" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="showThoughtDetail(row)">查看</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 分配培养联系人弹窗（PRD §4.3.4：必须 2 位） -->
    <el-dialog v-model="linkDialogVisible" title="分配培养联系人" width="640px" destroy-on-close>
      <el-form ref="linkFormRef" :model="linkForm" :rules="linkFormRules" label-width="100px">
        <el-form-item label="关联申请" prop="application_id">
          <el-select v-model="linkForm.application_id" placeholder="请选择入团申请" style="width: 100%" filterable @change="onApplicationChange">
            <el-option v-for="app in applications" :key="app.id" :label="`${app.student_name}（${app.biz_no}）`" :value="app.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="开始时间" prop="start_at">
          <el-date-picker v-model="linkForm.start_at" type="date" value-format="YYYY-MM-DD" placeholder="选择开始时间" style="width: 100%" />
        </el-form-item>
        <el-divider content-position="left">第 1 位培养联系人</el-divider>
        <el-form-item label="联系人 A" :prop="`mentors.0.mentor_student_id`" :rules="mentorStudentRules(0)">
          <el-select v-model="linkForm.mentors[0].mentor_student_id" placeholder="请选择学生" style="width: 100%" filterable clearable>
            <el-option
              v-for="s in mentorCandidateOptions"
              :key="s.id"
              :label="`${s.name}（${s.student_no} · ${politicalStatusLabel(s.political_status)}）`"
              :value="s.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="政治面貌" :prop="`mentors.0.mentor_type`" :rules="mentorTypeRules(0)">
          <el-radio-group v-model="linkForm.mentors[0].mentor_type">
            <el-radio value="league_member">团员</el-radio>
            <el-radio value="party_member">党员</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-divider content-position="left">第 2 位培养联系人</el-divider>
        <el-form-item label="联系人 B" :prop="`mentors.1.mentor_student_id`" :rules="mentorStudentRules(1)">
          <el-select v-model="linkForm.mentors[1].mentor_student_id" placeholder="请选择学生" style="width: 100%" filterable clearable>
            <el-option
              v-for="s in mentorCandidateOptions"
              :key="s.id"
              :label="`${s.name}（${s.student_no} · ${politicalStatusLabel(s.political_status)}）`"
              :value="s.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="政治面貌" :prop="`mentors.1.mentor_type`" :rules="mentorTypeRules(1)">
          <el-radio-group v-model="linkForm.mentors[1].mentor_type">
            <el-radio value="league_member">团员</el-radio>
            <el-radio value="party_member">党员</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="linkDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreateLink" :loading="linkSaving">确认分配</el-button>
      </template>
    </el-dialog>

    <!-- 新增培养记录弹窗 -->
    <el-dialog v-model="recordDialogVisible" title="新增培养记录" width="600px" destroy-on-close>
      <el-form ref="recordFormRef" :model="recordForm" :rules="recordFormRules" label-width="100px">
        <el-form-item label="关联申请" prop="application_id">
          <el-select v-model="recordForm.application_id" placeholder="请选择入团申请" style="width: 100%">
            <el-option v-for="app in applications" :key="app.id" :label="`${app.student_name}`" :value="app.id" />
          </el-select>
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="年份" prop="record_year">
              <el-input-number v-model="recordForm.record_year" :min="2020" :max="2030" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="月份" prop="record_month">
              <el-input-number v-model="recordForm.record_month" :min="1" :max="12" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="记录类型" prop="record_type">
          <el-radio-group v-model="recordForm.record_type">
            <el-radio value="monthly">月度</el-radio>
            <el-radio value="quarterly">季度</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="培养总结" prop="summary">
          <el-input v-model="recordForm.summary" type="textarea" :rows="5" placeholder="请输入培养总结" />
        </el-form-item>
        <el-form-item label="综合评分" prop="performance_score">
          <el-input-number v-model="recordForm.performance_score" :min="0" :max="100" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="recordDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreateRecord" :loading="recordSaving">保存</el-button>
      </template>
    </el-dialog>

    <!-- 新增团课弹窗 -->
    <el-dialog v-model="courseDialogVisible" title="新增团课记录" width="550px" destroy-on-close>
      <el-form ref="courseFormRef" :model="courseForm" :rules="courseFormRules" label-width="100px">
        <el-form-item label="学生" prop="student_id">
          <!-- 学生本人：只读展示，不允许改 -->
          <template v-if="isStudentSelf">
            <el-input :value="currentStudentDisplay" disabled placeholder="当前账号已自动绑定为学生本人" />
          </template>
          <!-- 管理员/教师：可下拉选有入团申请的学生 -->
          <el-select v-else v-model="courseForm.student_id" placeholder="请选择学生" style="width: 100%" filterable>
            <el-option v-for="app in applications" :key="app.student_id" :label="`${app.student_name}（${app.student_no}）`" :value="app.student_id" />
          </el-select>
        </el-form-item>
        <el-form-item label="课程名称" prop="course_name">
          <el-input v-model="courseForm.course_name" placeholder="请输入团课名称" />
        </el-form-item>
        <el-form-item label="学期" prop="semester">
          <el-input v-model="courseForm.semester" placeholder="如：2025-2026学年第一学期" />
        </el-form-item>
        <el-form-item label="学习时间" prop="study_at">
          <el-date-picker v-model="courseForm.study_at" type="date" value-format="YYYY-MM-DD" placeholder="选择日期" style="width: 100%" />
        </el-form-item>
        <el-form-item label="证书编号" prop="certificate_no">
          <el-input v-model="courseForm.certificate_no" placeholder="结业证书编号（选填）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="courseDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreateCourse" :loading="courseSaving">保存</el-button>
      </template>
    </el-dialog>

    <!-- 新增思想汇报弹窗 -->
    <el-dialog v-model="thoughtDialogVisible" title="新增思想汇报" width="700px" destroy-on-close>
      <el-form ref="thoughtFormRef" :model="thoughtForm" :rules="thoughtFormRules" label-width="100px">
        <el-form-item label="关联申请" prop="application_id">
          <el-select v-model="thoughtForm.application_id" placeholder="请选择入团申请" style="width: 100%">
            <el-option v-for="app in applications" :key="app.id" :label="app.student_name" :value="app.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题" prop="title">
          <el-input v-model="thoughtForm.title" placeholder="请输入汇报标题" />
        </el-form-item>
        <el-form-item label="季度" prop="quarter">
          <el-select v-model="thoughtForm.quarter" placeholder="请选择季度" style="width: 100%">
            <el-option label="第一季度" value="Q1" />
            <el-option label="第二季度" value="Q2" />
            <el-option label="第三季度" value="Q3" />
            <el-option label="第四季度" value="Q4" />
          </el-select>
        </el-form-item>
        <el-form-item label="汇报内容" prop="content">
          <el-input
            v-model="thoughtForm.content"
            type="textarea"
            :rows="15"
            placeholder="请详细撰写思想汇报内容（不少于1000字）"
            show-word-limit
          />
          <div class="word-count" :class="{ warning: thoughtContentLength < 1000 }">
            已输入 {{ thoughtContentLength }} 字（最少 1000 字）
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="thoughtDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreateThought" :loading="thoughtSaving">提交汇报</el-button>
      </template>
    </el-dialog>

    <!-- 思想汇报详情弹窗 -->
    <el-dialog v-model="thoughtDetailVisible" title="思想汇报详情" width="700px" destroy-on-close>
      <el-descriptions :column="2" border v-if="currentThought">
        <el-descriptions-item label="标题" :span="2">{{ currentThought.title }}</el-descriptions-item>
        <el-descriptions-item label="季度">{{ currentThought.quarter }}</el-descriptions-item>
        <el-descriptions-item label="AI相似度">{{ currentThought.ai_similarity != null ? (currentThought.ai_similarity * 100).toFixed(1) + '%' : '未检测' }}</el-descriptions-item>
        <el-descriptions-item label="是否合格">
          <el-tag :type="currentThought.is_qualified ? 'success' : 'danger'" size="small">
            {{ currentThought.is_qualified ? '合格' : '不合格' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="提交时间">{{ formatDateTime(currentThought.created_at) }}</el-descriptions-item>
      </el-descriptions>
      <el-divider content-position="left">汇报正文</el-divider>
      <div class="thought-content">{{ currentThought?.content }}</div>
    </el-dialog>

    <!-- 培养记录详情弹窗 -->
    <el-dialog v-model="recordDetailVisible" title="培养记录详情" width="700px" destroy-on-close>
      <el-descriptions :column="2" border v-if="currentRecord">
        <el-descriptions-item label="编号" :span="2">{{ currentRecord.biz_no }}</el-descriptions-item>
        <el-descriptions-item label="学号">{{ currentRecord.student_no || '—' }}</el-descriptions-item>
        <el-descriptions-item label="姓名">{{ currentRecord.student_name || '—' }}</el-descriptions-item>
        <el-descriptions-item label="年份">{{ currentRecord.record_year }}</el-descriptions-item>
        <el-descriptions-item label="月份">{{ currentRecord.record_month }}</el-descriptions-item>
        <el-descriptions-item label="记录类型">
          {{ currentRecord.record_type === 'monthly' ? '月度' : '季度' }}
        </el-descriptions-item>
        <el-descriptions-item label="综合评分">{{ currentRecord.performance_score }}</el-descriptions-item>
        <el-descriptions-item label="记录人">{{ currentRecord.recorded_by_name }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDateTime(currentRecord.created_at) }}</el-descriptions-item>
      </el-descriptions>
      <el-divider content-position="left">培养总结</el-divider>
      <div class="thought-content">{{ currentRecord?.summary }}</div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  tyCultivationLinkApi,
  tyCultivationRecordApi,
  tyCourseRecordApi,
  tyThoughtReportApi,
  tyApplicationApi
} from '@/api/ty'
import { idxStudentApi } from '@/api/idx'
import { formatDateTime, formatDate } from '@/utils/datetime'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

// 当前登录用户
const currentUser = computed(() => authStore.user || {})
const currentUserRoles = computed(() => currentUser.value?.roles?.map(r => r.code) || [])
const currentStudentId = computed(() => currentUser.value?.student_id || null)
const currentStudentDisplay = computed(() => {
  const u = currentUser.value
  if (!u) return ''
  const no = u.student_no || ''
  const name = u.display_name || u.name || ''
  return no ? `${no} ${name}` : name || `student#${currentStudentId.value}`
})

// 是否学生身份（未兼教师/管理员）
const isStudentSelf = computed(() => {
  const roles = currentUserRoles.value
  if (roles.length === 0) return !!currentStudentId.value
  const isAdminish = roles.some(r =>
    ['R-SY-ADMIN', 'R-SY-LEAGUE', 'R-SY-STU-AF', 'R-SY-FA', 'R-SY-OPS',
     'R-COL-ADMIN', 'R-COL-COUN', 'R-COL-TUTOR', 'R-COL-FLOOR', 'R-COL-LEAGUE'].includes(r))
  return !isAdminish && !!currentStudentId.value
})

// 是否可标记结业（教师/管理员/团支书）
const canMarkPass = computed(() => currentUserRoles.value.some(r =>
  ['R-SY-ADMIN', 'R-SY-LEAGUE', 'R-COL-LEAGUE', 'R-COL-COUN', 'R-STU-LEAGUE'].includes(r)))

const activeTab = ref('link')

// ========== 培养联系人 ==========
const linkList = ref([])
const linkLoading = ref(false)
async function fetchLinks() {
  linkLoading.value = true
  try {
    const data = await tyCultivationLinkApi.list({ page_size: 100 })
    linkList.value = data.items || []
  } catch (e) {
    console.error('获取培养联系人列表失败', e)
  } finally {
    linkLoading.value = false
  }
}

// 培养联系人候选人（仅取政治面貌为团员/党员的学生）
const mentorCandidateOptions = ref([])
async function fetchMentorCandidates() {
  try {
    const data = await idxStudentApi.list({ page: 1, page_size: 200 })
    const items = (data && data.items) || []
    mentorCandidateOptions.value = items.filter((s) => {
      const ps = (s.political_status || '').toLowerCase()
      return (
        ps === 'member' || ps === 'league_member' || ps === '共青团员' ||
        ps === 'party_member' || ps === 'party_probationary' || ps === '中共党员' || ps === '预备党员'
      )
    })
  } catch (e) {
    console.error('加载培养联系人候选名单失败', e)
    mentorCandidateOptions.value = []
  }
}

function politicalStatusLabel(ps) {
  const map = {
    member: '团员', league_member: '团员', '共青团员': '团员',
    party_member: '党员', party_probationary: '预备党员',
    '中共党员': '党员', '预备党员': '预备党员'
  }
  return map[ps] || (ps || '群众')
}

const linkDialogVisible = ref(false)
const linkSaving = ref(false)
const linkFormRef = ref()
const emptyMentor = () => ({ mentor_student_id: null, mentor_type: 'league_member' })
const linkForm = ref({
  application_id: null,
  start_at: '',
  mentors: [emptyMentor(), emptyMentor()]
})
const linkFormRules = {
  application_id: [{ required: true, message: '请选择入团申请', trigger: 'change' }],
  start_at: [{ required: true, message: '请选择开始时间', trigger: 'change' }]
}

// 动态校验：两位联系人必填 + 不可为同一人
function mentorStudentRules(idx) {
  return [
    { required: true, message: '请选择联系人', trigger: 'change' },
    {
      validator: (rule, value, callback) => {
        if (!value) return callback()
        const other = linkForm.value.mentors[1 - idx]
        if (other && other.mentor_student_id && value === other.mentor_student_id) {
          return callback(new Error('两位联系人不能是同一人'))
        }
        callback()
      },
      trigger: 'change'
    }
  ]
}
function mentorTypeRules(idx) {
  return [{ required: true, message: '请选择联系人政治面貌', trigger: 'change' }]
}

function openLinkDialog() {
  linkForm.value = {
    application_id: null,
    start_at: new Date().toISOString().slice(0, 10),
    mentors: [emptyMentor(), emptyMentor()]
  }
  linkDialogVisible.value = true
  fetchMentorCandidates()
}
function onApplicationChange() {
  // 申请变化时清空已选联系人，避免选到申请人本人
  linkForm.value.mentors = [emptyMentor(), emptyMentor()]
}
async function handleCreateLink() {
  try { await linkFormRef.value.validate() } catch { return }
  if (
    linkForm.value.mentors[0].mentor_student_id &&
    linkForm.value.mentors[1].mentor_student_id &&
    linkForm.value.mentors[0].mentor_student_id === linkForm.value.mentors[1].mentor_student_id
  ) {
    ElMessage.warning('两位联系人不能是同一人')
    return
  }
  linkSaving.value = true
  try {
    await tyCultivationLinkApi.create(linkForm.value)
    ElMessage.success('两位培养联系人已分配')
    linkDialogVisible.value = false
    fetchLinks()
  } catch (e) {} finally { linkSaving.value = false }
}
async function handleEndLink(id) {
  try {
    await ElMessageBox.confirm('确认结束该联系人的培养关系？', '结束确认')
    await tyCultivationLinkApi.end(id)
    ElMessage.success('已结束')
    fetchLinks()
  } catch (e) { if (e !== 'cancel') {} }
}

// ========== 培养记录 ==========
const recordList = ref([])
const recordLoading = ref(false)
async function fetchRecords() {
  recordLoading.value = true
  try {
    const data = await tyCultivationRecordApi.list({ page_size: 100 })
    recordList.value = data.items || []
  } catch (e) {
    console.error('获取培养记录失败', e)
  } finally {
    recordLoading.value = false
  }
}

const recordDialogVisible = ref(false)
const recordSaving = ref(false)
const recordFormRef = ref()
const recordForm = ref({ application_id: null, record_year: new Date().getFullYear(), record_month: new Date().getMonth() + 1, record_type: 'monthly', summary: '', performance_score: 80 })
const recordFormRules = {
  application_id: [{ required: true, message: '请选择入团申请', trigger: 'change' }],
  record_year: [{ required: true, message: '请填写年份', trigger: 'blur' }],
  summary: [{ required: true, message: '请填写培养总结', trigger: 'blur' }]
}

function openRecordDialog() {
  recordForm.value = { application_id: null, record_year: new Date().getFullYear(), record_month: new Date().getMonth() + 1, record_type: 'monthly', summary: '', performance_score: 80 }
  recordDialogVisible.value = true
}
async function handleCreateRecord() {
  try { await recordFormRef.value.validate() } catch { return }
  recordSaving.value = true
  try {
    await tyCultivationRecordApi.create(recordForm.value)
    ElMessage.success('培养记录已保存')
    recordDialogVisible.value = false
    fetchRecords()
  } catch (e) {} finally { recordSaving.value = false }
}

// ========== 团课记录 ==========
const courseList = ref([])
const courseLoading = ref(false)
async function fetchCourses() {
  courseLoading.value = true
  try {
    const data = await tyCourseRecordApi.list({ page_size: 100 })
    courseList.value = data.items || []
  } catch (e) {
    console.error('获取团课记录失败', e)
  } finally {
    courseLoading.value = false
  }
}

const courseDialogVisible = ref(false)
const courseSaving = ref(false)
const courseFormRef = ref()
const courseForm = ref({ student_id: null, course_name: '', semester: '', study_at: '', certificate_no: '' })
const courseFormRules = {
  student_id: [{ required: true, message: '请选择学生', trigger: 'change' }],
  course_name: [{ required: true, message: '请输入课程名称', trigger: 'blur' }],
  study_at: [{ required: true, message: '请选择学习时间', trigger: 'change' }]
}

function openCourseDialog() {
  // 学生身份：自动绑定为本人；管理员/教师：需手动选择
  const initialStudentId = isStudentSelf.value ? currentStudentId.value : null
  courseForm.value = { student_id: initialStudentId, course_name: '', semester: '', study_at: '', certificate_no: '' }
  courseDialogVisible.value = true
}
async function handleCreateCourse() {
  try { await courseFormRef.value.validate() } catch { return }
  courseSaving.value = true
  try {
    // 学生身份不传 student_id（由后端从登录账号注入）；
    // 管理员/教师显式传 student_id
    const payload = { ...courseForm.value }
    if (isStudentSelf.value) {
      delete payload.student_id
    }
    await tyCourseRecordApi.create(payload)
    ElMessage.success('团课记录已保存')
    courseDialogVisible.value = false
    fetchCourses()
  } catch (e) {} finally { courseSaving.value = false }
}
async function handlePassCourse(id) {
  try {
    await ElMessageBox.confirm('确认标记该团课为已结业？', '结业确认')
    await tyCourseRecordApi.updatePassStatus(id)
    ElMessage.success('已标记结业')
    fetchCourses()
  } catch (e) { if (e !== 'cancel') {} }
}

// ========== 思想汇报 ==========
const thoughtList = ref([])
const thoughtLoading = ref(false)
async function fetchThoughts() {
  thoughtLoading.value = true
  try {
    const data = await tyThoughtReportApi.list({ page_size: 100 })
    thoughtList.value = data.items || []
  } catch (e) {
    console.error('获取思想汇报列表失败', e)
  } finally {
    thoughtLoading.value = false
  }
}

const thoughtDialogVisible = ref(false)
const thoughtSaving = ref(false)
const thoughtFormRef = ref()
const thoughtForm = ref({ application_id: null, title: '', content: '', quarter: '' })
const thoughtFormRules = {
  application_id: [{ required: true, message: '请选择入团申请', trigger: 'change' }],
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  content: [{ required: true, message: '请填写汇报内容', trigger: 'blur' }],
  quarter: [{ required: true, message: '请选择季度', trigger: 'change' }]
}
const thoughtContentLength = computed(() => (thoughtForm.value.content || '').length)

function openThoughtDialog() {
  thoughtForm.value = { application_id: null, title: '', content: '', quarter: '' }
  thoughtDialogVisible.value = true
}
async function handleCreateThought() {
  try { await thoughtFormRef.value.validate() } catch { return }
  if ((thoughtForm.value.content || '').length < 1000) {
    ElMessage.warning('汇报内容不少于1000字')
    return
  }
  thoughtSaving.value = true
  try {
    // 汇报人由后端根据当前登录用户自动注入，前端无需传 student_id
    await tyThoughtReportApi.create(thoughtForm.value)
    ElMessage.success('思想汇报已提交')
    thoughtDialogVisible.value = false
    fetchThoughts()
  } catch (e) {} finally { thoughtSaving.value = false }
}

// 思想汇报详情
const thoughtDetailVisible = ref(false)
const currentThought = ref(null)
function showThoughtDetail(row) {
  currentThought.value = row
  thoughtDetailVisible.value = true
}

// 培养记录详情
const recordDetailVisible = ref(false)
const currentRecord = ref(null)
function showRecordDetail(row) {
  currentRecord.value = row
  recordDetailVisible.value = true
}

// ========== 共享数据 ==========
const applications = ref([])
async function fetchApplications() {
  try {
    const data = await tyApplicationApi.list({ page_size: 200 })
    applications.value = data.items || []
  } catch (e) {
    console.error('获取入团申请列表失败', e)
  }
}

onMounted(() => {
  fetchApplications()
  fetchLinks()
})

watch(activeTab, (val) => {
  switch (val) {
    case 'link': fetchLinks(); break
    case 'record': fetchRecords(); break
    case 'course': fetchCourses(); break
    case 'thought': fetchThoughts(); break
  }
})
</script>

<style scoped>
/* .tab-header / .word-count / .thought-content 已在全局定义 */

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
