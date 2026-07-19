package service

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"gorm.io/gorm"

	"student-system/internal/models"
	"student-system/internal/modules/idx/repository"
	"student-system/pkg/cryptox"
)

// utf8BOM UTF-8 编码的 BOM 标记。
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// decodeCSVReader 将上传的 CSV 字节流统一转为 UTF-8。
// 兼容三种常见来源（解决中文/标点显示为 "?" 的根因）：
//  1. UTF-8（含/不含 BOM）：直接使用，剥离 BOM；
//  2. GB18030 / GBK（Windows 下 Excel "另存为 CSV" 默认编码）：转 UTF-8；
//  3. 其他不可识别编码：回退按 UTF-8 处理（由上层报错）。
func decodeCSVReader(r io.Reader) (io.Reader, error) {
	br := bufio.NewReader(r)
	// 预读 BOM
	head, err := br.Peek(3)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("读取文件头失败: %w", err)
	}
	if len(head) >= 3 && bytes.Equal(head[:3], utf8BOM) {
		_, _ = br.Discard(3)
		return br, nil
	}
	// 读取全部内容用于编码探测（学生导入文件通常很小，可接受）
	all, err := io.ReadAll(br)
	if err != nil {
		return nil, fmt.Errorf("读取文件内容失败: %w", err)
	}
	if utf8.Valid(all) {
		return bytes.NewReader(all), nil
	}
	// 视为 GB18030（向下兼容 GBK），转换为 UTF-8
	utf8Bytes, _, convErr := transform.Bytes(simplifiedchinese.GB18030.NewDecoder(), all)
	if convErr != nil {
		return nil, fmt.Errorf("CSV 文件编码无法识别（仅支持 UTF-8 或 GBK/GB18030）: %w", convErr)
	}
	return bytes.NewReader(utf8Bytes), nil
}

// StudentService 学生业务服务层。
type StudentService struct {
	repo *repository.StudentRepository
}

// NewStudentService 创建学生服务。
func NewStudentService(repo *repository.StudentRepository) *StudentService {
	return &StudentService{repo: repo}
}

