package repository

import (
	"gorm.io/gorm"

	"student-system/internal/models"
)

// StudentRepository 学生数据访问层。
type StudentRepository struct {
	db *gorm.DB
}

// NewStudentRepository 创建学生仓储。
func NewStudentRepository(db *gorm.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

// List 分页查询学生列表，支持按院系/班级/关键字筛选。
func (r *StudentRepository) List(collegeID, classID int64, keyword string, page, pageSize int) ([]models.IdxStudent, int64, error) {
	query := r.db.Where("is_deleted = 0")

	if collegeID > 0 {
		query = query.Where("college_id = ?", collegeID)
	}
	if classID > 0 {
		query = query.Where("class_id = ?", classID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR student_no LIKE ?", like, like)
	}

	var total int64
	if err := query.Model(&models.IdxStudent{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var students []models.IdxStudent
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&students).Error; err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

// GetByID 按 ID 查询单个学生，Preload 关联组织。
func (r *StudentRepository) GetByID(id int64) (*models.IdxStudent, error) {
	var student models.IdxStudent
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&student).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

// GetByStudentNo 按学号查询学生。
func (r *StudentRepository) GetByStudentNo(studentNo string) (*models.IdxStudent, error) {
	var student models.IdxStudent
	if err := r.db.Where("student_no = ? AND is_deleted = 0", studentNo).First(&student).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

// Create 创建学生。
func (r *StudentRepository) Create(student *models.IdxStudent) error {
	db := r.db
	if student.IDCardHash == "" {
		db = db.Omit("id_card_enc", "id_card_hash")
	}
	if student.PhoneHash == "" {
		db = db.Omit("phone_enc", "phone_hash")
	}
	return db.Create(student).Error
}

// ExistsByIDCardHashExceptID 判断身份证哈希是否被其他学生占用。
func (r *StudentRepository) ExistsByIDCardHashExceptID(hash string, id int64) (bool, error) {
	var count int64
	if err := r.db.Model(&models.IdxStudent{}).
		Where("id_card_hash = ? AND id <> ? AND is_deleted = 0", hash, id).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// Update 更新学生。
func (r *StudentRepository) Update(student *models.IdxStudent) error {
	db := r.db
	if student.IDCardHash == "" {
		db = db.Omit("id_card_enc", "id_card_hash")
	}
	if student.PhoneHash == "" {
		db = db.Omit("phone_enc", "phone_hash")
	}
	return db.Save(student).Error
}

// UpdateStudentUserDisplayName 同步学生账号显示名。
func (r *StudentRepository) UpdateStudentUserDisplayName(studentID int64, name string) error {
	return r.db.Model(&models.SysUser{}).
		Where("student_id = ? AND is_deleted = 0", studentID).
		Update("display_name", name).Error
}

// SoftDelete 软删除学生。
func (r *StudentRepository) SoftDelete(id int64) error {
	return r.db.Model(&models.IdxStudent{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// GetByUserID 通过 sys_user.student_id 反查学生。
func (r *StudentRepository) GetByUserID(userID int64) (*models.IdxStudent, error) {
	var user models.SysUser
	if err := r.db.Where("id = ? AND is_deleted = 0", userID).First(&user).Error; err != nil {
		return nil, err
	}
	if user.StudentID == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.GetByID(*user.StudentID)
}

// BatchCreate 批量创建学生（CSV 导入用）。
func (r *StudentRepository) BatchCreate(students []models.IdxStudent) error {
	return r.db.CreateInBatches(students, 50).Error
}

// ---- 组织树查询 ----

// ListColleges 查询所有院系。
func (r *StudentRepository) ListColleges() ([]models.SysCollege, error) {
	var colleges []models.SysCollege
	if err := r.db.Where("is_deleted = 0").Order("id ASC").Find(&colleges).Error; err != nil {
		return nil, err
	}
	return colleges, nil
}

// ListMajorsByCollege 按院系查询专业。
func (r *StudentRepository) ListMajorsByCollege(collegeID int64) ([]models.SysMajor, error) {
	var majors []models.SysMajor
	query := r.db.Where("is_deleted = 0")
	if collegeID > 0 {
		query = query.Where("college_id = ?", collegeID)
	}
	if err := query.Order("id ASC").Find(&majors).Error; err != nil {
		return nil, err
	}
	return majors, nil
}

// ListClassesByMajor 按专业查询班级。
func (r *StudentRepository) ListClassesByMajor(majorID int64) ([]models.IdxClass, error) {
	var classes []models.IdxClass
	query := r.db.Where("is_deleted = 0")
	if majorID > 0 {
		query = query.Where("major_id = ?", majorID)
	}
	if err := query.Order("id ASC").Find(&classes).Error; err != nil {
		return nil, err
	}
	return classes, nil
}

// GetCollegeByID 按ID查询院系。
func (r *StudentRepository) GetCollegeByID(id int64) (*models.SysCollege, error) {
	var college models.SysCollege
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&college).Error; err != nil {
		return nil, err
	}
	return &college, nil
}

// GetMajorByID 按ID查询专业。
func (r *StudentRepository) GetMajorByID(id int64) (*models.SysMajor, error) {
	var major models.SysMajor
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&major).Error; err != nil {
		return nil, err
	}
	return &major, nil
}

// GetClassByID 按ID查询班级。
func (r *StudentRepository) GetClassByID(id int64) (*models.IdxClass, error) {
	var class models.IdxClass
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&class).Error; err != nil {
		return nil, err
	}
	return &class, nil
}

// GetDB 返回底层数据库连接（用于跨模块操作，如自动创建团员花名册）。
func (r *StudentRepository) GetDB() *gorm.DB {
	return r.db
}
