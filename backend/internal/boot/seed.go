package boot

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"student-system/internal/models"
	authservice "student-system/internal/modules/auth/service"
	"student-system/pkg/cryptox"
)

// SeedAdmin 启动时若无任何用户则创建 admin/admin@123，并赋角色 R-SY-ADMIN。
func SeedAdmin(db *gorm.DB, zlog *zap.Logger) {
	var count int64
	db.Model(&models.SysUser{}).Where("is_deleted = 0").Count(&count)
	if count > 0 {
		zlog.Info("seed skipped: users already exist", zap.Int64("count", count))
		return
	}

	zlog.Info("seeding admin user and roles...")

	// 1. 创建 R-SY-ADMIN 角色
	var adminRole models.SysRole
	if err := db.Where("code = ?", "R-SY-ADMIN").First(&adminRole).Error; err != nil {
		adminRole = models.SysRole{
			Code:        "R-SY-ADMIN",
			Name:        "系统管理员",
			Scope:       "school",
			Description: "校级系统管理员，拥有所有模块权限",
		}
		if err := db.Create(&adminRole).Error; err != nil {
			zlog.Error("seed admin role failed", zap.Error(err))
			return
		}
		zlog.Info("seeded role", zap.String("code", adminRole.Code))
	}

	// 2. 创建 admin 用户
	hash, err := authservice.HashPassword("admin@123")
	if err != nil {
		zlog.Error("hash password failed", zap.Error(err))
		return
	}

	adminUser := models.SysUser{
		Username:     "admin",
		PasswordHash: hash,
		DisplayName:  "系统管理员",
		Status:       "active",
	}
	if err := db.Create(&adminUser).Error; err != nil {
		zlog.Error("seed admin user failed", zap.Error(err))
		return
	}
	zlog.Info("seeded user", zap.String("username", adminUser.Username), zap.Int64("id", adminUser.ID))

	// 3. 关联用户-角色
	userRole := models.SysUserRole{
		UserID:    adminUser.ID,
		RoleID:    adminRole.ID,
		GrantedBy: &adminUser.ID,
	}
	if err := db.Create(&userRole).Error; err != nil {
		zlog.Error("seed user-role failed", zap.Error(err))
		return
	}
	zlog.Info("seeded user-role", zap.Int64("user_id", adminUser.ID), zap.Int64("role_id", adminRole.ID))

	// 4. 创建其他基础角色
	roles := []models.SysRole{
		{Code: "R-SY-LEAGUE", Name: "校团委管理员", Scope: "school", Description: "校级团委管理员"},
		{Code: "R-SY-AFFAIRS", Name: "学生处管理员", Scope: "school", Description: "学生处管理员"},
		{Code: "R-COL-LEAGUE", Name: "院系团委书记", Scope: "college", Description: "院系级团委书记"},
		{Code: "R-COL-COUN", Name: "院系辅导员", Scope: "college", Description: "院系级辅导员"},
		{Code: "R-COL-TUTOR", Name: "社团指导教师", Scope: "college", Description: "社团指导教师"},
		{Code: "R-STU-NORM", Name: "普通学生", Scope: "student", Description: "普通学生"},
		{Code: "R-STU-LEAGUE", Name: "团支书", Scope: "student", Description: "团支部书记"},
		{Code: "R-STU-ASSOC", Name: "社团社长/干部", Scope: "student", Description: "社团干部"},
		{Code: "R-STU-COMMUNITY", Name: "楼层长/寝室长", Scope: "student", Description: "社区自治干部"},
	}
	for _, r := range roles {
		var existing models.SysRole
		if err := db.Where("code = ?", r.Code).First(&existing).Error; err != nil {
			if err := db.Create(&r).Error; err != nil {
				zlog.Warn("seed role failed", zap.String("code", r.Code), zap.Error(err))
			}
		}
	}

	zlog.Info("seed completed")
}

// SeedStudentUser 为所有学生创建登录账号。
//
// 登录账号直接使用学号（idx_student.student_no），密码 student@123。
// 已有账号的学生自动跳过；旧版 "student01" 用户名回填为学号。
func SeedStudentUser(db *gorm.DB, zlog *zap.Logger) {
	var students []models.IdxStudent
	db.Where("is_deleted = 0").Order("id ASC").Find(&students)
	if len(students) == 0 {
		return
	}

	// 回填：旧版 seed 用 "student01" 作 username，需改为学号
	var legacyUser models.SysUser
	if err := db.Where("username = ? AND is_deleted = 0", "student01").First(&legacyUser).Error; err == nil {
		if legacyUser.StudentID != nil {
			var stu models.IdxStudent
			if sErr := db.Where("id = ? AND is_deleted = 0", *legacyUser.StudentID).First(&stu).Error; sErr == nil {
				db.Model(&models.SysUser{}).Where("id = ?", legacyUser.ID).
					Updates(map[string]interface{}{"username": stu.StudentNo, "staff_no": ""})
				zlog.Info("backfill student username: student01 -> " + stu.StudentNo)
			}
		}
	}

	// 获取学生角色
	var stuRole models.SysRole
	db.Where("code = ?", "R-STU-NORM").First(&stuRole)

	stuHash, hashErr := authservice.HashPassword("student@123")
	if hashErr != nil {
		return
	}

	created := 0
	for _, stu := range students {
		// 检查是否已有账号（按 student_id 或 username=学号）
		var existing models.SysUser
		if err := db.Where("(student_id = ? OR username = ?) AND is_deleted = 0", stu.ID, stu.StudentNo).First(&existing).Error; err == nil {
			continue // 已有账号，跳过
		}

		user := models.SysUser{
			Username:     stu.StudentNo,
			PasswordHash: stuHash,
			DisplayName:  stu.Name,
			Status:       "active",
			StudentID:    &stu.ID,
		}
		if err := db.Create(&user).Error; err != nil {
			zlog.Warn("seed student user failed", zap.String("student_no", stu.StudentNo), zap.Error(err))
			continue
		}

		// 关联普通学生角色
		if stuRole.ID > 0 {
			ur := models.SysUserRole{
				UserID:    user.ID,
				RoleID:    stuRole.ID,
				GrantedBy: &user.ID,
			}
			db.Create(&ur)
		}
		created++
	}
	if created > 0 {
		zlog.Info("seeded student users", zap.Int("count", created))
	}

	// 顺带播种审批账号
	SeedApprovalUsers(db, zlog)
}

// SeedApprovalUsers 播种三级审批账号。
//
// 登录账号直接使用工号（staff_no），密码 pwd@123：
//   - T001 / pwd@123 → R-COL-COUN，scope_college_id=第一院系
//   - T002 / pwd@123 → R-COL-LEAGUE，scope_college_id=第一院系
//   - T003 / pwd@123 → R-SY-LEAGUE
func SeedApprovalUsers(db *gorm.DB, zlog *zap.Logger) {
	var firstCollege models.SysCollege
	if err := db.Where("is_deleted = 0").Order("id ASC").First(&firstCollege).Error; err != nil {
		return
	}

	specs := []struct {
		OldUsername string // 旧版 username，用于回填
		Username    string // 新 username = 工号
		StaffNo     string
		Display     string
		RoleCode    string
		WithSco     bool
	}{
		{"counselor01", "T001", "T001", "张辅导员", "R-COL-COUN", true},
		{"college01", "T002", "T002", "李院系团委", "R-COL-LEAGUE", true},
		{"league01", "T003", "T003", "王校团委", "R-SY-LEAGUE", false},
	}

	for _, sp := range specs {
		// 回填：旧版用 counselor01 等作 username，需改为工号
		var existing models.SysUser
		if err := db.Where("username = ? AND is_deleted = 0", sp.OldUsername).First(&existing).Error; err == nil {
			db.Model(&models.SysUser{}).Where("id = ?", existing.ID).
				Updates(map[string]interface{}{"username": sp.Username, "staff_no": sp.StaffNo})
			zlog.Info("backfill teacher username: " + sp.OldUsername + " -> " + sp.Username)
			continue
		}
		// 已是新版 username（工号），跳过
		if err := db.Where("username = ? AND is_deleted = 0", sp.Username).First(&existing).Error; err == nil {
			continue
		}

		hash, err := authservice.HashPassword("pwd@123")
		if err != nil {
			continue
		}
		u := models.SysUser{
			Username:     sp.Username,
			PasswordHash: hash,
			StaffNo:      sp.StaffNo,
			DisplayName:  sp.Display,
			Status:       "active",
		}
		if err := db.Create(&u).Error; err != nil {
			zlog.Warn("seed approval user failed", zap.String("username", sp.Username), zap.Error(err))
			continue
		}

		var role models.SysRole
		if err := db.Where("code = ?", sp.RoleCode).First(&role).Error; err != nil {
			continue
		}
		ur := models.SysUserRole{
			UserID:    u.ID,
			RoleID:    role.ID,
			GrantedBy: &u.ID,
		}
		if sp.WithSco {
			cid := firstCollege.ID
			ur.ScopeCollegeID = &cid
			ur.ScopeOrgType = "college"
		}
		if err := db.Create(&ur).Error; err != nil {
			zlog.Warn("seed approval user-role failed", zap.String("username", sp.Username), zap.Error(err))
		}
		zlog.Info("seeded approval user", zap.String("username", sp.Username), zap.String("role", sp.RoleCode))
	}
}