// StudentListResult 学生列表结果。
type StudentListResult struct {
	Items    []StudentView `json:"items"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// StudentView 学生视图（含脱敏字段）。
type StudentView struct {
	ID              int64  `json:"id"`
	StudentNo       string `json:"student_no"`
	Name            string `json:"name"`
	IDCardMasked    string `json:"id_card_masked"`
	Gender          string `json:"gender"`
	BirthDate       string `json:"birth_date,omitempty"`
	Ethnicity       string `json:"ethnicity"`
	PoliticalStatus string `json:"political_status"`
	JoinAt          string `json:"join_at,omitempty"`
	MemberCardNo    string `json:"member_card_no"`
	CollegeID       *int64 `json:"college_id,omitempty"`
	CollegeName     string `json:"college_name,omitempty"`
	MajorID         *int64 `json:"major_id,omitempty"`
	MajorName       string `json:"major_name,omitempty"`
	ClassID         *int64 `json:"class_id,omitempty"`
	ClassName       string `json:"class_name,omitempty"`
	Grade           *int   `json:"grade,omitempty"`
	PhoneMasked     string `json:"phone_masked"`
	Email           string `json:"email"`
	EnrollmentAt    string `json:"enrollment_at,omitempty"`
	Status          string `json:"status"`
	IsDifficulty    int    `json:"is_difficulty"`
	DifficultyLevel string `json:"difficulty_level"`
	CreatedAt       string `json:"created_at"`
}

// toView 将模型转为视图（脱敏处理）。
func toView(s models.IdxStudent, collegeName, majorName, className string) StudentView {
	v := StudentView{
		ID:              s.ID,
		StudentNo:       s.StudentNo,
		Name:            s.Name,
		Gender:          s.Gender,
		Ethnicity:       s.Ethnicity,
		PoliticalStatus: s.PoliticalStatus,
		MemberCardNo:    s.MemberCardNo,
		CollegeID:       s.CollegeID,
		CollegeName:     collegeName,
		MajorID:         s.MajorID,
		MajorName:       majorName,
		ClassID:         s.ClassID,
		ClassName:       className,
		Grade:           s.Grade,
		Email:           s.Email,
		Status:          s.Status,
		IsDifficulty:    s.IsDifficulty,
		DifficultyLevel: s.DifficultyLevel,
	}

	// 脱敏处理
	if s.IDCardEnc != "" {
		plain, err := cryptox.Decrypt(s.IDCardEnc)
		if err == nil {
			v.IDCardMasked = cryptox.MaskIDCard(plain)
		}
	}
	if s.PhoneEnc != "" {
		plain, err := cryptox.Decrypt(s.PhoneEnc)
		if err == nil {
			v.PhoneMasked = cryptox.MaskPhone(plain)
		}
	}

	if s.BirthDate != nil {
		v.BirthDate = s.BirthDate.Format("2006-01-02")
	}
	if s.JoinAt != nil {
		v.JoinAt = s.JoinAt.Format("2006-01-02")
	}
	if s.EnrollmentAt != nil {
		v.EnrollmentAt = s.EnrollmentAt.Format("2006-01-02")
	}
	v.CreatedAt = s.CreatedAt.Format("2006-01-02T15:04:05+08:00")

	return v
}

// List 分页查询学生列表。
func (s *StudentService) List(collegeID, classID int64, keyword string, page, pageSize int) (*StudentListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	students, total, err := s.repo.List(collegeID, classID, keyword, page, pageSize)
	if err != nil {
		return nil, err
	}

	// 预加载组织名称
	colleges, _ := s.repo.ListColleges()
	collegeMap := make(map[int64]string)
	for _, c := range colleges {
		collegeMap[c.ID] = c.Name
	}

	majors, _ := s.repo.ListMajorsByCollege(0)
	majorMap := make(map[int64]string)
	for _, m := range majors {
		majorMap[m.ID] = m.Name
	}

	classes, _ := s.repo.ListClassesByMajor(0)
	classMap := make(map[int64]string)
	for _, c := range classes {
		classMap[c.ID] = c.Name
	}

	items := make([]StudentView, 0, len(students))
	for _, stu := range students {
		cName := ""
		if stu.CollegeID != nil {
			cName = collegeMap[*stu.CollegeID]
		}
		mName := ""
		if stu.MajorID != nil {
			mName = majorMap[*stu.MajorID]
		}
		clName := ""
		if stu.ClassID != nil {
			clName = classMap[*stu.ClassID]
		}
		items = append(items, toView(stu, cName, mName, clName))
	}

	return &StudentListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Get 获取学生详情（脱敏）。
func (s *StudentService) Get(id int64) (*StudentView, error) {
	student, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	cName, mName, clName := s.loadOrgNames(student)
	v := toView(*student, cName, mName, clName)
	return &v, nil
}

// CreateStudentRequest 创建学生请求。
type CreateStudentRequest struct {
	StudentNo       string `json:"student_no" binding:"required"`
	Name            string `json:"name" binding:"required"`
	IDCard          string `json:"id_card"`
	Gender          string `json:"gender"`
	BirthDate       string `json:"birth_date"`
	Ethnicity       string `json:"ethnicity"`
	PoliticalStatus string `json:"political_status"`
	JoinAt          string `json:"join_at"`
	MemberCardNo    string `json:"member_card_no"`
	CollegeID       *int64 `json:"college_id"`
	MajorID         *int64 `json:"major_id"`
	ClassID         *int64 `json:"class_id"`
	Grade           *int   `json:"grade"`
	Phone           string `json:"phone"`
	Email           string `json:"email"`
	EnrollmentAt    string `json:"enrollment_at"`
}

// Create 创建学生。
func (s *StudentService) Create(req *CreateStudentRequest) (*StudentView, error) {
	// 检查学号唯一
	existing, err := s.repo.GetByStudentNo(req.StudentNo)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("学号已存在: %s", req.StudentNo)
	}

	student := models.IdxStudent{
		StudentNo:       req.StudentNo,
		Name:            req.Name,
		Gender:          req.Gender,
		Ethnicity:       req.Ethnicity,
		PoliticalStatus: req.PoliticalStatus,
		MemberCardNo:    req.MemberCardNo,
		CollegeID:       req.CollegeID,
		MajorID:         req.MajorID,
		ClassID:         req.ClassID,
		Grade:           req.Grade,
		Email:           req.Email,
		Status:          "enrolled",
	}

	// 入团时间
	if req.JoinAt != "" {
		student.JoinAt = parseDate(req.JoinAt)
	}

	// 加密身份证
	if req.IDCard != "" {
		student.IDCardEnc = cryptox.Encrypt(req.IDCard)
		// SHA-256 哈希用于去重索引
		hash := sha256.Sum256([]byte(req.IDCard))
		student.IDCardHash = fmt.Sprintf("%x", hash)
		// 从身份证提取出生日期
		if req.BirthDate == "" && len(req.IDCard) >= 14 {
			student.BirthDate = parseDate(req.IDCard[6:10] + "-" + req.IDCard[10:12] + "-" + req.IDCard[12:14])
		}
	}

	if req.BirthDate != "" && student.BirthDate == nil {
		student.BirthDate = parseDate(req.BirthDate)
	}

	// 加密手机号
	if req.Phone != "" {
		student.PhoneEnc = cryptox.Encrypt(req.Phone)
		hash := sha256.Sum256([]byte(req.Phone))
		student.PhoneHash = fmt.Sprintf("%x", hash)
	}

	if req.EnrollmentAt != "" {
		student.EnrollmentAt = parseDate(req.EnrollmentAt)
	}

	if err := s.repo.Create(&student); err != nil {
		return nil, err
	}

	// 政治面貌为共青团员时，自动创建团员花名册
	if req.PoliticalStatus == "member" && req.JoinAt != "" && student.CollegeID != nil {
		s.autoCreateRoster(&student)
	}

	return s.Get(student.ID)
}

// autoCreateRoster 为中学入团的学生自动创建团员花名册记录。
func (s *StudentService) autoCreateRoster(student *models.IdxStudent) {
	// 查找该学生所属团支部
	var branch models.TyBranch
	if err := s.repo.GetDB().Where("college_id = ? AND is_deleted = 0", *student.CollegeID).First(&branch).Error; err != nil {
		return
	}

	// 检查是否已有花名册记录
	var count int64
	s.repo.GetDB().Model(&models.TyMemberRoster{}).Where("student_id = ? AND is_deleted = 0", student.ID).Count(&count)
	if count > 0 {
		return
	}

	rosterBizNo := fmt.Sprintf("TY-ROSTER-%d-%04d", student.JoinAt.Year(), student.ID)
	roster := models.TyMemberRoster{
		BizNo:         rosterBizNo,
		StudentID:     student.ID,
		BranchID:      branch.ID,
		JoinAt:        *student.JoinAt,
		Status:        "active",
	}
	if err := s.repo.GetDB().Create(&roster).Error; err != nil {
		return
	}
}

// UpdateStudentRequest 更新学生请求。
type UpdateStudentRequest struct {
	Name            *string `json:"name"`
	IDCard          *string `json:"id_card"`
	Gender          *string `json:"gender"`
	BirthDate       *string `json:"birth_date"`
	Ethnicity       *string `json:"ethnicity"`
	PoliticalStatus *string `json:"political_status"`
	CollegeID       *int64  `json:"college_id"`
	MajorID         *int64  `json:"major_id"`
	ClassID         *int64  `json:"class_id"`
	Grade           *int    `json:"grade"`
	Phone           *string `json:"phone"`
	Email           *string `json:"email"`
	Status          *string `json:"status"`
}

// Update 更新学生信息。
func (s *StudentService) Update(id int64, req *UpdateStudentRequest) (*StudentView, error) {
	student, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		student.Name = *req.Name
	}
	if req.Gender != nil {
		student.Gender = *req.Gender
	}
	if req.Ethnicity != nil {
		student.Ethnicity = *req.Ethnicity
	}
	if req.PoliticalStatus != nil {
		student.PoliticalStatus = *req.PoliticalStatus
	}
	if req.CollegeID != nil {
		student.CollegeID = req.CollegeID
	}
	if req.MajorID != nil {
		student.MajorID = req.MajorID
	}
	if req.ClassID != nil {
		student.ClassID = req.ClassID
	}
	if req.Grade != nil {
		student.Grade = req.Grade
	}
	if req.Email != nil {
		student.Email = *req.Email
	}
	if req.Status != nil {
		student.Status = *req.Status
	}
	if req.IDCard != nil && *req.IDCard != "" && !containsMask(*req.IDCard) {
		hash := sha256.Sum256([]byte(*req.IDCard))
		idCardHash := fmt.Sprintf("%x", hash)
		exists, existsErr := s.repo.ExistsByIDCardHashExceptID(idCardHash, student.ID)
		if existsErr != nil {
			return nil, existsErr
		}
		if exists {
			return nil, fmt.Errorf("身份证号已被其他学生使用")
		}
		student.IDCardEnc = cryptox.Encrypt(*req.IDCard)
		student.IDCardHash = idCardHash
	}
	if req.Phone != nil && *req.Phone != "" && !containsMask(*req.Phone) {
		student.PhoneEnc = cryptox.Encrypt(*req.Phone)
		hash := sha256.Sum256([]byte(*req.Phone))
		student.PhoneHash = fmt.Sprintf("%x", hash)
	}
	if req.BirthDate != nil {
		student.BirthDate = parseDate(*req.BirthDate)
	}

	if err := s.repo.Update(student); err != nil {
		return nil, err
	}
	if req.Name != nil {
		if err := s.repo.UpdateStudentUserDisplayName(student.ID, student.Name); err != nil {
			return nil, err
		}
	}

	return s.Get(student.ID)
}

func containsMask(value string) bool {
	return strings.Contains(value, "*")
}

// SoftDelete 软删除学生。
func (s *StudentService) SoftDelete(id int64) error {
	return s.repo.SoftDelete(id)
}

// GetProfileByUserID 通过用户ID获取学生画像。
func (s *StudentService) GetProfileByUserID(userID int64) (*StudentView, error) {
	student, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	cName, mName, clName := s.loadOrgNames(student)
	v := toView(*student, cName, mName, clName)
	return &v, nil
}

// ImportResult 导入结果。
type ImportResult struct {
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors,omitempty"`
}

