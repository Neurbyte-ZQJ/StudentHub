<template>
  <div class="student-list">
    <div class="left-panel">
      <el-card shadow="never">
        <template #header>
          <span>组织架构</span>
        </template>
        <el-tree
          ref="orgTreeRef"
          :data="orgTree"
          :props="{ label: 'label', children: 'children' }"
          node-key="unique_key"
          highlight-current
          default-expand-all
          @node-click="onNodeClick"
        />
      </el-card>
    </div>
    <div class="right-panel">
      <el-card shadow="never">
        <template #header>
          <div class="card-header">
            <span>学生列表</span>
            <div>
              <el-input
                v-model="keyword"
                placeholder="搜索学号/姓名"
                clearable
                style="width: 200px; margin-right: 12px"
                @clear="fetchStudents"
                @keyup.enter="fetchStudents"
              />
              <el-button type="primary" @click="showForm()">新增学生</el-button>
            </div>
          </div>
        </template>
        <el-table :data="students" stripe v-loading="loading">
          <el-table-column prop="student_no" label="学号" width="120" />
          <el-table-column prop="name" label="姓名" width="100" />
          <el-table-column prop="gender" label="性别" width="60">
            <template #default="{ row }">
              {{ genderMap[row.gender] || row.gender }}
            </template>
          </el-table-column>
          <el-table-column prop="college_name" label="院系" min-width="140" />
          <el-table-column prop="major_name" label="专业" min-width="120" />
          <el-table-column prop="class_name" label="班级" min-width="100" />
          <el-table-column prop="id_card_masked" label="身份证" width="180" />
          <el-table-column prop="phone_masked" label="手机号" width="130" />
          <el-table-column prop="status" label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="statusType[row.status]" size="small">
                {{ statusMap[row.status] || row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="showForm(row)">编辑</el-button>
              <el-popconfirm title="确认删除？" @confirm="handleDelete(row.id)">
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
            @size-change="fetchStudents"
            @current-change="fetchStudents"
          />
        </div>
      </el-card>
    </div>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="formVisible" :title="formTitle" width="600px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="90px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="学号" prop="student_no">
              <el-input v-model="form.student_no" :disabled="!!form.id" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="姓名" prop="name">
              <el-input v-model="form.name" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="性别" prop="gender">
              <el-select v-model="form.gender" placeholder="请选择" style="width: 100%">
                <el-option label="男" value="M" />
                <el-option label="女" value="F" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="身份证号" prop="id_card">
              <el-input v-model="form.id_card" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="院系" prop="college_id">
              <el-select v-model="form.college_id" placeholder="请选择院系" style="width: 100%" @change="onCollegeChange">
                <el-option v-for="c in colleges" :key="c.id" :label="c.name" :value="c.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="专业" prop="major_id">
              <el-select v-model="form.major_id" placeholder="请选择专业" style="width: 100%" @change="onMajorChange">
                <el-option v-for="m in filteredMajors" :key="m.id" :label="m.name" :value="m.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="班级" prop="class_id">
              <el-select v-model="form.class_id" placeholder="请选择班级" style="width: 100%">
                <el-option v-for="cl in filteredClasses" :key="cl.id" :label="cl.name" :value="cl.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="年级">
              <el-input-number v-model="form.grade" :min="2000" :max="2099" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="手机号">
              <el-input v-model="form.phone" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="邮箱">
              <el-input v-model="form.email" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="政治面貌">
              <el-select v-model="form.political_status" placeholder="请选择" style="width: 100%">
                <el-option label="群众" value="masses" />
                <el-option label="入团积极分子" value="activist" />
                <el-option label="预备团员" value="probationary" />
                <el-option label="共青团员" value="member" />
                <el-option label="预备党员" value="party_probationary" />
                <el-option label="中共党员" value="party_member" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="入学日期">
              <el-date-picker v-model="form.enrollment_at" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row v-if="form.political_status === 'member'" :gutter="16">
          <el-col :span="12">
            <el-form-item label="入团时间">
              <el-date-picker v-model="form.join_at" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="团员证号">
              <el-input v-model="form.member_card_no" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { studentApi } from '@/api/idx'
import { collegeApi, majorApi, classApi } from '@/api/sys-org'

// 性别/状态映射
const genderMap = { M: '男', F: '女', U: '未知' }
const statusMap = { enrolled: '在读', suspended: '休学', graduated: '毕业', withdrawn: '退学' }
const statusType = { enrolled: 'success', suspended: 'warning', graduated: 'info', withdrawn: 'danger' }

// 组织树
const orgTree = ref([])
const orgTreeRef = ref()
const selectedNode = ref(null)

// 学生列表
const students = ref([])
const loading = ref(false)
const keyword = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

// 院系/专业/班级下拉
const colleges = ref([])
const majors = ref([])
const classes = ref([])

// 表单
const formVisible = ref(false)
const formTitle = ref('新增学生')
const submitting = ref(false)
const formRef = ref()
const form = ref({
  id: null,
  student_no: '',
  name: '',
  gender: 'M',
  id_card: '',
  college_id: null,
  major_id: null,
  class_id: null,
  grade: new Date().getFullYear(),
  phone: '',
  email: '',
  political_status: 'masses',
  join_at: '',
  member_card_no: '',
  enrollment_at: ''
})

const formRules = {
  student_no: [{ required: true, message: '请输入学号', trigger: 'blur' }],
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  gender: [{ required: true, message: '请选择性别', trigger: 'change' }]
}

const filteredMajors = computed(() => {
  if (!form.value.college_id) return majors.value
  return majors.value.filter(m => m.college_id === form.value.college_id)
})

const filteredClasses = computed(() => {
  if (!form.value.major_id) return classes.value
  return classes.value.filter(c => c.major_id === form.value.major_id)
})

// 获取组织树
async function fetchOrgTree() {
  try {
    const data = await studentApi.getOrgTree()
    orgTree.value = data.tree || []
  } catch (e) {
    console.error('获取组织树失败', e)
  }
}

// 获取学生列表
async function fetchStudents() {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: keyword.value
    }
    if (selectedNode.value) {
      if (selectedNode.value.type === 'college') params.college_id = selectedNode.value.id
      if (selectedNode.value.type === 'major') params.college_id = selectedNode.value.parent_id
      if (selectedNode.value.type === 'class') {
        params.class_id = selectedNode.value.id
      }
    }
    const data = await studentApi.list(params)
    students.value = data.items || []
    total.value = data.total || 0
  } catch (e) {
    console.error('获取学生列表失败', e)
  } finally {
    loading.value = false
  }
}