// SeedDicts 播种字典数据。
func SeedDicts(db *gorm.DB, zlog *zap.Logger) {
	var count int64
	db.Model(&models.SysDict{}).Where("is_deleted = 0").Count(&count)
	if count > 0 {
		zlog.Info("seed dicts skipped: already exist", zap.Int64("count", count))
		return
	}

	zlog.Info("seeding dict data...")

	dicts := []models.SysDict{
		// 性别 101-109
		{ID: 101, Category: "gender", Code: "M", NameZh: "男", Sort: 1},
		{ID: 102, Category: "gender", Code: "F", NameZh: "女", Sort: 2},
		{ID: 103, Category: "gender", Code: "U", NameZh: "未知", Sort: 3},

		// 政治面貌 201-209（编号按 Sort 顺序：党员→群众）
		{ID: 201, Category: "political_status", Code: "party_member", NameZh: "中共党员", Sort: 1},
		{ID: 202, Category: "political_status", Code: "party_probationary", NameZh: "预备党员", Sort: 2},
		{ID: 203, Category: "political_status", Code: "member", NameZh: "共青团员", Sort: 3},
		{ID: 204, Category: "political_status", Code: "probationary", NameZh: "预备团员", Sort: 4},
		{ID: 205, Category: "political_status", Code: "activist", NameZh: "入团积极分子", Sort: 5},
		{ID: 206, Category: "political_status", Code: "masses", NameZh: "群众", Sort: 6},

		// 活动等级 301-309
		{ID: 301, Category: "activity_level", Code: "A", NameZh: "A级（跨校/省/全国）", Sort: 1},
		{ID: 302, Category: "activity_level", Code: "B", NameZh: "B级（跨院系/500人+）", Sort: 2},
		{ID: 303, Category: "activity_level", Code: "C", NameZh: "C级（院系内/100人+）", Sort: 3},
		{ID: 304, Category: "activity_level", Code: "D", NameZh: "D级（100人以下）", Sort: 4},

		// 岗位类型 401-409
		{ID: 401, Category: "position_type", Code: "admin", NameZh: "行政办公", Sort: 1},
		{ID: 402, Category: "position_type", Code: "teaching", NameZh: "教学辅助", Sort: 2},
		{ID: 403, Category: "position_type", Code: "research", NameZh: "科研助理", Sort: 3},
		{ID: 404, Category: "position_type", Code: "culture", NameZh: "校园文化", Sort: 4},

		// 困难等级 501-509
		{ID: 501, Category: "difficulty_level", Code: "special", NameZh: "特别困难", Sort: 1},
		{ID: 502, Category: "difficulty_level", Code: "hard", NameZh: "困难", Sort: 2},
		{ID: 503, Category: "difficulty_level", Code: "normal", NameZh: "一般困难", Sort: 3},
		{ID: 504, Category: "difficulty_level", Code: "none", NameZh: "不困难", Sort: 4},

		// 社团状态 601-609
		{ID: 601, Category: "assoc_status", Code: "preparing", NameZh: "筹备中", Sort: 1},
		{ID: 602, Category: "assoc_status", Code: "trial", NameZh: "试运行", Sort: 2},
		{ID: 603, Category: "assoc_status", Code: "registered", NameZh: "注册成立", Sort: 3},
		{ID: 604, Category: "assoc_status", Code: "rectifying", NameZh: "评估整顿", Sort: 4},
		{ID: 605, Category: "assoc_status", Code: "cancelled", NameZh: "注销", Sort: 5},

		// 民族 701-709
		{ID: 701, Category: "ethnicity", Code: "han", NameZh: "汉族", Sort: 1},
		{ID: 702, Category: "ethnicity", Code: "zhuang", NameZh: "壮族", Sort: 2},
		{ID: 703, Category: "ethnicity", Code: "hui", NameZh: "回族", Sort: 3},
		{ID: 704, Category: "ethnicity", Code: "manchu", NameZh: "满族", Sort: 4},
		{ID: 705, Category: "ethnicity", Code: "uyghur", NameZh: "维吾尔族", Sort: 5},
		{ID: 706, Category: "ethnicity", Code: "miao", NameZh: "苗族", Sort: 6},
		{ID: 799, Category: "ethnicity", Code: "other", NameZh: "其他", Sort: 99},

		// 学生状态 801-809
		{ID: 801, Category: "student_status", Code: "enrolled", NameZh: "在读", Sort: 1},
		{ID: 802, Category: "student_status", Code: "suspended", NameZh: "休学", Sort: 2},
		{ID: 803, Category: "student_status", Code: "graduated", NameZh: "毕业", Sort: 3},
		{ID: 804, Category: "student_status", Code: "withdrawn", NameZh: "退学", Sort: 4},

		// 巡查类型 901-909
		{ID: 901, Category: "inspection_type", Code: "hygiene", NameZh: "卫生巡查", Sort: 1},
		{ID: 902, Category: "inspection_type", Code: "late_return", NameZh: "晚归检查", Sort: 2},
		{ID: 903, Category: "inspection_type", Code: "illegal_appliance", NameZh: "违规电器", Sort: 3},
		{ID: 904, Category: "inspection_type", Code: "safety", NameZh: "安全隐患", Sort: 4},
		{ID: 905, Category: "inspection_type", Code: "fire_passage", NameZh: "消防通道", Sort: 5},

		// 事件等级 1001-1009
		{ID: 1001, Category: "incident_level", Code: "L1", NameZh: "L1-常规报修", Sort: 1},
		{ID: 1002, Category: "incident_level", Code: "L2", NameZh: "L2-违规/矛盾", Sort: 2},
		{ID: 1003, Category: "incident_level", Code: "L3", NameZh: "L3-聚众/打架/隐患", Sort: 3},
		{ID: 1004, Category: "incident_level", Code: "L4", NameZh: "L4-火警/群体/伤害", Sort: 4},
	}

	for _, d := range dicts {
		// is_active 默认为 1（启用）
		isActive := d.IsActive
		if isActive == 0 {
			isActive = 1
		}
		if err := db.Exec(
			"INSERT INTO sys_dict (id, category, code, name_zh, name_en, sort, extra_json, is_active, is_deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			d.ID, d.Category, d.Code, d.NameZh, d.NameEn, d.Sort, d.ExtraJSON, isActive, d.IsDeleted,
		).Error; err != nil {
			zlog.Warn("seed dict failed", zap.String("category", d.Category), zap.String("code", d.Code), zap.Error(err))
		}
	}
	zlog.Info("seed dicts completed", zap.Int("count", len(dicts)))
}