// ImportCSV 从 CSV Reader 批量导入学生。
// CSV 格式：学号,姓名,性别,身份证号,手机号,院系ID,专业ID,班级ID,年级,邮箱
// 自动识别 UTF-8（含 BOM）/ GBK / GB18030 三种常见编码，统一转为 UTF-8 后再解析，
// 避免中文/全角标点写入 SQLite 后变成 "?"。
func (s *StudentService) ImportCSV(reader io.Reader) (*ImportResult, error) {
	utf8Reader, err := decodeCSVReader(reader)
	if err != nil {
		return nil, err
	}
	csvReader := csv.NewReader(utf8Reader)
	// 跳过标题行
	if _, err := csvReader.Read(); err != nil {
		return nil, fmt.Errorf("读取 CSV 标题行失败: %w", err)
	}

	result := &ImportResult{}
	var students []models.IdxStudent

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("行读取错误: %v", err))
			continue
		}

		if len(record) < 4 {
			result.Failed++
			result.Errors = append(result.Errors, "字段不足，至少需要学号、姓名、性别、身份证号")
			continue
		}

		student := models.IdxStudent{
			StudentNo: record[0],
			Name:      record[1],
			Gender:    record[2],
			Status:    "enrolled",
		}

		// 身份证加密
		if record[3] != "" {
			student.IDCardEnc = cryptox.Encrypt(record[3])
			hash := sha256.Sum256([]byte(record[3]))
			student.IDCardHash = fmt.Sprintf("%x", hash)
		}

		// 手机号加密
		if len(record) > 4 && record[4] != "" {
			student.PhoneEnc = cryptox.Encrypt(record[4])
			hash := sha256.Sum256([]byte(record[4]))
			student.PhoneHash = fmt.Sprintf("%x", hash)
		}

		// 院系ID
		if len(record) > 5 && record[5] != "" {
			if id, err := strconv.ParseInt(record[5], 10, 64); err == nil {
				student.CollegeID = &id
			}
		}

		// 专业ID
		if len(record) > 6 && record[6] != "" {
			if id, err := strconv.ParseInt(record[6], 10, 64); err == nil {
				student.MajorID = &id
			}
		}

		// 班级ID
		if len(record) > 7 && record[7] != "" {
			if id, err := strconv.ParseInt(record[7], 10, 64); err == nil {
				student.ClassID = &id
			}
		}

		// 年级
		if len(record) > 8 && record[8] != "" {
			if g, err := strconv.Atoi(record[8]); err == nil {
				student.Grade = &g
			}
		}

		// 邮箱
		if len(record) > 9 && record[9] != "" {
			student.Email = record[9]
		}

		students = append(students, student)
	}

	if len(students) > 0 {
		if err := s.repo.BatchCreate(students); err != nil {
			return nil, err
		}
		result.Success = len(students)
	}

	return result, nil
}

