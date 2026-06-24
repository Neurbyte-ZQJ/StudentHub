<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>{{ isEdit ? '编辑社团' : '新建社团' }}</span>
        </div>
      </template>

      <el-form :model="form" :rules="rules" ref="formRef" label-width="120px" style="max-width: 700px">
        <el-form-item label="社团名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入社团名称" />
        </el-form-item>

        <el-form-item label="所属院系" prop="college_id">
          <el-select v-model="form.college_id" placeholder="请选择院系" style="width: 100%">
            <el-option v-for="c in colleges" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="指导教师" prop="tutor_user_id">
          <el-select v-model="form.tutor_user_id" placeholder="请选择指导教师" clearable style="width: 100%">
            <el-option v-for="u in users" :key="u.id" :label="u.display_name" :value="u.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="社长" prop="president_student_id">
          <el-select v-model="form.president_student_id" placeholder="请选择社长" clearable filterable style="width: 100%">
            <el-option
              v-for="s in students"
              :key="s.id"
              :label="`${s.student_no} ${s.name}`"
              :value="s.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="业务范围" prop="business_scope">
          <el-input v-model="form.business_scope" type="textarea" :rows="3" placeholder="请描述社团业务范围" />
        </el-form-item>

        <el-form-item label="发起人学号" prop="founderInput" v-if="!isEdit">
          <div style="width: 100%">
            <div style="display: flex; gap: 8px; margin-bottom: 8px">
              <el-input v-model="founderInput" placeholder="输入学号后回车添加" @keyup.enter="addFounder" />
              <el-button @click="addFounder" :disabled="!founderInput.trim()">添加</el-button>
            </div>
            <el-tag v-for="(fid, idx) in form.founders" :key="idx" closable :disable-transitions="false" style="margin-right: 6px; margin-bottom: 4px" @close="removeFounder(idx)">
              {{ fid }}
            </el-tag>
            <div v-if="form.founders.length > 0" style="margin-top: 4px; font-size: 12px; color: #909399">
              已添加 {{ form.founders.length }} 人（须 5-20 人）
            </div>
          </div>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSave" :loading="submitting">保存</el-button>
          <el-button @click="router.back()">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { stAssociationApi } from '@/api/st'
import { collegeApi } from '@/api/sys-org'

const router = useRouter()
const route = useRoute()
const isEdit = !!route.params.id

const formRef = ref()
const submitting = ref(false)
const colleges = ref([])
const users = ref([])
const students = ref([])
const founderInput = ref('')

const form = reactive({
  name: '',
  college_id: null,
  tutor_user_id: null,
  president_student_id: null,
  business_scope: '',
  founders: []
})

const rules = {
  name: [{ required: true, message: '请输入社团名称', trigger: 'blur' }],
  college_id: [{ required: true, message: '请选择所属院系', trigger: 'change' }],
  business_scope: [{ required: true, message: '请输入业务范围', trigger: 'blur' }]
}

function addFounder() {
  const val = founderInput.value.trim()
  if (!val) return
  if (form.founders.includes(val)) {
    ElMessage.warning('该学号已添加')
    return
  }
  form.founders.push(val)
  founderInput.value = ''
}

function removeFounder(idx) {
  form.founders.splice(idx, 1)
}

async function fetchColleges() {
  try {
    const data = await collegeApi.list()
    colleges.value = data.items || data || []
  } catch (e) {
    console.error('获取院系列表失败', e)
  }
}

async function fetchUsers() {
  try {
    const data = await stAssociationApi.listUsers()
    users.value = data.items || []
  } catch (e) {
    console.error('获取用户列表失败', e)
  }
}

async function fetchStudents() {
  try {
    const data = await stAssociationApi.listStudents()
    students.value = data.items || []
  } catch (e) {
    console.error('获取学生列表失败', e)
  }
}

async function fetchDetail() {
  if (!isEdit) return
  try {
    const data = await stAssociationApi.get(route.params.id)
    form.name = data.name
    form.college_id = data.college_id
    form.tutor_user_id = data.tutor_user_id || null
    form.president_student_id = data.president_student_id || null
    form.business_scope = data.business_scope
  } catch (e) {
    ElMessage.error('获取社团信息失败')
  }
}

async function handleSave() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  if (!isEdit && (form.founders.length < 5 || form.founders.length > 20)) {
    ElMessage.warning('发起人须 5-20 名')
    return
  }

  submitting.value = true
  try {
    if (isEdit) {
      await stAssociationApi.update(route.params.id, {
        name: form.name,
        college_id: form.college_id,
        tutor_user_id: form.tutor_user_id,
        president_student_id: form.president_student_id,
        business_scope: form.business_scope
      })
      ElMessage.success('更新成功')
    } else {
      await stAssociationApi.create({
        name: form.name,
        college_id: form.college_id,
        tutor_user_id: form.tutor_user_id,
        president_student_id: form.president_student_id,
        business_scope: form.business_scope,
        founders: form.founders
      })
      ElMessage.success('创建成功')
    }
    router.push('/st/association')
  } catch (e) {
    // 错误已由 http 拦截器处理
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  fetchColleges()
  fetchUsers()
  fetchStudents()
  if (isEdit) fetchDetail()
})
</script>

<style scoped>
/* .card-header 已在 App.vue 全局定义 */
</style>