// SeedMenus 播种菜单数据（按 code 逐条 upsert，已存在则跳过）。
func SeedMenus(db *gorm.DB, zlog *zap.Logger) {
	zlog.Info("seeding menu data (upsert by code)...")

	// adminRoles 管理类角色
	adminRoles, _ := json.Marshal([]string{"R-SY-ADMIN", "R-SY-LEAGUE", "R-SY-AFFAIRS", "R-COL-LEAGUE", "R-COL-COUN"})
	// sysOnly 仅系统管理员
	sysOnly, _ := json.Marshal([]string{"R-SY-ADMIN"})
	// allRoles 所有角色可见（空数组）
	allRoles := "[]"

	menus := []models.SysMenu{
		// 一级菜单
		{Code: "dashboard", Title: "工作台", Icon: "Odometer", Path: "/dashboard", Component: "views/Dashboard.vue", Sort: 1, Roles: allRoles},
		{Code: "ty", Title: "团员发展", Icon: "Flag", Path: "/ty", Sort: 2, Roles: string(adminRoles)},
		{Code: "st", Title: "社团活动", Icon: "Trophy", Path: "/st", Sort: 3, Roles: string(adminRoles)},
		{Code: "sq", Title: "学生社区", Icon: "House", Path: "/sq", Sort: 4, Roles: string(adminRoles)},
		{Code: "qg", Title: "勤工助学", Icon: "Briefcase", Path: "/qg", Sort: 5, Roles: string(adminRoles)},
		{Code: "cmp", Title: "综合看板", Icon: "DataAnalysis", Path: "/cmp", Sort: 6, Roles: string(adminRoles)},
		{Code: "mine", Title: "我的申请", Icon: "Document", Path: "/mine", Sort: 7, Roles: allRoles},
		{Code: "idx", Title: "学生管理", Icon: "User", Path: "/idx", Sort: 7, Roles: string(adminRoles)},
		{Code: "sys", Title: "系统管理", Icon: "Setting", Path: "/sys", Sort: 99, Roles: string(sysOnly)},

		// TY 子菜单
		{Code: "ty-application", Title: "入团申请", Icon: "", Path: "/ty/application", Component: "views/ty/ApplicationList.vue", Sort: 1, Roles: string(adminRoles)},
		{Code: "ty-approval", Title: "审批中心", Icon: "", Path: "/ty/approval", Component: "views/ty/ApprovalCenter.vue", Sort: 2, Roles: string(adminRoles)},
		{Code: "ty-recommendation-meeting", Title: "推优大会", Icon: "", Path: "/ty/recommendation-meeting", Component: "views/ty/RecommendationMeetingList.vue", Sort: 3, Roles: string(adminRoles)},
		{Code: "ty-cultivation", Title: "培养记录", Icon: "", Path: "/ty/cultivation", Component: "views/ty/CultivationView.vue", Sort: 4, Roles: string(adminRoles)},
		{Code: "ty-development-object", Title: "发展对象", Icon: "", Path: "/ty/development-object", Component: "views/ty/DevelopmentObjectView.vue", Sort: 5, Roles: string(adminRoles)},
		{Code: "ty-political-review", Title: "政审管理", Icon: "", Path: "/ty/political-review", Component: "views/ty/PoliticalReviewView.vue", Sort: 6, Roles: string(adminRoles)},
		{Code: "ty-development-meeting", Title: "发展大会", Icon: "", Path: "/ty/development-meeting", Component: "views/ty/DevelopmentMeetingView.vue", Sort: 7, Roles: string(adminRoles)},
		{Code: "ty-probationary", Title: "转正流程", Icon: "", Path: "/ty/probationary", Component: "views/ty/ProbationaryView.vue", Sort: 8, Roles: string(adminRoles)},
		{Code: "ty-member-roster", Title: "团员花名册", Icon: "", Path: "/ty/member-roster", Component: "views/ty/MemberRoster.vue", Sort: 9, Roles: string(adminRoles)},

		// ST 子菜单
		{Code: "st-association", Title: "社团管理", Icon: "", Path: "/st/association", Component: "views/st/AssociationList.vue", Sort: 1, Roles: string(adminRoles)},
		{Code: "st-activity", Title: "活动管理", Icon: "", Path: "/st/activity", Component: "views/st/ActivityList.vue", Sort: 2, Roles: string(adminRoles)},

		// SQ 子菜单
		{Code: "sq-building", Title: "楼栋管理", Icon: "", Path: "/sq/building", Component: "views/sq/BuildingTree.vue", Sort: 1, Roles: string(adminRoles)},
		{Code: "sq-inspection", Title: "巡查记录", Icon: "", Path: "/sq/inspection", Component: "views/sq/InspectionList.vue", Sort: 2, Roles: string(adminRoles)},
		{Code: "sq-incident", Title: "异常事件", Icon: "", Path: "/sq/incident", Component: "views/sq/IncidentList.vue", Sort: 3, Roles: string(adminRoles)},

		// QG 子菜单
		{Code: "qg-difficulty", Title: "困难认定", Icon: "", Path: "/qg/difficulty", Component: "views/qg/DifficultyList.vue", Sort: 1, Roles: string(adminRoles)},
		{Code: "qg-position", Title: "岗位管理", Icon: "", Path: "/qg/position", Component: "views/qg/PositionList.vue", Sort: 2, Roles: string(adminRoles)},
		{Code: "qg-attendance", Title: "工时打卡", Icon: "", Path: "/qg/attendance", Component: "views/qg/AttendanceRecord.vue", Sort: 3, Roles: string(adminRoles)},

		// CMP 子菜单
		{Code: "cmp-dashboard", Title: "管理驾驶舱", Icon: "", Path: "/cmp/dashboard", Component: "views/cmp/Dashboard.vue", Sort: 1, Roles: string(adminRoles)},
		{Code: "cmp-ranking", Title: "综合分排行", Icon: "", Path: "/cmp/ranking", Component: "views/cmp/ScoreRanking.vue", Sort: 2, Roles: string(adminRoles)},

		// 我的申请 子菜单
		{Code: "mine-development", Title: "我的团员发展", Icon: "", Path: "/mine/ty-development", Component: "views/ty/MyDevelopment.vue", Sort: 1, Roles: allRoles},
		{Code: "mine-application", Title: "我的入团申请", Icon: "", Path: "/mine/ty-application", Component: "views/ty/ApplicationList.vue", Sort: 2, Roles: allRoles},
		{Code: "mine-thought-report", Title: "我的思想汇报", Icon: "", Path: "/mine/thought-report", Component: "views/ty/MyThoughtReport.vue", Sort: 3, Roles: allRoles},
		{Code: "mine-activity", Title: "我的社团", Icon: "", Path: "/mine/activity", Component: "views/st/ActivityList.vue", Sort: 4, Roles: allRoles},
		{Code: "mine-work", Title: "我的勤工", Icon: "", Path: "/mine/work", Component: "views/qg/AttendanceRecord.vue", Sort: 5, Roles: allRoles},
		{Code: "mine-score", Title: "我的综合分", Icon: "", Path: "/mine/score", Component: "views/cmp/MyScore.vue", Sort: 6, Roles: allRoles},
		{Code: "mine-profile", Title: "我的档案", Icon: "", Path: "/mine/profile", Component: "views/idx/MyProfile.vue", Sort: 7, Roles: allRoles},

		// IDX 学生管理 子菜单
		{Code: "idx-student", Title: "学生列表", Icon: "", Path: "/idx/student", Component: "views/idx/StudentList.vue", Sort: 1, Roles: string(adminRoles)},
		{Code: "idx-import", Title: "学生导入", Icon: "", Path: "/idx/import", Component: "views/idx/StudentImport.vue", Sort: 2, Roles: string(adminRoles)},

		// SYS 子菜单
		{Code: "sys-dict", Title: "字典管理", Icon: "", Path: "/sys/dict", Component: "views/sys/DictManage.vue", Sort: 1, Roles: string(sysOnly)},
		{Code: "sys-user", Title: "用户管理", Icon: "", Path: "/sys/user", Component: "views/sys/UserManage.vue", Sort: 2, Roles: string(sysOnly)},
		{Code: "sys-org", Title: "组织管理", Icon: "", Path: "/sys/org", Component: "views/sys/OrgManage.vue", Sort: 3, Roles: string(sysOnly)},
		{Code: "sys-job", Title: "任务监控", Icon: "", Path: "/sys/job", Component: "views/sys/JobMonitor.vue", Sort: 4, Roles: string(sysOnly)},

		// 通知中心（所有角色可见）
		{Code: "noti", Title: "通知中心", Icon: "Bell", Path: "/notifications", Component: "views/notifications/NotificationCenter.vue", Sort: 8, Roles: allRoles},
	}

	// 先 upsert 一级菜单，获取 ID 后再设置 ParentID
	parentMap := map[string]int64{} // code -> id
	for i := range menus {
		// 显式赋 visible=1，避免 Go 零值导致菜单被「WHERE visible=1」过滤掉
		menus[i].Visible = 1
		isTop := menus[i].ParentID == nil && !isSubMenu(menus[i].Code)
		if isTop {
			upsertMenuByCode(db, zlog, &menus[i])
			parentMap[menus[i].Code] = menus[i].ID
		}
	}

	// upsert 子菜单
	for i := range menus {
		menus[i].Visible = 1
		if isSubMenu(menus[i].Code) {
			parentCode := getParentCode(menus[i].Code)
			if parentID, ok := parentMap[parentCode]; ok {
				menus[i].ParentID = &parentID
				upsertMenuByCode(db, zlog, &menus[i])
			}
		}
	}

	// 兜底：把历史数据中 visible=0 的菜单强制修正为 1
	if err := db.Model(&models.SysMenu{}).
		Where("is_deleted = 0 AND visible = 0").
		Update("visible", 1).Error; err != nil {
		zlog.Warn("backfill menu visible failed", zap.Error(err))
	}

	zlog.Info("seed menus completed", zap.Int("count", len(menus)))
}

// upsertMenuByCode 按 code 查找菜单：已存在则同步关键字段（title/path/component/parent_id/icon/sort/roles/visible），
// 不存在则插入。这样多次启动后菜单不会和 seed 源漂移。
func upsertMenuByCode(db *gorm.DB, zlog *zap.Logger, m *models.SysMenu) {
	var existing models.SysMenu
	err := db.Where("code = ? AND is_deleted = 0", m.Code).First(&existing).Error
	if err == nil {
		// 已存在：只同步关键展示字段，保留历史 id 与其它字段
		updates := map[string]interface{}{}
		if existing.Title != m.Title {
			updates["title"] = m.Title
		}
		if existing.Path != m.Path {
			updates["path"] = m.Path
		}
		if existing.Component != m.Component {
			updates["component"] = m.Component
		}
		if existing.Icon != m.Icon {
			updates["icon"] = m.Icon
		}
		if existing.Sort != m.Sort {
			updates["sort"] = m.Sort
		}
		if existing.Roles != m.Roles {
			updates["roles"] = m.Roles
		}
		if existing.Visible != m.Visible {
			updates["visible"] = m.Visible
		}
		if !equalParentID(existing.ParentID, m.ParentID) {
			updates["parent_id"] = m.ParentID
		}
		if len(updates) > 0 {
			if err := db.Model(&models.SysMenu{}).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
				zlog.Warn("sync menu fields failed", zap.String("code", m.Code), zap.Error(err))
			}
		}
		// 回填 ID 以便后续引用
		m.ID = existing.ID
		return
	}
	if err := db.Create(m).Error; err != nil {
		zlog.Warn("seed menu failed", zap.String("code", m.Code), zap.Error(err))
	}
}