// loadOrgNames 加载组织名称。
func (s *StudentService) loadOrgNames(student *models.IdxStudent) (collegeName, majorName, className string) {
	if student.CollegeID != nil {
		if c, err := s.repo.GetCollegeByID(*student.CollegeID); err == nil {
			collegeName = c.Name
		}
	}
	if student.MajorID != nil {
		if m, err := s.repo.GetMajorByID(*student.MajorID); err == nil {
			majorName = m.Name
		}
	}
	if student.ClassID != nil {
		if c, err := s.repo.GetClassByID(*student.ClassID); err == nil {
			className = c.Name
		}
	}
	return
}

// parseDate 解析日期字符串。
func parseDate(s string) *time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}

// OrgTreeNode 组织树节点。
type OrgTreeNode struct {
	ID        int64          `json:"id"`
	UniqueKey string         `json:"unique_key"` // 全局唯一键（type_id），解决 el-tree node-key 冲突
	Label     string         `json:"label"`
	Type      string         `json:"type"` // college / major / class
	ParentID  *int64         `json:"parent_id,omitempty"`
	Children  []*OrgTreeNode `json:"children,omitempty"`
}

// BuildOrgTree 构建组织树（院系→专业→班级）。
func (s *StudentService) BuildOrgTree() ([]*OrgTreeNode, error) {
	colleges, err := s.repo.ListColleges()
	if err != nil {
		return nil, err
	}

	majors, err := s.repo.ListMajorsByCollege(0)
	if err != nil {
		return nil, err
	}

	classes, err := s.repo.ListClassesByMajor(0)
	if err != nil {
		return nil, err
	}

	// 按院系分组专业
	majorByCollege := make(map[int64][]models.SysMajor)
	for _, m := range majors {
		majorByCollege[m.CollegeID] = append(majorByCollege[m.CollegeID], m)
	}

	// 按专业分组班级
	classByMajor := make(map[int64][]models.IdxClass)
	for _, c := range classes {
		classByMajor[c.MajorID] = append(classByMajor[c.MajorID], c)
	}

	// 构建树
	var roots []*OrgTreeNode
	for _, college := range colleges {
		cNode := &OrgTreeNode{
			ID:        college.ID,
			UniqueKey: "college_" + strconv.FormatInt(college.ID, 10),
			Label:     college.Name,
			Type:      "college",
		}

		for _, major := range majorByCollege[college.ID] {
			mNode := &OrgTreeNode{
				ID:        major.ID,
				UniqueKey: "major_" + strconv.FormatInt(major.ID, 10),
				Label:     major.Name,
				Type:      "major",
				ParentID:  &college.ID,
			}

			for _, class := range classByMajor[major.ID] {
				clNode := &OrgTreeNode{
					ID:        class.ID,
					UniqueKey: "class_" + strconv.FormatInt(class.ID, 10),
					Label:     class.Name,
					Type:      "class",
					ParentID:  &major.ID,
				}
				mNode.Children = append(mNode.Children, clNode)
			}

			cNode.Children = append(cNode.Children, mNode)
		}

		roots = append(roots, cNode)
	}

	return roots, nil
}

// Ensure interface compliance.
var _ = gorm.ErrRecordNotFound