// 获取下拉数据
async function fetchDropdowns() {
  try {
    const [colData, majData, clsData] = await Promise.all([
      collegeApi.list(),
      majorApi.list(),
      classApi.list()
    ])
    colleges.value = colData.items || []
    majors.value = majData.items || []
    classes.value = clsData.items || []
  } catch (e) {
    console.error('获取下拉数据失败', e)
  }
}

// 组织树节点点击
function onNodeClick(nodeData) {
  selectedNode.value = nodeData
  page.value = 1
  fetchStudents()
}

// 院系变化
function onCollegeChange() {
  form.value.major_id = null
  form.value.class_id = null
}

// 专业变化
function onMajorChange() {
  form.value.class_id = null
}

// 显示表单
function showForm(row) {
  if (row) {
    formTitle.value = '编辑学生'
    form.value = {
      id: row.id,
      student_no: row.student_no,
      name: row.name,
      gender: row.gender,
      id_card: row.id_card_masked || '',
      college_id: row.college_id,
      major_id: row.major_id,
      class_id: row.class_id,
      grade: row.grade,
      phone: row.phone_masked || '',
      email: row.email,
      political_status: row.political_status,
      join_at: row.join_at || '',
      member_card_no: row.member_card_no || '',
      enrollment_at: row.enrollment_at
    }
  } else {
    formTitle.value = '新增学生'
    form.value = {
      id: null,
      student_no: '',
      name: '',
      gender: 'M',
      id_card: '',
      college_id: null,
      major_id: null,
      class_id: null,
      grade: new Date().getFullYear(),
      phone: '',
      email: '',
      political_status: 'masses',
      enrollment_at: ''
    }
  }
  formVisible.value = true
}

// 提交表单
async function handleSubmit() {
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  submitting.value = true
  try {
    if (form.value.id) {
      await studentApi.update(form.value.id, form.value)
      ElMessage.success('更新成功')
    } else {
      await studentApi.create(form.value)
      ElMessage.success('创建成功')
    }
    formVisible.value = false
    fetchStudents()
  } catch (e) {
    // 错误已由 http 拦截器处理
  } finally {
    submitting.value = false
  }
}

// 删除
async function handleDelete(id) {
  try {
    await studentApi.delete(id)
    ElMessage.success('删除成功')
    fetchStudents()
  } catch (e) {
    // 错误已由 http 拦截器处理
  }
}

onMounted(() => {
  fetchOrgTree()
  fetchStudents()
  fetchDropdowns()
})
</script>

<style scoped>
.student-list {
  display: flex;
  gap: var(--sh-space-md);
  height: calc(100vh - 160px);
  padding: var(--sh-space-lg);
}
.left-panel {
  width: 260px;
  flex-shrink: 0;
}
.left-panel :deep(.el-card__body) {
  padding: var(--sh-space-sm);
  overflow-y: auto;
  max-height: calc(100vh - 220px);
}
.right-panel {
  flex: 1;
  min-width: 0;
}
/* .card-header, .pagination-wrap 已在 App.vue 全局定义 */
</style>