// equalParentID 比较两个 ParentID 指针是否指向相同的 parent。
func equalParentID(a, b *int64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// isSubMenu 判断是否为子菜单（包含 `-` 且前缀匹配一级菜单 code）。
func isSubMenu(code string) bool {
	topMenus := []string{"ty", "st", "sq", "qg", "mine", "idx", "sys", "cmp", "noti"}
	for _, t := range topMenus {
		if code == t {
			return false
		}
		if len(code) > len(t) && code[:len(t)] == t && code[len(t)] == '-' {
			return true
		}
	}
	return false
}

// getParentCode 从子菜单 code 提取父菜单 code。
func getParentCode(code string) string {
	topMenus := []string{"ty", "st", "sq", "qg", "mine", "idx", "sys", "cmp", "noti"}
	for _, t := range topMenus {
		if len(code) > len(t) && code[:len(t)] == t && code[len(t)] == '-' {
			return t
		}
	}
	return ""
}

// SeedTyBranches 播种团支部数据（依赖 sys_college 已存在）。
func SeedTyBranches(db *gorm.DB, zlog *zap.Logger) {
	var count int64
	db.Model(&models.TyBranch{}).Where("is_deleted = 0").Count(&count)
	if count > 0 {
		zlog.Info("seed ty_branches skipped: already exist", zap.Int64("count", count))
		return
	}

	// 确保至少有院系数据
	var colleges []models.SysCollege
	db.Where("is_deleted = 0").Find(&colleges)
	if len(colleges) == 0 {
		zlog.Info("seed ty_branches skipped: no colleges found")
		return
	}

	zlog.Info("seeding ty_branch data...")

	// 为每个院系创建一个团支部
	branches := make([]models.TyBranch, 0, len(colleges))
	for i, c := range colleges {
		bizNo := fmt.Sprintf("TY-B-%04d", i+1)
		branches = append(branches, models.TyBranch{
			BizNo:               bizNo,
			Name:                c.Name + "团支部",
			CollegeID:           c.ID,
			ExpectedMemberCount: 30,
		})
	}

	for _, b := range branches {
		if err := db.Create(&b).Error; err != nil {
			zlog.Warn("seed ty_branch failed", zap.String("biz_no", b.BizNo), zap.Error(err))
		}
	}
	zlog.Info("seed ty_branches completed", zap.Int("count", len(branches)))
}

type seedCollegeSpec struct {
	Code   string
	Name   string
	NameEn string
	Majors []seedMajorSpec
}

type seedMajorSpec struct {
	Code          string
	Name          string
	CounselorNo   string
	CounselorName string
}

var seedStudentNames = []string{
	"林亦辰", "陈思妍", "顾明轩", "许知夏", "沈嘉禾", "周雨桐", "陆景行", "唐语嫣",
	"苏沐阳", "叶清欢", "何嘉言", "宋一诺", "秦若曦", "韩子墨", "程予安", "夏星澜",
	"郑可欣", "冯宇航", "蒋依依", "邓皓然", "曾雨薇", "彭子涵", "曹若琳", "袁浩宇",
	"谢梓萱", "董嘉琪", "潘奕辰", "于思远", "余婉清", "杜明哲", "叶梓晨", "蔡欣怡",
	"梁泽宇", "薛芷晴", "魏晨曦", "邹嘉懿", "石沐辰", "高语彤", "罗文昊", "龙思源",
	"马若溪", "丁一凡", "方子墨", "田雨欣", "任嘉宁", "姜睿哲", "范佳怡", "熊浩然",
	"秦子昂", "白若涵", "江明远", "金语晨", "贺思齐", "尹嘉悦", "阎梓豪", "邱沐阳",
	"崔雨泽", "康诗涵", "卢俊熙", "孟欣然", "段奕凡", "汪子琪", "戴明轩", "邵佳宁",
	"姚思睿", "侯若曦", "毛嘉诚", "孔雨萌", "武子轩", "汤语菲", "邢浩辰", "赖欣妍",
	"龚一鸣", "乔梓涵", "黎书瑶", "翟宇辰", "严可馨", "牛嘉乐", "温思远", "雷雨晴",
	"毕晨阳", "郝若彤", "安子涵", "常嘉瑞", "傅诗琪", "卜浩宇", "顾清扬", "万欣悦",
	"席子昂", "祁雨桐", "穆嘉豪", "盛一诺", "童语嫣", "欧阳明轩", "司雨辰", "上官若琳",
	"林嘉树", "陈安宁", "顾云舟", "许念初", "沈知远", "周清越", "陆星河", "唐书瑶",
	"苏景辰", "叶南星", "何沐白", "宋予棠", "秦景行", "韩知许", "程星野", "夏晚晴",
	"郑云舒", "冯清和", "蒋南乔", "邓修远", "曾星语", "彭若安", "曹嘉音", "袁景澄",
	"谢清宁", "董亦航", "潘知意", "于怀瑾", "余景宁", "杜若川", "叶澄心", "蔡安琪",
	"梁知夏", "薛景然", "魏嘉木", "邹云溪", "石清扬", "高若宁",
}

var seedColleges = []seedCollegeSpec{
	{
		Code:   "CS",
		Name:   "计算机学院",
		NameEn: "Computer Science",
		Majors: []seedMajorSpec{
			{Code: "AI", Name: "人工智能技术应用", CounselorNo: "T101", CounselorName: "赵敏"},
			{Code: "SE", Name: "软件技术", CounselorNo: "T102", CounselorName: "钱芳"},
		},
	},
	{
		Code:   "EE",
		Name:   "电子工程学院",
		NameEn: "Electronic Engineering",
		Majors: []seedMajorSpec{
			{Code: "IOT", Name: "物联网应用技术", CounselorNo: "T201", CounselorName: "孙磊"},
			{Code: "EC", Name: "电子信息工程技术", CounselorNo: "T202", CounselorName: "周宁"},
		},
	},
}

// SeedColleges 播种院系、专业、辅导员和行政班数据。
func SeedColleges(db *gorm.DB, zlog *zap.Logger) {
	createdColleges := 0
	createdMajors := 0
	createdCounselors := 0
	createdClasses := 0

	for _, collegeSpec := range seedColleges {
		college, collegeCreated := upsertCollege(db, zlog, collegeSpec)
		if college.ID == 0 {
			continue
		}
		if collegeCreated {
			createdColleges++
		}

		for _, majorSpec := range collegeSpec.Majors {
			major, majorCreated := upsertMajor(db, zlog, college.ID, majorSpec)
			if major.ID == 0 {
				continue
			}
			if majorCreated {
				createdMajors++
			}

			counselor, counselorCreated := upsertCounselor(db, zlog, college.ID, majorSpec)
			if counselorCreated {
				createdCounselors++
			}

			for _, grade := range []int{2023, 2024} {
				for classNo := 1; classNo <= 2; classNo++ {
					if upsertClass(db, zlog, major.ID, grade, classNo, majorSpec.Name, counselor.ID) {
						createdClasses++
					}
				}
			}
		}
	}

	zlog.Info("seed organization completed",
		zap.Int("colleges", createdColleges),
		zap.Int("majors", createdMajors),
		zap.Int("counselors", createdCounselors),
		zap.Int("classes", createdClasses),
	)
}

// SeedStudents 播种测试学生数据。
func SeedStudents(db *gorm.DB, zlog *zap.Logger) {
	var classes []models.IdxClass
	if err := db.Where("is_deleted = 0").Order("grade ASC, major_id ASC, code ASC").Find(&classes).Error; err != nil {
		zlog.Warn("seed students skipped: query classes failed", zap.Error(err))
		return
	}
	if len(classes) == 0 {
		zlog.Warn("seed students skipped: no classes found")
		return
	}

	created := 0
	studentSeq := 0
	for _, class := range classes {
		var major models.SysMajor
		if err := db.Where("id = ? AND is_deleted = 0", class.MajorID).First(&major).Error; err != nil {
			continue
		}

		var college models.SysCollege
		if err := db.Where("id = ? AND is_deleted = 0", major.CollegeID).First(&college).Error; err != nil {
			continue
		}

		enrollDate, _ := time.Parse("2006-01-02", fmt.Sprintf("%d-09-01", class.Grade))
		for i := 0; i < 8; i++ {
			name := seedStudentNames[studentSeq%len(seedStudentNames)]
			studentSeq++
			studentNo := fmt.Sprintf("%d%02d%02d%02d", class.Grade, major.ID, class.ID, i+1)
			var existing models.IdxStudent
			if err := db.Where("student_no = ? AND is_deleted = 0", studentNo).First(&existing).Error; err == nil {
				continue
			}

			idCard := fmt.Sprintf("310115%d%02d%02d%04d", 2004+i%3, (i%12)+1, (i%28)+1, class.ID*10+int64(i)+1)
			phone := fmt.Sprintf("138%08d", class.ID*100+int64(i)+1)
			birth, _ := time.Parse("2006-01-02", fmt.Sprintf("%d-%02d-%02d", 2004+i%3, (i%12)+1, (i%28)+1))
			gender := "M"
			if i%2 == 1 {
				gender = "F"
			}
			politicalStatus := "masses"
			if i%3 == 0 {
				politicalStatus = "member"
			}

			collegeID := college.ID
			majorID := major.ID
			classID := class.ID
			grade := class.Grade
			student := models.IdxStudent{
				StudentNo:       studentNo,
				Name:            name,
				IDCardEnc:       cryptox.Encrypt(idCard),
				IDCardHash:      fmt.Sprintf("%x", sha256.Sum256([]byte(idCard))),
				Gender:          gender,
				BirthDate:       &birth,
				Ethnicity:       "汉族",
				PoliticalStatus: politicalStatus,
				CollegeID:       &collegeID,
				MajorID:         &majorID,
				ClassID:         &classID,
				Grade:           &grade,
				PhoneEnc:        cryptox.Encrypt(phone),
				PhoneHash:       fmt.Sprintf("%x", sha256.Sum256([]byte(phone))),
				Email:           fmt.Sprintf("%s@example.edu.cn", studentNo),
				EnrollmentAt:    &enrollDate,
				Status:          "enrolled",
			}
			if err := db.Create(&student).Error; err != nil {
				zlog.Warn("seed student failed", zap.String("student_no", student.StudentNo), zap.Error(err))
				continue
			}
			created++
		}
	}
	zlog.Info("seed students completed", zap.Int("count", created))
}

func upsertCollege(db *gorm.DB, zlog *zap.Logger, spec seedCollegeSpec) (models.SysCollege, bool) {
	var college models.SysCollege
	if err := db.Where("code = ? AND is_deleted = 0", spec.Code).First(&college).Error; err == nil {
		return college, false
	}
	college = models.SysCollege{Code: spec.Code, Name: spec.Name, NameEn: spec.NameEn}
	if err := db.Create(&college).Error; err != nil {
		zlog.Warn("seed college failed", zap.String("code", spec.Code), zap.Error(err))
		return models.SysCollege{}, false
	}
	return college, true
}

func upsertMajor(db *gorm.DB, zlog *zap.Logger, collegeID int64, spec seedMajorSpec) (models.SysMajor, bool) {
	var major models.SysMajor
	if err := db.Where("college_id = ? AND code = ? AND is_deleted = 0", collegeID, spec.Code).First(&major).Error; err == nil {
		return major, false
	}
	major = models.SysMajor{CollegeID: collegeID, Code: spec.Code, Name: spec.Name}
	if err := db.Create(&major).Error; err != nil {
		zlog.Warn("seed major failed", zap.String("code", spec.Code), zap.Error(err))
		return models.SysMajor{}, false
	}
	return major, true
}

func upsertCounselor(db *gorm.DB, zlog *zap.Logger, collegeID int64, spec seedMajorSpec) (models.SysUser, bool) {
	var user models.SysUser
	if err := db.Where("username = ? AND is_deleted = 0", spec.CounselorNo).First(&user).Error; err == nil {
		return user, false
	}

	hash, err := authservice.HashPassword("counselor@123")
	if err != nil {
		zlog.Warn("hash counselor password failed", zap.String("staff_no", spec.CounselorNo), zap.Error(err))
		return models.SysUser{}, false
	}
	user = models.SysUser{
		Username:     spec.CounselorNo,
		PasswordHash: hash,
		StaffNo:      spec.CounselorNo,
		DisplayName:  spec.CounselorName,
		Status:       "active",
	}
	if err := db.Create(&user).Error; err != nil {
		zlog.Warn("seed counselor failed", zap.String("staff_no", spec.CounselorNo), zap.Error(err))
		return models.SysUser{}, false
	}

	var role models.SysRole
	if err := db.Where("code = ? AND is_deleted = 0", "R-COL-COUN").First(&role).Error; err == nil {
		cid := collegeID
		ur := models.SysUserRole{UserID: user.ID, RoleID: role.ID, ScopeCollegeID: &cid, ScopeOrgType: "college", GrantedBy: &user.ID}
		if err := db.Create(&ur).Error; err != nil {
			zlog.Warn("seed counselor role failed", zap.String("staff_no", spec.CounselorNo), zap.Error(err))
		}
	}
	return user, true
}

func upsertClass(db *gorm.DB, zlog *zap.Logger, majorID int64, grade int, classNo int, majorName string, counselorID int64) bool {
	code := fmt.Sprintf("%d-%02d", grade, classNo)
	name := fmt.Sprintf("%d%s%d班", grade, majorName, classNo)
	var class models.IdxClass
	if err := db.Where("major_id = ? AND code = ? AND is_deleted = 0", majorID, code).First(&class).Error; err == nil {
		updates := map[string]interface{}{"name": name}
		if counselorID > 0 {
			updates["counselor_id"] = counselorID
		}
		db.Model(&models.IdxClass{}).Where("id = ?", class.ID).Updates(updates)
		return false
	}
	class = models.IdxClass{MajorID: majorID, Grade: grade, Code: code, Name: name}
	if counselorID > 0 {
		class.CounselorID = &counselorID
	}
	if err := db.Create(&class).Error; err != nil {
		zlog.Warn("seed class failed", zap.String("name", name), zap.Error(err))
		return false
	}
	return true
}

// SeedSQData 播种楼栋/楼层/寝室/床位种子数据。
func SeedSQData(db *gorm.DB, zlog *zap.Logger) {
	// 楼层长用户回填/创建（必须在 early return 之前，否则 buildings 已存在时跳过）
	backfillFloorLeader(db, zlog)

	var count int64
	db.Model(&models.IdxDormBuilding{}).Where("is_deleted = 0").Count(&count)
	if count > 0 {
		zlog.Info("seed sq data skipped: buildings already exist", zap.Int64("count", count))
		return
	}

	zlog.Info("seeding sq dorm data...")

	// 1. 楼栋
	buildings := []models.IdxDormBuilding{
		{Code: "BLD-01", Name: "1号楼", FloorCount: 6},
		{Code: "BLD-02", Name: "2号楼", FloorCount: 6},
	}
	for i := range buildings {
		if err := db.Create(&buildings[i]).Error; err != nil {
			zlog.Warn("seed building failed", zap.String("code", buildings[i].Code), zap.Error(err))
		}
	}

	// 2. 楼层
	for bIdx := range buildings {
		for f := 1; f <= buildings[bIdx].FloorCount; f++ {
			floor := models.IdxDormFloor{
				BuildingID: buildings[bIdx].ID,
				FloorNo:    f,
			}
			if err := db.Create(&floor).Error; err != nil {
				zlog.Warn("seed floor failed", zap.Int("floor_no", f), zap.Error(err))
			}

			// 3. 寝室（每层 10 间）
			for r := 1; r <= 10; r++ {
				roomNo := fmt.Sprintf("%d%02d", f, r)
				room := models.IdxDormRoom{
					BuildingID: buildings[bIdx].ID,
					FloorID:    floor.ID,
					RoomNo:     roomNo,
					BedCount:   4,
				}
				if err := db.Create(&room).Error; err != nil {
					zlog.Warn("seed room failed", zap.String("room_no", roomNo), zap.Error(err))
					continue
				}

				// 4. 床位（每间 4 个）
				for bed := 1; bed <= 4; bed++ {
					bedRec := models.IdxDormBed{
						RoomID: room.ID,
						BedNo:  fmt.Sprintf("%d", bed),
					}
					if err := db.Create(&bedRec).Error; err != nil {
						zlog.Warn("seed bed failed", zap.Error(err))
					}
				}
			}
		}
	}

	zlog.Info("seed sq data completed")
}

// backfillFloorLeader 回填/创建楼层长用户（工号 T004，密码 floor@123）。
func backfillFloorLeader(db *gorm.DB, zlog *zap.Logger) {
	var existingFloorUser models.SysUser
	// 回填：旧版用 "floor01" 作 username
	if err := db.Where("username = ? AND is_deleted = 0", "floor01").First(&existingFloorUser).Error; err == nil {
		db.Model(&models.SysUser{}).Where("id = ?", existingFloorUser.ID).
			Updates(map[string]interface{}{"username": "T004", "staff_no": "T004"})
		zlog.Info("backfill floor leader username: floor01 -> T004")
		return
	}
	// 已是新版 username（工号），跳过
	if err := db.Where("username = ? AND is_deleted = 0", "T004").First(&existingFloorUser).Error; err == nil {
		return
	}
	// 新版不存在，创建
	hash, hashErr := authservice.HashPassword("floor@123")
	if hashErr != nil {
		return
	}
	floorUser := models.SysUser{
		Username:     "T004",
		PasswordHash: hash,
		StaffNo:      "T004",
		DisplayName:  "赵楼层长",
		Status:       "active",
	}
	if createErr := db.Create(&floorUser).Error; createErr == nil {
		var communityRole models.SysRole
		if roleErr := db.Where("code = ?", "R-STU-COMMUNITY").First(&communityRole).Error; roleErr == nil {
			ur := models.SysUserRole{
				UserID:    floorUser.ID,
				RoleID:    communityRole.ID,
				GrantedBy: &floorUser.ID,
			}
			db.Create(&ur)
		}
		zlog.Info("seeded floor leader user", zap.String("username", "T004"))
	}
}

// SeedQGData 启动时若无勤工助学数据则创建种子数据。
func SeedQGData(db *gorm.DB, zlog *zap.Logger) {
	var count int64
	db.Model(&models.QgPosition{}).Where("is_deleted = 0").Count(&count)
	if count > 0 {
		zlog.Info("seed qg data skipped: positions already exist", zap.Int64("count", count))
		return
	}

	zlog.Info("seeding qg work-study data...")

	// 1. 创建困难认定（使用已有学生）
	var students []models.IdxStudent
	db.Where("is_deleted = 0").Limit(3).Find(&students)
	if len(students) == 0 {
		zlog.Warn("seed qg data skipped: no students found")
		return
	}

	// 获取管理员用户ID
	var adminUser models.SysUser
	db.Where("username = ? AND is_deleted = 0", "admin").First(&adminUser)
	adminID := adminUser.ID

	for i, stu := range students {
		level := "special"
		if i == 1 {
			level = "hard"
		} else if i == 2 {
			level = "normal"
		}
		cert := models.QgDifficultyCert{
			BizNo:        fmt.Sprintf("QG-DIF-%04d", i+1),
			StudentID:    stu.ID,
			AcademicYear: "2025-2026",
			Level:        level,
			Status:       "S3",
			CreatedBy:    &adminID,
			UpdatedBy:    &adminID,
		}
		if err := db.Create(&cert).Error; err != nil {
			zlog.Warn("seed difficulty cert failed", zap.Error(err))
		} else {
			// 更新学生困难标记
			db.Model(&models.IdxStudent{}).Where("id = ?", stu.ID).
				Updates(map[string]interface{}{"is_difficulty": 1, "difficulty_level": level})
		}
	}

	// 2. 创建岗位（dept_type 必须符合 CHECK 约束: admin/teaching/research/culture）
	positions := []models.QgPosition{
		{
			BizNo: "QG-POS-0001", DeptType: "admin", DeptName: "图书馆",
			Title: "图书整理员", Description: "负责图书分类整理和上架",
			Headcount: 2, WeeklyHoursLimit: 15, HourlyRateCents: 1500,
			StartAt: time.Date(2025, 9, 1, 0, 0, 0, 0, time.Local),
			EndAt:   time.Date(2026, 6, 30, 0, 0, 0, 0, time.Local),
			Status:  "S3", SupervisorUserID: &adminID,
			CreatedBy: &adminID, UpdatedBy: &adminID,
		},
		{
			BizNo: "QG-POS-0002", DeptType: "teaching", DeptName: "食堂",
			Title: "食堂帮厨", Description: "协助食堂日常清洁和配餐",
			Headcount: 3, WeeklyHoursLimit: 20, HourlyRateCents: 1200,
			StartAt: time.Date(2025, 9, 1, 0, 0, 0, 0, time.Local),
			EndAt:   time.Date(2026, 6, 30, 0, 0, 0, 0, time.Local),
			Status:  "S3", SupervisorUserID: &adminID,
			CreatedBy: &adminID, UpdatedBy: &adminID,
		},
		{
			BizNo: "QG-POS-0003", DeptType: "admin", DeptName: "学生处",
			Title: "行政助理", Description: "协助学生处日常行政事务",
			Headcount: 1, WeeklyHoursLimit: 10, HourlyRateCents: 1800,
			StartAt: time.Date(2025, 9, 1, 0, 0, 0, 0, time.Local),
			EndAt:   time.Date(2026, 6, 30, 0, 0, 0, 0, time.Local),
			Status:  "S3", SupervisorUserID: &adminID,
			CreatedBy: &adminID, UpdatedBy: &adminID,
		},
	}
	for i := range positions {
		if err := db.Create(&positions[i]).Error; err != nil {
			zlog.Warn("seed position failed", zap.String("biz_no", positions[i].BizNo), zap.Error(err))
		}
	}

	// 3. 为第一个学生创建岗位申请（在岗状态）
	if len(students) > 0 && len(positions) > 0 {
		apply := models.QgPositionApply{
			BizNo:       "QG-0001",
			PositionID:  positions[0].ID,
			StudentID:   students[0].ID,
			ApplyStatus: "accepted",
			Status:      "on_job",
		}
		if err := db.Create(&apply).Error; err != nil {
			zlog.Warn("seed position apply failed", zap.Error(err))
		}
	}

	zlog.Info("seed qg data completed")
}

// SeedCmpRuleVersion 启动时若无规则版本则创建默认版本（v-default）。
func SeedCmpRuleVersion(db *gorm.DB, zlog *zap.Logger) {
	var count int64
	db.Model(&models.CmpRuleVersion{}).Where("is_deleted = 0").Count(&count)
	if count > 0 {
		zlog.Info("seed cmp rule version skipped: already exist", zap.Int64("count", count))
		return
	}

	zlog.Info("seeding cmp default rule version...")

	defaultRules := `{
  "weights": {
    "league": 0.25,
    "assoc": 0.20,
    "community": 0.15,
    "workstudy": 0.15,
    "academic": 0.25
  },
  "dimensions": [
    { "dimension": "league",    "rules": [
      { "sub_item": "团内身份", "score": 5, "weight": 0.05, "max": 5 },
      { "sub_item": "团内任职", "score": 6, "weight": 0.06, "max": 10 },
      { "sub_item": "团内活动参与", "score": 6, "weight": 0.06, "max": 15 },
      { "sub_item": "推优通过", "score": 4, "weight": 0.04, "max": 10 },
      { "sub_item": "培训结业", "score": 4, "weight": 0.04, "max": 10 }
    ]},
    { "dimension": "assoc",     "rules": [
      { "sub_item": "社团任职", "score": 5, "weight": 0.05, "max": 10 },
      { "sub_item": "活动组织", "score": 8, "weight": 0.08, "max": 15 },
      { "sub_item": "活动参与", "score": 5, "weight": 0.05, "max": 10 },
      { "sub_item": "评优获奖", "score": 2, "weight": 0.02, "max": 5 }
    ]},
    { "dimension": "community", "rules": [
      { "sub_item": "自治职务", "score": 3, "weight": 0.03, "max": 5 },
      { "sub_item": "巡查考核", "score": 5, "weight": 0.05, "max": 10 },
      { "sub_item": "文明寝室", "score": 3, "weight": 0.03, "max": 5 },
      { "sub_item": "事件处置", "score": 4, "weight": 0.04, "max": 10 }
    ]},
    { "dimension": "workstudy", "rules": [
      { "sub_item": "岗位履职", "score": 7, "weight": 0.07, "max": 10 },
      { "sub_item": "工时完成度", "score": 3, "weight": 0.03, "max": 5 },
      { "sub_item": "考核合格", "score": 5, "weight": 0.05, "max": 10 }
    ]},
    { "dimension": "academic",  "rules": [
      { "sub_item": "GPA", "score": 15, "weight": 0.15, "max": 25 },
      { "sub_item": "排名", "score": 10, "weight": 0.10, "max": 20 }
    ]}
  ],
  "academic": {
    "gpa_per_100": 0.25,
    "rank_top_5": 20.0,
    "rank_top_20": 15.0,
    "rank_default": 10.0
  }
}`

	effectiveAt, _ := time.Parse("2006-01-02", "2026-01-01")
	v := models.CmpRuleVersion{
		Version:     "v2026.1",
		RulesJSON:   defaultRules,
		EffectiveAt: effectiveAt,
		IsActive:    1,
	}
	if err := db.Create(&v).Error; err != nil {
		zlog.Warn("seed cmp rule version failed", zap.Error(err))
		return
	}
	zlog.Info("seeded cmp rule version", zap.String("version", v.Version), zap.Int64("id", v.ID))
}

// SeedTyApplicationForLisi 为李四播种入团申请+审批记录（演示发展轨迹用）。
func SeedTyApplicationForLisi(db *gorm.DB, zlog *zap.Logger) {
	var count int64
	db.Model(&models.TyApplication{}).Where("is_deleted = 0").Count(&count)
	if count > 0 {
		zlog.Info("seed ty_application for lisi skipped: applications already exist", zap.Int64("count", count))
		return
	}

	// 查找李四的学生记录
	var student models.IdxStudent
	if err := db.Where("student_no = ? AND is_deleted = 0", "20231001").First(&student).Error; err != nil {
		zlog.Warn("seed ty_application for lisi skipped: student not found", zap.Error(err))
		return
	}

	// 查找团支部
	var branch models.TyBranch
	if err := db.Where("college_id = ? AND is_deleted = 0", student.CollegeID).First(&branch).Error; err != nil {
		zlog.Warn("seed ty_application for lisi skipped: branch not found", zap.Error(err))
		return
	}

	// 查找审批用户
	var counselorUser models.SysUser
	db.Where("username = ? AND is_deleted = 0", "counselor01").First(&counselorUser)
	var collegeUser models.SysUser
	db.Where("username = ? AND is_deleted = 0", "college01").First(&collegeUser)
	var schoolUser models.SysUser
	db.Where("username = ? AND is_deleted = 0", "league01").First(&schoolUser)

	zlog.Info("seeding ty_application for lisi...")

	now := time.Now()
	applyDate := now.AddDate(0, -3, 0) // 3个月前申请

	// 创建入团申请（S3 已通过状态，模拟完整审批流程）
	app := models.TyApplication{
		BizNo:         "TY-2026-0001",
		StudentID:     student.ID,
		BranchID:      branch.ID,
		ApplyDate:     applyDate,
		SelfStatement: "我志愿加入中国共产主义青年团，坚决拥护中国共产党的领导，遵守团的章程，执行团的决议，履行团员义务，严守团的纪律，勤奋学习，积极工作，吃苦在前，享受在后，为共产主义事业而奋斗。\n\n通过对《中国共产主义青年团章程》的系统学习，我对共青团的性质、宗旨和任务有了更加深刻的认识。共青团是中国共产党领导的先进青年的群团组织，是广大青年在实践中学习中国特色社会主义和共产主义的学校，是党的助手和后备军。作为一名新时代的大学生，我深知自己肩负的历史使命和时代责任。\n\n在思想方面，我始终坚持以习近平新时代中国特色社会主义思想为指导，认真学习党的二十大精神，积极参加学校组织的各类思想政治教育活动。我关注时事政治，关心国家大事，努力提高自己的政治觉悟和理论水平。我坚信只有在中国共产党的领导下，青年一代才能健康成长，才能在实现中华民族伟大复兴中国梦的征程中贡献自己的青春力量。\n\n在学习方面，我始终保持严谨求实的学风，认真对待每一门课程，按时完成各项学习任务。我注重理论与实践相结合，积极参与课堂讨论和小组合作学习，不断提升自己的专业素养和综合能力。我相信扎实的专业知识是为人民服务的基础，也是实现人生价值的重要保障。\n\n在工作方面，作为班级的一员，我积极参与班级建设和集体活动，主动承担力所能及的工作任务。我尊敬师长、团结同学，乐于助人，努力营造良好的班级氛围。我也参加了学校组织的志愿服务活动，在服务他人的过程中体会到了奉献的快乐和成长的喜悦。\n\n在生活方面，我严格要求自己，养成良好的生活习惯。我坚持体育锻炼，保持健康的体魄；注重个人品德修养，诚实守信、勤俭节约；尊重他人、关心集体，努力做一个有理想、有道德、有文化、有纪律的新时代好青年。\n\n如果能够加入中国共产主义青年团，我将更加严格地要求自己，在学习和工作中发挥模范带头作用，积极参加团组织的各项活动，认真完成组织交给的任务。我将虚心向优秀团员学习，不断提高自己的思想素质和工作能力，争取早日成为一名合格的共青团员，为团组织的建设和发展贡献自己的力量。请团组织在实践中考验我！",
		FamilyMembers: `[{"name":"李父","relation":"父亲","political_status":"群众"},{"name":"李母","relation":"母亲","political_status":"群众"}]`,
		RewardsPunish: "2024年获得校级优秀学生干部称号",
		Status:        "S3",
		CreatedBy:     &student.ID,
	}
	if err := db.Create(&app).Error; err != nil {
		zlog.Warn("seed ty_application for lisi failed", zap.Error(err))
		return
	}

	// 创建三级审批记录（仅针对入团申请阶段，与实际发展轨迹一致）
	approvalRecords := []models.TyApprovalRecord{
		{
			ApplicationID: app.ID,
			Module:        "application",
			TargetID:      app.ID,
			Step:          "counselor",
			ApproverID:    counselorUser.ID,
			ApproverName:  "张辅导员",
			ApproverRole:  "R-COL-COUN",
			Result:        "approve",
			Opinion:       "该生提交的入团申请书内容详实，对团的认识清晰端正。日常学习态度认真，能主动参与班级集体活动，与同学关系融洽，群众基础良好。经班级团支部讨论通过，同意推荐该生进入入团积极分子培养阶段。",
			FromStatus:    "S1",
			ToStatus:      "S2",
			OccurredAt:    applyDate.AddDate(0, 0, 3),
		},
		{
			ApplicationID: app.ID,
			Module:        "application",
			TargetID:      app.ID,
			Step:          "college",
			ApproverID:    collegeUser.ID,
			ApproverName:  "李院系团委",
			ApproverRole:  "R-COL-LEAGUE",
			Result:        "approve",
			Opinion:       "经院系团委复核，该生入团申请书材料完整规范，思想表现自述真实诚恳，家庭成员情况清楚，无不良记录。辅导员初审意见客观公正。院系层面审核通过，报请校团委终审。",
			FromStatus:    "S2",
			ToStatus:      "S2",
			OccurredAt:    applyDate.AddDate(0, 0, 5),
		},
		{
			ApplicationID: app.ID,
			Module:        "application",
			TargetID:      app.ID,
			Step:          "school",
			ApproverID:    schoolUser.ID,
			ApproverName:  "王校团委",
			ApproverRole:  "R-SY-LEAGUE",
			Result:        "approve",
			Opinion:       "经校团委终审，该生入团申请材料齐全，三级审批流程完整合规。该生政治立场坚定、学习刻苦努力、作风正派诚实，符合《中国共产主义青年团章程》规定的团员发展条件。批准通过入团申请，可进入推优及后续培养流程。",
			FromStatus:    "S2",
			ToStatus:      "S3",
			OccurredAt:    applyDate.AddDate(0, 0, 7),
		},
	}

	for _, rec := range approvalRecords {
		if err := db.Create(&rec).Error; err != nil {
			zlog.Warn("seed ty_approval_record for lisi failed", zap.Error(err))
		}
	}

	// 更新李四的政治面貌为"入团积极分子"（已通过入团申请但尚未完成推优）
	db.Model(&models.IdxStudent{}).Where("id = ?", student.ID).Update("political_status", "activist")

	zlog.Info("seed ty_application for lisi completed",
		zap.Int64("application_id", app.ID),
		zap.String("biz_no", app.BizNo),
		zap.Int("approval_records", len(approvalRecords)),
	)
}

// SeedOtherBusinessData 补充社团、社区、勤工、综合测评等业务测试数据。
func SeedOtherBusinessData(db *gorm.DB, zlog *zap.Logger) {
	seedSTBusinessData(db, zlog)
	seedSQBusinessData(db, zlog)
	seedQGBusinessDetails(db, zlog)
	seedCmpScores(db, zlog)
}

func seedSTBusinessData(db *gorm.DB, zlog *zap.Logger) {
	var count int64
	db.Model(&models.StAssociation{}).Where("is_deleted = 0").Count(&count)
	if count > 0 {
		zlog.Info("seed st business data skipped: associations already exist", zap.Int64("count", count))
		return
	}

	var college models.SysCollege
	if err := db.Where("is_deleted = 0").Order("id ASC").First(&college).Error; err != nil {
		zlog.Warn("seed st business data skipped: no college found", zap.Error(err))
		return
	}
	var admin models.SysUser
	db.Where("username = ? AND is_deleted = 0", "admin").First(&admin)
	var students []models.IdxStudent
	db.Where("is_deleted = 0").Order("id ASC").Limit(12).Find(&students)
	if len(students) < 6 {
		zlog.Warn("seed st business data skipped: students not enough", zap.Int("count", len(students)))
		return
	}

	now := time.Now()
	registeredAt := now.AddDate(-1, 0, 0)
	rating := 4
	assoc := models.StAssociation{
		BizNo:              "ST-2026-0001",
		Name:               "人工智能创新社",
		CollegeID:          college.ID,
		TutorUserID:        &admin.ID,
		PresidentStudentID: &students[0].ID,
		BusinessScope:      "面向全校学生开展人工智能技术学习、项目实践、竞赛训练与科普活动。",
		Status:             "registered",
		RegisteredAt:       &registeredAt,
		FoundedAt:          &registeredAt,
		StarRating:         &rating,
		CreatedBy:          &admin.ID,
		UpdatedBy:          &admin.ID,
	}
	if err := db.Create(&assoc).Error; err != nil {
		zlog.Warn("seed st association failed", zap.Error(err))
		return
	}

	for i, stu := range students[:6] {
		role := "member"
		core := 0
		if i == 0 {
			role = "president"
			core = 1
		} else if i == 1 {
			role = "vice_president"
			core = 1
		} else if i == 2 {
			role = "director"
			core = 1
		}
		db.Create(&models.StFounder{AssociationID: assoc.ID, StudentID: stu.ID, JoinedAt: registeredAt})
		db.Create(&models.StAssocMember{AssociationID: assoc.ID, StudentID: stu.ID, Role: role, JoinedAt: registeredAt, IsCoreOfficer: core})
	}

	interviewAt := now.AddDate(0, -1, 0)
	deadline := now.AddDate(0, -1, 7)
	plan := models.StRecruitPlan{
		BizNo:            "ST-REC-2026-0001",
		AssociationID:    assoc.ID,
		Season:           "spring",
		AcademicYear:     "2025-2026",
		TargetCount:      20,
		AssessmentMethod: "简历筛选 + 项目展示 + 面试",
		InterviewAt:      &interviewAt,
		Status:           "S3",
		ResultDeadline:   &deadline,
	}
	db.Create(&plan)
	for i, stu := range students[6:10] {
		result := "pending"
		if i < 2 {
			result = "accepted"
		}
		db.Create(&models.StRecruitApply{PlanID: plan.ID, StudentID: stu.ID, Result: result, ResultAt: &deadline})
	}

	startedAt := now.AddDate(0, 0, 7)
	endedAt := startedAt.Add(3 * time.Hour)
	expected := 80
	activity := models.StActivity{
		BizNo:                "ST-ACT-2026-0001",
		AssociationID:        assoc.ID,
		Title:                "校园 AI 应用创意工作坊",
		ActivityLevel:        "C",
		ExpectedParticipants: expected,
		ExpectedCount:        &expected,
		BudgetCents:          350000,
		Location:             "创新创业中心 302",
		StartedAt:            startedAt,
		EndedAt:              endedAt,
		Status:               "S3",
		LastAction:           "approve",
	}
	if err := db.Create(&activity).Error; err != nil {
		zlog.Warn("seed st activity failed", zap.Error(err))
		return
	}
	decidedAt := now.AddDate(0, 0, -2)
	db.Create(&models.StActivityApproval{ActivityID: activity.ID, StepNo: 1, ApproverRole: "R-COL-TUTOR", ApproverUserID: &admin.ID, Decision: "pass", Opinion: "活动主题契合专业特色，组织方案完整，安全预案充分，同意开展。", DecidedAt: &decidedAt})
	db.Create(&models.StActivityApproval{ActivityID: activity.ID, StepNo: 2, ApproverRole: "R-SY-LEAGUE", ApproverUserID: &admin.ID, Decision: "pass", Opinion: "活动规模和预算合理，审批材料齐全，符合社团活动管理要求，同意立项。", DecidedAt: &decidedAt})
	for i, stu := range students[:8] {
		late := 0
		lateMinutes := 0
		checkinAt := startedAt.Add(time.Duration(i) * time.Minute)
		if i == 7 {
			late = 1
			lateMinutes = 18
			checkinAt = startedAt.Add(18 * time.Minute)
		}
		db.Create(&models.StActivityCheckin{ActivityID: activity.ID, StudentID: stu.ID, CheckinAt: checkinAt, Method: "qrcode", IsLate: late, LateMinutes: lateMinutes, IsPresent: 1})
	}
	score := 92
	db.Create(&models.StActivitySummary{ActivityID: activity.ID, ActualParticipants: 76, AchievementScore: &score, Suggestions: "活动参与度较高，建议后续增加真实项目路演环节。", SubmittedAt: endedAt.Add(24 * time.Hour), IsOverdue: 0})
	reviewedAt := now
	db.Create(&models.StExpense{BizNo: "ST-EXP-2026-0001", ActivityID: activity.ID, AmountCents: 286000, InvoiceCount: 3, InvoiceFiles: "[]", Status: "S3", ReviewedBy: &admin.ID, ReviewedAt: &reviewedAt, CoSignedBy: &admin.ID, PaidAt: &reviewedAt})
	db.Create(&models.StRating{AssociationID: assoc.ID, AcademicYear: "2025-2026", DimensionActivity: 90, DimensionMemberActive: 86, DimensionFinance: 88, DimensionBrand: 91, DimensionSatisfaction: 93, WeightedScore: 89.8, Star: 4, Status: "S3"})

	zlog.Info("seed st business data completed", zap.Int64("association_id", assoc.ID), zap.Int64("activity_id", activity.ID))
}

func seedSQBusinessData(db *gorm.DB, zlog *zap.Logger) {
	var count int64
	db.Model(&models.SqInspection{}).Where("is_deleted = 0").Count(&count)
	if count > 0 {
		zlog.Info("seed sq business data skipped: inspections already exist", zap.Int64("count", count))
		return
	}
	var admin models.SysUser
	db.Where("username = ? AND is_deleted = 0", "admin").First(&admin)
	var students []models.IdxStudent
	db.Where("is_deleted = 0").Order("id ASC").Limit(8).Find(&students)
	var building models.IdxDormBuilding
	if err := db.Where("is_deleted = 0").Order("id ASC").First(&building).Error; err != nil {
		zlog.Warn("seed sq business data skipped: no building found", zap.Error(err))
		return
	}
	var floor models.IdxDormFloor
	db.Where("building_id = ? AND is_deleted = 0", building.ID).Order("floor_no ASC").First(&floor)
	var rooms []models.IdxDormRoom
	db.Where("building_id = ? AND is_deleted = 0", building.ID).Order("room_no ASC").Limit(2).Find(&rooms)
	if len(students) < 4 || len(rooms) < 2 {
		zlog.Warn("seed sq business data skipped: base data not enough")
		return
	}
	now := time.Now()
	startAt := now.AddDate(0, -2, 0)
	pos := models.SqSelfgovPosition{BizNo: "SQ-POS-2026-0001", StudentID: students[0].ID, ScopeType: "floor", ScopeID: floor.ID, Position: "floor_leader", StartAt: startAt, Status: "formal", AppointedBy: &admin.ID}
	db.Create(&pos)
	score := 86
	inspection := models.SqInspection{BizNo: "SQ-INSP-2026-0001", InspectionType: "hygiene", BuildingID: building.ID, FloorID: &floor.ID, RoomID: &rooms[0].ID, InspectorUserID: admin.ID, InspectedAt: now.AddDate(0, 0, -3), Score: &score, Summary: "整体卫生良好，桌面和地面较整洁，阳台堆放杂物需整改。", Status: "submitted"}
	db.Create(&inspection)
	db.Create(&models.SqInspectionDeduction{InspectionID: inspection.ID, Item: "阳台杂物未及时清理", Deduction: 6})
	closedAt := now.AddDate(0, 0, -1)
	incident := models.SqIncident{BizNo: "SQ-INC-2026-0001", IncidentLevel: "L2", IncidentType: "晚归聚集", OccurredAt: now.AddDate(0, 0, -4), BuildingID: building.ID, FloorID: &floor.ID, RoomID: &rooms[0].ID, LocationDetail: "1号楼101寝室", ReporterUserID: admin.ID, InvolvedStudentIDs: fmt.Sprintf("[%d,%d]", students[1].ID, students[2].ID), InitialAction: "已联系辅导员并进行现场提醒", Status: "closed", ClosedAt: &closedAt, ClosedBy: &admin.ID}
	db.Create(&incident)
	db.Create(&models.SqIncidentAction{IncidentID: incident.ID, ActionText: "楼层长现场核实情况，提醒学生遵守社区作息规定。", ActionAt: now.AddDate(0, 0, -4), ActionBy: admin.ID})
	db.Create(&models.SqIncidentAction{IncidentID: incident.ID, ActionText: "辅导员完成谈话教育，学生提交情况说明，本事件关闭。", ActionAt: closedAt, ActionBy: admin.ID, IsFinal: 1})
	actStart := now.AddDate(0, 0, 10)
	db.Create(&models.SqActivity{BizNo: "SQ-ACT-2026-0001", BuildingID: building.ID, Title: "文明寝室共建日", ActivityType: "community_service", ExpectedParticipants: 60, BudgetCents: 120000, StartedAt: actStart, EndedAt: actStart.Add(2 * time.Hour), Summary: "组织楼栋学生共同维护公共区域卫生。", Status: "S3", CoSignedBy: &admin.ID})
	db.Create(&models.SqAssessment{BizNo: "SQ-ASM-2026-0001", CycleType: "monthly", CycleKey: "2026-03", TargetUserID: admin.ID, TargetPositionID: &pos.ID, ScoreInspection: 88, ScoreIncident: 85, ScoreActivity: 92, ScoreSatisfaction: 90, ScoreBonus: 3, WeightedScore: 89.6, Rating: "good"})
	db.Create(&models.SqLateReturn{StudentID: students[1].ID, OccurredAt: now.AddDate(0, 0, -5), ReportedBy: admin.ID, Reason: "实验室项目调试延时，已补充说明", Semester: "2025-2026-2"})
	db.Create(&models.SqViolation{StudentID: students[2].ID, RoomID: rooms[0].ID, ApplianceName: "违规电热锅", SeizedAt: now.AddDate(0, 0, -6), ReportedBy: admin.ID, Status: "reported_to_college"})
	db.Create(&models.SqVacationStay{BizNo: "SQ-STAY-2026-0001", StudentID: students[3].ID, Semester: "2025-2026-2", StartAt: now.AddDate(0, 1, 0), EndAt: now.AddDate(0, 1, 14), Reason: "参加校级创新创业项目集中训练", Status: "S3", SubmittedAt: now})
	movedAt := now.AddDate(0, 0, -2)
	db.Create(&models.SqRoomChange{BizNo: "SQ-RC-2026-0001", StudentID: students[4].ID, FromRoomID: rooms[0].ID, ToRoomID: rooms[1].ID, Reason: "因参加创新项目团队协作需要调整寝室", CounselorSignedBy: &admin.ID, CouncilSignedBy: &admin.ID, MovedAt: &movedAt, Status: "S3"})

	zlog.Info("seed sq business data completed", zap.Int64("inspection_id", inspection.ID), zap.Int64("incident_id", incident.ID))
}

func seedQGBusinessDetails(db *gorm.DB, zlog *zap.Logger) {
	var count int64
	db.Model(&models.QgAttendance{}).Where("is_deleted = 0").Count(&count)
	if count > 0 {
		zlog.Info("seed qg business details skipped: attendance already exists", zap.Int64("count", count))
		return
	}
	var apply models.QgPositionApply
	if err := db.Where("is_deleted = 0 AND status = ?", "on_job").Order("id ASC").First(&apply).Error; err != nil {
		zlog.Warn("seed qg business details skipped: no on-job apply found", zap.Error(err))
		return
	}
	var position models.QgPosition
	db.Where("id = ?", apply.PositionID).First(&position)
	var admin models.SysUser
	db.Where("username = ? AND is_deleted = 0", "admin").First(&admin)
	now := time.Now()
	var attendances []models.QgAttendance
	for i := 0; i < 5; i++ {
		workDate := time.Date(2026, 3, 3+i*2, 0, 0, 0, 0, time.Local)
		clockIn := time.Date(2026, 3, 3+i*2, 14, 0, 0, 0, time.Local)
		clockOut := clockIn.Add(3 * time.Hour)
		att := models.QgAttendance{BizNo: fmt.Sprintf("QG-ATT-2026-%04d", i+1), ApplyID: apply.ID, StudentID: apply.StudentID, WorkDate: workDate, ClockInAt: &clockIn, ClockOutAt: &clockOut, EffectiveHours: 3, ClockMethod: "card", IP: "127.0.0.1", Geo: "图书馆一楼服务台"}
		if err := db.Create(&att).Error; err == nil {
			attendances = append(attendances, att)
		}
	}
	if len(attendances) == 0 {
		return
	}
	assess := models.QgMonthlyAssess{BizNo: "QG-ASM-2026-0001", ApplyID: apply.ID, StudentID: apply.StudentID, AssessYear: 2026, AssessMonth: 3, ScoreAttendance: 96, ScoreWorkComplete: 92, ScoreComprehensive: 94, WeightedScore: 94.0, Coefficient: 1.0, Status: "S3", Note: "出勤稳定，任务完成质量较高。"}
	db.Create(&assess)
	totalHours := float64(len(attendances)) * 3
	gross := int64(totalHours * float64(position.HourlyRateCents))
	payroll := models.QgPayroll{BizNo: "QG-PAY-2026-0001", StudentID: apply.StudentID, ApplyID: apply.ID, PayYear: 2026, PayMonth: 3, TotalHours: totalHours, GrossCents: gross, NetCents: gross, Coefficient: 1.0, BankAccountLast4Enc: cryptox.Encrypt("1234"), Status: "reviewed", ReviewedBy: &admin.ID}
	db.Create(&payroll)
	for _, att := range attendances {
		amount := int64(att.EffectiveHours * float64(position.HourlyRateCents))
		db.Create(&models.QgPayrollDetail{PayrollID: payroll.ID, AttendanceID: att.ID, WorkDate: att.WorkDate, Hours: att.EffectiveHours, RateCents: position.HourlyRateCents, AmountCents: amount})
	}
	leaveStart := now.AddDate(0, 0, 4)
	db.Create(&models.QgLeave{ApplyID: apply.ID, StudentID: apply.StudentID, StartAt: leaveStart, EndAt: leaveStart.Add(24 * time.Hour), Reason: "参加专业课程集中实训，需请假一天。", Status: "S3"})
	db.Create(&models.QgComplaint{BizNo: "QG-CMP-2026-0001", StudentID: apply.StudentID, TargetType: "payroll", TargetID: payroll.ID, Reason: "对本月薪酬明细中部分工时统计存在疑问，申请复核具体出勤记录和计算过程。", ExpectedReplyDays: 10, Status: "S3", Result: "经复核，薪酬明细与考勤记录一致。", HandledBy: &admin.ID})

	zlog.Info("seed qg business details completed", zap.Int("attendance_count", len(attendances)), zap.Int64("payroll_id", payroll.ID))
}

func seedCmpScores(db *gorm.DB, zlog *zap.Logger) {
	var count int64
	db.Model(&models.CmpScore{}).Where("is_deleted = 0").Count(&count)
	if count > 0 {
		zlog.Info("seed cmp scores skipped: already exist", zap.Int64("count", count))
		return
	}
	var rule models.CmpRuleVersion
	if err := db.Where("is_deleted = 0 AND is_active = 1").Order("id DESC").First(&rule).Error; err != nil {
		zlog.Warn("seed cmp scores skipped: no active rule", zap.Error(err))
		return
	}
	var students []models.IdxStudent
	db.Where("is_deleted = 0").Order("id ASC").Limit(12).Find(&students)
	for i, stu := range students {
		total := 92.5 - float64(i)*1.8
		classRank := i + 1
		collegeRank := i + 3
		score := models.CmpScore{StudentID: stu.ID, AcademicYear: "2025-2026", TotalScore: total, RankInClass: &classRank, RankInCollege: &collegeRank, RuleVersionID: rule.ID, ComputedAt: time.Now()}
		if err := db.Create(&score).Error; err != nil {
			zlog.Warn("seed cmp score failed", zap.Int64("student_id", stu.ID), zap.Error(err))
			continue
		}
		details := []models.CmpScoreDetail{
			{ScoreID: score.ID, Dimension: "league", SubItem: "团内表现", SourceModule: "TY", RawValue: "入团申请与团内活动", Score: 18 - float64(i%3), Weight: 0.25},
			{ScoreID: score.ID, Dimension: "assoc", SubItem: "社团活动", SourceModule: "ST", RawValue: "社团任职与活动参与", Score: 16 - float64(i%4), Weight: 0.20},
			{ScoreID: score.ID, Dimension: "community", SubItem: "社区自治", SourceModule: "SQ", RawValue: "寝室巡查与自治表现", Score: 14 - float64(i%3), Weight: 0.15},
			{ScoreID: score.ID, Dimension: "workstudy", SubItem: "勤工履职", SourceModule: "QG", RawValue: "岗位工时与考核", Score: 13 - float64(i%2), Weight: 0.15},
			{ScoreID: score.ID, Dimension: "academic", SubItem: "学业成绩", SourceModule: "IDX", RawValue: "GPA 与排名", Score: 24 - float64(i%5), Weight: 0.25},
		}
		for _, detail := range details {
			db.Create(&detail)
		}
	}
	zlog.Info("seed cmp scores completed", zap.Int("count", len(students)))
}
