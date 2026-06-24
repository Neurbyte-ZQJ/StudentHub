<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>工时打卡管理</span>
        </div>
      </template>

      <!-- 顶部筛选区 -->
      <div class="filter-bar">
        <el-input v-model="filterPositionTitle" placeholder="岗位名称" clearable style="width: 160px" />
        <el-input v-model="filterStudentId" placeholder="学号/姓名/学生ID" clearable style="width: 200px" />
        <el-date-picker
          v-model="filterDateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          value-format="YYYY-MM-DD"
          style="width: 260px"
        />
        <el-button type="primary" @click="handleFilter">查询</el-button>
      </div>

      <!-- 标签页分区 -->
      <el-tabs v-model="activeTab">
        <!-- 打卡记录 -->
        <el-tab-pane label="打卡记录" name="attendance">
          <div class="action-bar">
            <el-button type="primary" @click="showClockInDialog">上班打卡</el-button>
            <el-button type="success" @click="showSummaryDialog">月度汇总</el-button>
          </div>

          <el-table :data="attendanceList" stripe v-loading="attendanceLoading">
            <el-table-column prop="biz_no" label="业务编号" width="150" />
            <el-table-column prop="position_title" label="岗位" min-width="120" />
            <el-table-column prop="student_name" label="学生" width="100" />
            <el-table-column prop="work_date" label="工作日期" width="110">
              <template #default="{ row }">{{ formatDate(row.work_date) }}</template>
            </el-table-column>
            <el-table-column prop="clock_in_at" label="上班时间" width="170">
              <template #default="{ row }">{{ formatDateTime(row.clock_in_at) }}</template>
            </el-table-column>
            <el-table-column prop="clock_out_at" label="下班时间" width="170">
              <template #default="{ row }">
                {{ row.clock_out_at ? formatDateTime(row.clock_out_at) : '--' }}
              </template>
            </el-table-column>
            <el-table-column prop="effective_hours" label="有效工时" width="90" />
            <el-table-column prop="late_minutes" label="迟到(分)" width="90" />
            <el-table-column prop="early_minutes" label="早退(分)" width="90" />
            <el-table-column prop="clock_method" label="打卡方式" width="90">
              <template #default="{ row }">
                {{ clockMethodMap[row.clock_method] || row.clock_method }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="140" fixed="right">
              <template #default="{ row }">
                <el-button v-if="!row.clock_out_at" link type="primary" size="small" @click="handleClockOut(row.id)">下班打卡</el-button>
                <el-popconfirm title="确认删除此打卡记录？" @confirm="handleDeleteAttendance(row.id)">
                  <template #reference>
                    <el-button link type="danger" size="small">删除</el-button>
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination-wrap">
            <el-pagination
              v-model:current-page="attendancePage"
              v-model:page-size="attendancePageSize"
              :total="attendanceTotal"
              :page-sizes="[20, 50, 100]"
              layout="total, sizes, prev, pager, next"
              @size-change="fetchAttendanceList"
              @current-change="fetchAttendanceList"
            />
          </div>
        </el-tab-pane>

        <!-- 月度考核 -->
        <el-tab-pane label="月度考核" name="assess">
          <div class="action-bar">
            <el-button type="primary" @click="showAssessDialog">创建考核</el-button>
          </div>

          <el-table :data="assessList" stripe v-loading="assessLoading">
            <el-table-column prop="position_title" label="岗位" min-width="120" />
            <el-table-column prop="student_name" label="学生" width="100" />
            <el-table-column label="年月" width="100">
              <template #default="{ row }">
                {{ row.assess_year }}-{{ String(row.assess_month).padStart(2, '0') }}
              </template>
            </el-table-column>
            <el-table-column prop="weighted_score" label="加权分" width="90" />
            <el-table-column label="考核系数" width="120">
              <template #default="{ row }">
                {{ row.coefficient }}（{{ row.coefficient_text || coefficientTextMap[row.coefficient] || '--' }}）
              </template>
            </el-table-column>
            <el-table-column prop="status_text" label="状态" width="100">
              <template #default="{ row }">
                <el-tag size="small">{{ row.status_text || '--' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button
                  v-if="row.status === 'S1'"
                  link
                  type="success"
                  size="small"
                  @click="handleConfirmAssess(row.id)"
                >确认</el-button>
                <span v-else class="qg-text-muted">--</span>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination-wrap">
            <el-pagination
              v-model:current-page="assessPage"
              v-model:page-size="assessPageSize"
              :total="assessTotal"
              :page-sizes="[20, 50, 100]"
              layout="total, sizes, prev, pager, next"
              @size-change="fetchAssessList"
              @current-change="fetchAssessList"
            />
          </div>
        </el-tab-pane>

        <!-- 薪酬管理 -->
        <el-tab-pane label="薪酬管理" name="payroll">
          <div class="action-bar">
            <el-button type="primary" @click="showPayrollDialog">计算薪酬</el-button>
          </div>

          <el-table :data="payrollList" stripe v-loading="payrollLoading">
            <el-table-column prop="student_name" label="学生" width="100" />
            <el-table-column prop="position_title" label="岗位" min-width="120" />
            <el-table-column label="年月" width="100">
              <template #default="{ row }">
                {{ row.pay_year }}-{{ String(row.pay_month).padStart(2, '0') }}
              </template>
            </el-table-column>
            <el-table-column prop="total_hours" label="总工时" width="90" />
            <el-table-column label="应发(元)" width="100">
              <template #default="{ row }">
                {{ (row.gross_cents / 100).toFixed(2) }}
              </template>
            </el-table-column>
            <el-table-column label="扣税(元)" width="100">
              <template #default="{ row }">
                {{ (row.tax_cents / 100).toFixed(2) }}
              </template>
            </el-table-column>
            <el-table-column label="实发(元)" width="100">
              <template #default="{ row }">
                {{ (row.net_cents / 100).toFixed(2) }}
              </template>
            </el-table-column>
            <el-table-column prop="coefficient" label="系数" width="80" />
            <el-table-column prop="status_text" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="payrollStatusType[row.status]" size="small">
                  {{ row.status_text || payrollStatusMap[row.status] || '--' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="140" fixed="right">
              <template #default="{ row }">
                <el-button v-if="row.status === 'draft'" link type="primary" size="small" @click="handleReview(row.id)">复核</el-button>
                <el-button v-if="row.status === 'reviewed'" link type="success" size="small" @click="handlePay(row.id)">发放</el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination-wrap">
            <el-pagination
              v-model:current-page="payrollPage"
              v-model:page-size="payrollPageSize"
              :total="payrollTotal"
              :page-sizes="[20, 50, 100]"
              layout="total, sizes, prev, pager, next"
              @size-change="fetchPayrollList"
              @current-change="fetchPayrollList"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 上班打卡对话框 -->
    <el-dialog v-model="clockInVisible" title="上班打卡" width="520px" destroy-on-close>
      <el-form :model="clockInForm" label-width="100px">
        <el-form-item label="岗位申请">
          <el-select
            v-model="clockInForm.apply_id"
            filterable
            remote
            :remote-method="searchApplies"
            :loading="applyLoading"
            placeholder="请选择岗位(搜索岗位/学生/学号)"
            style="width: 100%"
            @focus="searchApplies('')"
          >
            <el-option
              v-for="item in applyOptions"
              :key="item.id"
              :label="`${item.position_title} - ${item.student_name} (#${item.id})`"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="工作日期">
          <el-date-picker v-model="clockInForm.work_date" type="date" value-format="YYYY-MM-DD" placeholder="选择工作日期" style="width: 100%" />
        </el-form-item>
        <el-form-item label="打卡方式">
          <el-select v-model="clockInForm.clock_method" placeholder="请选择打卡方式" style="width: 100%">
            <el-option label="刷卡" value="card" />
            <el-option label="人脸识别" value="gps_face" />
            <el-option label="手动" value="manual" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="clockInVisible = false">取消</el-button>
        <el-button type="primary" @click="handleClockIn">确认打卡</el-button>
      </template>
    </el-dialog>

    <!-- 月度汇总对话框 -->
    <el-dialog v-model="summaryVisible" title="月度汇总" width="500px" destroy-on-close>
      <el-form :model="summaryForm" label-width="80px">
        <el-form-item label="学生ID">
          <el-input v-model="summaryForm.student_id" placeholder="请输入学生ID" />
        </el-form-item>
        <el-form-item label="年份">
          <el-input v-model="summaryForm.year" placeholder="例如 2026" />
        </el-form-item>
        <el-form-item label="月份">
          <el-select v-model="summaryForm.month" placeholder="请选择月份" style="width: 100%">
            <el-option v-for="m in 12" :key="m" :label="m + '月'" :value="m" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="summaryVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSummary">查询汇总</el-button>
      </template>
      <div v-if="summaryResult" class="summary-result">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="学生">{{ summaryResult.student_name }}</el-descriptions-item>
          <el-descriptions-item label="出勤天数">{{ summaryResult.attendance_days }}</el-descriptions-item>
          <el-descriptions-item label="总工时">{{ summaryResult.total_hours }}</el-descriptions-item>
          <el-descriptions-item label="迟到次数">{{ summaryResult.late_count }}</el-descriptions-item>
          <el-descriptions-item label="早退次数">{{ summaryResult.early_count }}</el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>

    <!-- 创建考核对话框 -->
    <el-dialog v-model="assessVisible" title="创建月度考核" width="560px" destroy-on-close>
      <el-form :model="assessForm" label-width="100px">
        <el-form-item label="岗位申请">
          <el-select
            v-model="assessForm.apply_id"
            filterable
            remote
            :remote-method="searchApplies"
            :loading="applyLoading"
            placeholder="请选择岗位(搜索岗位/学生/学号)"
            style="width: 100%"
            @focus="searchApplies('')"
          >
            <el-option
              v-for="item in applyOptions"
              :key="item.id"
              :label="`${item.position_title} - ${item.student_name} (#${item.id})`"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="年月">
          <el-date-picker v-model="assessForm.assess_ym" type="month" value-format="YYYY-MM" placeholder="选择年月" style="width: 100%" />
        </el-form-item>
        <el-form-item label="出勤分">
          <el-input v-model.number="assessForm.score_attendance" type="number" placeholder="0-100(自动按工时算出,可手动调整)" />
          <div v-if="attendancePreview" class="qg-attendance-hint">
            实出勤 {{ attendancePreview.actual_hours }}h / 标准工时 {{ attendancePreview.should_hours }}h
            <span class="qg-attendance-hint__formula">（{{ attendancePreview.formula }}）</span>
          </div>
        </el-form-item>
        <el-form-item label="工作完成分">
          <el-input v-model.number="assessForm.score_work_complete" type="number" placeholder="0-100" />
        </el-form-item>
        <el-form-item label="综合分">
          <el-input v-model.number="assessForm.score_comprehensive" type="number" placeholder="0-100(人工评分:满意度/仪容仪表/协作度)" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="assessVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreateAssess">确认创建</el-button>
      </template>
    </el-dialog>

    <!-- 计算薪酬对话框 -->
    <el-dialog v-model="payrollVisible" title="计算薪酬" width="520px" destroy-on-close>
      <el-form :model="payrollForm" label-width="100px">
        <el-form-item label="岗位申请">
          <el-select
            v-model="payrollForm.apply_id"
            filterable
            remote
            :remote-method="searchApplies"
            :loading="applyLoading"
            placeholder="请选择岗位(搜索岗位/学生/学号)"
            style="width: 100%"
            @focus="searchApplies('')"
          >
            <el-option
              v-for="item in applyOptions"
              :key="item.id"
              :label="`${item.position_title} - ${item.student_name} (#${item.id})`"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="年月">
          <el-date-picker v-model="payrollForm.pay_ym" type="month" value-format="YYYY-MM" placeholder="选择年月" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="payrollVisible = false">取消</el-button>
        <el-button type="primary" @click="handleComputePayroll">确认计算</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { qgAttendanceApi, qgAssessApi, qgPayrollApi, qgApplyApi } from '@/api/qg'
import { formatDateTime, formatDate } from '@/utils/datetime'

// ========== 映射常量 ==========
const clockMethodMap = { card: '刷卡', gps_face: '人脸识别', manual: '手动' }
const payrollStatusMap = { draft: '草稿', reviewed: '已复核', paid: '已发放', failed: '发放失败' }
const payrollStatusType = { draft: 'info', reviewed: '', paid: 'success', failed: 'danger' }
const coefficientTextMap = { 1.0: '全额', 0.8: '八折', 0.5: '五折' }

// ========== 顶部筛选 ==========
const filterPositionTitle = ref('')
const filterStudentId = ref('')
const filterDateRange = ref(null)
const activeTab = ref('attendance')

function handleFilter() {
  if (activeTab.value === 'attendance') {
    attendancePage.value = 1
    fetchAttendanceList()
  } else if (activeTab.value === 'assess') {
    assessPage.value = 1
    fetchAssessList()
  } else if (activeTab.value === 'payroll') {
    payrollPage.value = 1
    fetchPayrollList()
  }
}

// ========== 打卡记录 ==========
const attendanceList = ref([])
const attendanceLoading = ref(false)
const attendancePage = ref(1)
const attendancePageSize = ref(20)
const attendanceTotal = ref(0)

async function fetchAttendanceList() {
  attendanceLoading.value = true
  try {
    const params = {
      page: attendancePage.value,
      page_size: attendancePageSize.value
    }
    if (filterPositionTitle.value) params.position_title = filterPositionTitle.value
    if (filterStudentId.value) params.student_keyword = filterStudentId.value
    if (filterDateRange.value && filterDateRange.value.length === 2) {
      params.date_from = filterDateRange.value[0]
      params.date_to = filterDateRange.value[1]
    }
    const data = await qgAttendanceApi.list(params)
    attendanceList.value = data.items || []
    attendanceTotal.value = data.total || 0
  } catch (e) {
    console.error('获取打卡记录失败', e)
  } finally {
    attendanceLoading.value = false
  }
}

// 上班打卡
const clockInVisible = ref(false)
const clockInForm = ref({ apply_id: '', student_id: '', student_name: '', work_date: '', clock_method: 'card' })

// 岗位申请下拉选项（远程搜索用）
const applyOptions = ref([])
const applyLoading = ref(false)

async function searchApplies(keyword = '') {
  applyLoading.value = true
  try {
    const data = await qgApplyApi.list({ keyword, page_size: 50 })
    applyOptions.value = data.items || []
  } catch (e) {
    console.error('加载岗位申请失败', e)
    applyOptions.value = []
  } finally {
    applyLoading.value = false
  }
}

function showClockInDialog() {
  clockInForm.value = { apply_id: '', student_id: '', student_name: '', work_date: '', clock_method: 'card' }
  searchApplies('')
  clockInVisible.value = true
}

async function handleClockIn() {
  if (!clockInForm.value.apply_id) {
    ElMessage.warning('请选择岗位申请')
    return
  }
  if (!clockInForm.value.work_date) {
    ElMessage.warning('请选择工作日期')
    return
  }
  // 从已选岗位补出 student_id / student_name
  const picked = applyOptions.value.find(a => a.id === clockInForm.value.apply_id)
  if (!picked) {
    ElMessage.warning('岗位申请数据已失效，请重新选择')
    return
  }
  try {
    const { apply_id, work_date, clock_method } = clockInForm.value
    await qgAttendanceApi.clockIn({ apply_id, work_date, clock_method }, picked.student_id)
    ElMessage.success('上班打卡成功')
    clockInVisible.value = false
    fetchAttendanceList()
  } catch (e) {
    console.error('上班打卡失败', e)
  }
}

// 下班打卡
async function handleClockOut(id) {
  try {
    await ElMessageBox.confirm('确认进行下班打卡？', '下班打卡')
    await qgAttendanceApi.clockOut(id)
    ElMessage.success('下班打卡成功')
    fetchAttendanceList()
  } catch (e) {
    if (e !== 'cancel') {
      console.error('下班打卡失败', e)
    }
  }
}

// 删除打卡记录
async function handleDeleteAttendance(id) {
  try {
    await qgAttendanceApi.delete(id)
    ElMessage.success('删除成功')
    fetchAttendanceList()
  } catch (e) {
    console.error('删除打卡记录失败', e)
  }
}

// 月度汇总
const summaryVisible = ref(false)
const summaryForm = ref({ student_id: '', year: '', month: '' })
const summaryResult = ref(null)

function showSummaryDialog() {
  summaryForm.value = { student_id: '', year: '', month: '' }
  summaryResult.value = null
  summaryVisible.value = true
}

async function handleSummary() {
  if (!summaryForm.value.student_id) {
    ElMessage.warning('请输入学生ID')
    return
  }
  if (!summaryForm.value.year) {
    ElMessage.warning('请输入年份')
    return
  }
  if (!summaryForm.value.month) {
    ElMessage.warning('请选择月份')
    return
  }
  try {
    const data = await qgAttendanceApi.monthlySummary({
      student_id: Number(summaryForm.value.student_id),
      year: Number(summaryForm.value.year),
      month: Number(summaryForm.value.month)
    })
    summaryResult.value = data
  } catch (e) {
    console.error('查询月度汇总失败', e)
  }
}

// ========== 月度考核 ==========
const assessList = ref([])
const assessLoading = ref(false)
const assessPage = ref(1)
const assessPageSize = ref(20)
const assessTotal = ref(0)

async function fetchAssessList() {
  assessLoading.value = true
  try {
    const params = {
      page: assessPage.value,
      page_size: assessPageSize.value
    }
    if (filterPositionTitle.value) params.position_title = filterPositionTitle.value
    if (filterStudentId.value) params.student_id = filterStudentId.value
    const data = await qgAssessApi.list(params)
    assessList.value = data.items || []
    assessTotal.value = data.total || 0
  } catch (e) {
    console.error('获取考核列表失败', e)
  } finally {
    assessLoading.value = false
  }
}

// 创建考核
const assessVisible = ref(false)
const assessForm = ref({
  apply_id: '',
  assess_ym: '',
  score_attendance: 0,
  score_work_complete: 0,
  score_comprehensive: 0
})

function showAssessDialog() {
  assessForm.value = {
    apply_id: '',
    assess_ym: '',
    // 三项分数均默认空,等待用户填写或触发"出勤分自动计算"。
    // 留空提交时后端 binding:"required" 对 int 0 不拦截,最终会被存为 0。
    score_attendance: null,
    score_work_complete: null,
    score_comprehensive: null
  }
  attendancePreview.value = null
  searchApplies('')
  assessVisible.value = true
}

// 出勤分预览（自动算 score_attendance）
const attendancePreview = ref(null)
let previewTimer = null

async function refreshAttendancePreview() {
  const applyId = assessForm.value.apply_id
  const ym = assessForm.value.assess_ym
  if (!applyId || !ym) {
    attendancePreview.value = null
    return
  }
  const [year, month] = ym.split('-').map(Number)
  try {
    const data = await qgAssessApi.previewAttendance(applyId, year, month)
    attendancePreview.value = data
    // 自动回填出勤分（允许用户后续手动覆盖）
    assessForm.value.score_attendance = data.score_attendance
  } catch (e) {
    attendancePreview.value = null
    console.error('出勤分预览失败', e)
  }
}

// 监听 apply_id / assess_ym 变化,触发出勤分预览(防抖 250ms)
watch(
  () => [assessForm.value.apply_id, assessForm.value.assess_ym],
  () => {
    if (previewTimer) clearTimeout(previewTimer)
    previewTimer = setTimeout(refreshAttendancePreview, 250)
  }
)

async function handleCreateAssess() {
  if (!assessForm.value.apply_id) {
    ElMessage.warning('请选择岗位申请')
    return
  }
  if (!assessForm.value.assess_ym) {
    ElMessage.warning('请选择年月')
    return
  }
  assessLoading.value = true
  try {
    const [year, month] = assessForm.value.assess_ym.split('-').map(Number)
    await qgAssessApi.create({
      apply_id: Number(assessForm.value.apply_id),
      assess_year: year,
      assess_month: month,
      score_attendance: assessForm.value.score_attendance,
      score_work_complete: assessForm.value.score_work_complete,
      score_comprehensive: assessForm.value.score_comprehensive
    })
    ElMessage.success('考核创建成功')
    assessVisible.value = false
    fetchAssessList()
  } catch (e) {
    // 业务错误已由 http.js 拦截器 ElMessage.error 提示,这里只记录日志。
    console.error('创建考核失败', e)
  } finally {
    assessLoading.value = false
  }
}

// 确认月度考核（S1 → S3），状态机推进，由学生处 / 财务管理员触发
async function handleConfirmAssess(id) {
  try {
    await ElMessageBox.confirm('确认将此月度考核标记为"已确认"？', '确认月度考核')
    await qgAssessApi.confirm(id)
    ElMessage.success('已确认')
    fetchAssessList()
  } catch (e) {
    if (e !== 'cancel') {
      console.error('确认月度考核失败', e)
    }
  }
}

// ========== 薪酬管理 ==========
const payrollList = ref([])
const payrollLoading = ref(false)
const payrollPage = ref(1)
const payrollPageSize = ref(20)
const payrollTotal = ref(0)

async function fetchPayrollList() {
  payrollLoading.value = true
  try {
    const params = {
      page: payrollPage.value,
      page_size: payrollPageSize.value
    }
    if (filterPositionTitle.value) params.position_title = filterPositionTitle.value
    if (filterStudentId.value) params.student_id = filterStudentId.value
    const data = await qgPayrollApi.list(params)
    payrollList.value = data.items || []
    payrollTotal.value = data.total || 0
  } catch (e) {
    console.error('获取薪酬列表失败', e)
  } finally {
    payrollLoading.value = false
  }
}

// 计算薪酬
const payrollVisible = ref(false)
const payrollForm = ref({ apply_id: '', pay_ym: '' })

function showPayrollDialog() {
  payrollForm.value = { apply_id: '', pay_ym: '' }
  searchApplies('')
  payrollVisible.value = true
}

async function handleComputePayroll() {
  if (!payrollForm.value.apply_id) {
    ElMessage.warning('请选择岗位申请')
    return
  }
  if (!payrollForm.value.pay_ym) {
    ElMessage.warning('请选择年月')
    return
  }
  try {
    const [year, month] = payrollForm.value.pay_ym.split('-').map(Number)
    await qgPayrollApi.compute({
      apply_id: Number(payrollForm.value.apply_id),
      year,
      month
    })
    ElMessage.success('薪酬计算成功')
    payrollVisible.value = false
    fetchPayrollList()
  } catch (e) {
    console.error('计算薪酬失败', e)
  }
}

// 复核
async function handleReview(id) {
  try {
    await ElMessageBox.confirm('确认复核此薪酬记录？', '薪酬复核')
    await qgPayrollApi.review(id)
    ElMessage.success('复核成功')
    fetchPayrollList()
  } catch (e) {
    if (e !== 'cancel') {
      console.error('复核失败', e)
    }
  }
}

// 发放
async function handlePay(id) {
  try {
    await ElMessageBox.confirm('确认发放此薪酬？', '薪酬发放')
    await qgPayrollApi.pay(id)
    ElMessage.success('发放成功')
    fetchPayrollList()
  } catch (e) {
    if (e !== 'cancel') {
      console.error('发放失败', e)
    }
  }
}

// ========== 初始化 ==========
onMounted(() => {
  // 三个 tab 数据都预加载：避免依赖 tab 切换触发（如 Vite HMR 失败/浏览器缓存时也能立刻看到所有数据）
  fetchAttendanceList()
  fetchAssessList()
  fetchPayrollList()
})
</script>

<style scoped>
/* .card-header, .filter-bar, .action-bar, .pagination-wrap 已在 App.vue 全局定义 */
.summary-result {
  margin-top: var(--sh-space-md);
}
.qg-attendance-hint {
  font-size: 12px;
  color: var(--el-color-info-light-3, #909399);
  margin-top: 4px;
  line-height: 1.5;
}
.qg-attendance-hint__formula {
  color: var(--el-text-color-secondary, #606266);
  font-size: 11px;
}
.qg-text-muted {
  color: var(--el-text-color-placeholder, #c0c4cc);
  font-size: 12px;
}
</style>
