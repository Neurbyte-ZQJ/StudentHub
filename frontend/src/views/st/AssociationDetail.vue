<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>社团详情</span>
          <div>
            <el-button v-if="assoc && assoc.status !== 'cancelled'" type="primary" size="small" @click="goEdit">编辑</el-button>
            <el-button @click="goBack">返回</el-button>
          </div>
        </div>
      </template>

      <el-descriptions :column="2" border v-if="assoc">
        <el-descriptions-item label="编号" :span="2">{{ assoc.biz_no }}</el-descriptions-item>
        <el-descriptions-item label="社团名称">{{ assoc.name }}</el-descriptions-item>
        <el-descriptions-item label="所属院系">{{ assoc.college_name }}</el-descriptions-item>
        <el-descriptions-item label="指导教师">{{ assoc.tutor_name || '未指定' }}</el-descriptions-item>
        <el-descriptions-item label="社长">{{ assoc.president_name || '未设置' }}</el-descriptions-item>
        <el-descriptions-item label="状态" :span="2">
          <el-tag :type="statusType[assoc.status]" size="small">{{ assoc.status_text }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="业务范围" :span="2">{{ assoc.business_scope }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDateTime(assoc.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatDateTime(assoc.updated_at) }}</el-descriptions-item>
      </el-descriptions>

      <el-tabs v-model="activeTab" style="margin-top: 20px">
        <el-tab-pane label="发起人" name="founders">
          <el-table :data="founders" stripe v-loading="foundersLoading">
            <el-table-column prop="student_no" label="学号" width="150" />
            <el-table-column prop="student_name" label="姓名" width="120" />
          </el-table>
        </el-tab-pane>
        <el-tab-pane label="成员" name="members">
          <el-table :data="members" stripe v-loading="membersLoading">
            <el-table-column prop="student_no" label="学号" width="150" />
            <el-table-column prop="student_name" label="姓名" width="120" />
            <el-table-column prop="role_text" label="角色" width="100">
              <template #default="{ row }">
                <el-tag size="small">{{ row.role_text }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="joined_at" label="加入时间" width="170">
              <template #default="{ row }">{{ formatDateTime(row.joined_at) }}</template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
        <el-tab-pane label="活动" name="activities">
          <el-table :data="activities" stripe v-loading="activitiesLoading">
            <el-table-column prop="biz_no" label="编号" width="150" />
            <el-table-column prop="title" label="活动名称" min-width="160" />
            <el-table-column prop="level" label="等级" width="80">
              <template #default="{ row }">
                <el-tag :type="levelType[row.level]" size="small">{{ row.level }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="status_text" label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="actStatusType[row.status]" size="small">{{ row.status_text }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="started_at" label="开始时间" width="170">
              <template #default="{ row }">{{ formatDateTime(row.started_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="80">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="goActivityDetail(row.id)">查看</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { stAssociationApi, stActivityApi } from '@/api/st'
import { formatDateTime } from '@/utils/datetime'

const router = useRouter()
const route = useRoute()

const statusType = { preparing: 'info', trial: 'warning', registered: 'success', rectifying: 'danger', cancelled: 'info' }
const levelType = { A: 'danger', B: 'warning', C: '', D: 'info' }
const actStatusType = { S0: 'info', S1: 'warning', S2: '', S3: 'success', S4: 'danger', cancelled: 'info' }

const assoc = ref(null)
const activeTab = ref('founders')
const founders = ref([])
const foundersLoading = ref(false)
const members = ref([])
const membersLoading = ref(false)
const activities = ref([])
const activitiesLoading = ref(false)

async function fetchDetail() {
  try {
    assoc.value = await stAssociationApi.get(route.params.id)
  } catch (e) {
    console.error('获取社团详情失败', e)
  }
}

async function fetchFounders() {
  foundersLoading.value = true
  try {
    const data = await stAssociationApi.listFounders(route.params.id)
    founders.value = data.items || []
  } catch (e) {
    console.error('获取发起人列表失败', e)
  } finally {
    foundersLoading.value = false
  }
}

async function fetchMembers() {
  membersLoading.value = true
  try {
    const data = await stAssociationApi.listMembers(route.params.id)
    members.value = data.items || []
  } catch (e) {
    console.error('获取成员列表失败', e)
  } finally {
    membersLoading.value = false
  }
}

async function fetchActivities() {
  activitiesLoading.value = true
  try {
    const data = await stActivityApi.list({ association_id: route.params.id, page_size: 100 })
    activities.value = data.items || []
  } catch (e) {
    console.error('获取活动列表失败', e)
  } finally {
    activitiesLoading.value = false
  }
}

function goEdit() {
  router.push(`/st/association/${route.params.id}/edit`)
}

function goBack() {
  // 有历史记录则返回，否则回到列表页（避免新开页签或刷新后无历史可退）
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/st/association')
  }
}

function goActivityDetail(id) {
  router.push(`/st/activity/${id}`)
}

watch(activeTab, (tab) => {
  if (tab === 'founders' && founders.value.length === 0) fetchFounders()
  if (tab === 'members' && members.value.length === 0) fetchMembers()
  if (tab === 'activities' && activities.value.length === 0) fetchActivities()
})

onMounted(() => {
  fetchDetail()
  fetchFounders()
})
</script>

<style scoped>
/* .card-header 已在 App.vue 全局定义 */
</style>
