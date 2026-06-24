import http from './http'

// 困难认定 API
export const qgDifficultyApi = {
  list(params) {
    return http.get('/qg/difficulty-certs', { params })
  },
  get(id) {
    return http.get(`/qg/difficulty-certs/${id}`)
  },
  create(data) {
    return http.post('/qg/difficulty-certs', data)
  },
  submit(id) {
    return http.post(`/qg/difficulty-certs/${id}/submit`)
  },
  approve(id, data) {
    return http.post(`/qg/difficulty-certs/${id}/approve`, data)
  },
  reject(id, data) {
    return http.post(`/qg/difficulty-certs/${id}/reject`, data)
  },
  delete(id) {
    return http.delete(`/qg/difficulty-certs/${id}`)
  }
}

// 岗位 API
export const qgPositionApi = {
  list(params) {
    return http.get('/qg/positions', { params })
  },
  get(id) {
    return http.get(`/qg/positions/${id}`)
  },
  create(data) {
    return http.post('/qg/positions', data)
  },
  submit(id) {
    return http.post(`/qg/positions/${id}/submit`)
  },
  approve(id, data) {
    return http.post(`/qg/positions/${id}/approve`, data)
  },
  reject(id, data) {
    return http.post(`/qg/positions/${id}/reject`, data)
  },
  delete(id) {
    return http.delete(`/qg/positions/${id}`)
  }
}

// 岗位申请 API
export const qgApplyApi = {
  apply(data, studentId) {
    return http.post('/qg/applies', data, { params: { student_id: studentId } })
  },
  list(params) {
    return http.get('/qg/applies', { params })
  },
  get(id) {
    return http.get(`/qg/applies/${id}`)
  },
  accept(id) {
    return http.post(`/qg/applies/${id}/accept`)
  },
  confirm(id) {
    return http.post(`/qg/applies/${id}/confirm`)
  },
  onboard(id) {
    return http.post(`/qg/applies/${id}/onboard`)
  }
}

// 工时打卡 API
export const qgAttendanceApi = {
  list(params) {
    return http.get('/qg/attendances', { params })
  },
  clockIn(data, studentId) {
    return http.post('/qg/attendances/clock-in', data, { params: { student_id: studentId } })
  },
  clockOut(id) {
    return http.post(`/qg/attendances/${id}/clock-out`)
  },
  monthlySummary(params) {
    return http.get('/qg/attendances/monthly-summary', { params })
  },
  delete(id) {
    return http.delete(`/qg/attendances/${id}`)
  }
}

// 考核 API
export const qgAssessApi = {
  create(data) {
    return http.post('/qg/monthly-assessments', data)
  },
  list(params) {
    return http.get('/qg/monthly-assessments', { params })
  },
  get(id) {
    return http.get(`/qg/monthly-assessments/${id}`)
  },
  // 确认月度考核（S1 → S3），由学生处 / 财务管理员触发
  confirm(id) {
    return http.post(`/qg/monthly-assessments/${id}/confirm`)
  },
  // 出勤分预览（不写库），用于"创建月度考核"对话框自动回填出勤分
  previewAttendance(applyId, year, month) {
    return http.get('/qg/monthly-assessments/attendance-preview', {
      params: { apply_id: applyId, year, month }
    })
  }
}

// 薪酬 API
export const qgPayrollApi = {
  compute(data) {
    return http.post('/qg/payrolls/compute', data)
  },
  list(params) {
    return http.get('/qg/payrolls', { params })
  },
  get(id) {
    return http.get(`/qg/payrolls/${id}`)
  },
  review(id) {
    return http.post(`/qg/payrolls/${id}/review`)
  },
  pay(id) {
    return http.post(`/qg/payrolls/${id}/pay`)
  }
}
