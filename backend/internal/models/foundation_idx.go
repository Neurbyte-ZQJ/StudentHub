package models

import "time"

// SysCollege 院系。docs/03 §4.1.1。
type SysCollege struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Code      string    `gorm:"column:code;type:text;not null;uniqueIndex:uniq_sys_college_code" json:"code"`
	Name      string    `gorm:"column:name;type:text;not null;index:idx_sys_college_name" json:"name"`
	NameEn    string    `gorm:"column:name_en;type:text" json:"name_en"`
	IsDeleted int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SysCollege) TableName() string { return "sys_college" }

// SysMajor 专业。docs/03 §4.1.2。
type SysMajor struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CollegeID int64     `gorm:"column:college_id;not null;index:idx_sys_major_college;uniqueIndex:uniq_sys_major_college_code,priority:1" json:"college_id"`
	Code      string    `gorm:"column:code;type:text;not null;uniqueIndex:uniq_sys_major_college_code,priority:2" json:"code"`
	Name      string    `gorm:"column:name;type:text;not null" json:"name"`
	IsDeleted int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (SysMajor) TableName() string { return "sys_major" }

// IdxClass 行政班 / 团支部对应行政班。docs/03 §4.1.3。
type IdxClass struct {
	ID           int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	MajorID      int64     `gorm:"column:major_id;not null;index;uniqueIndex:uniq_idx_class_major_code,priority:1" json:"major_id"`
	Grade        int       `gorm:"column:grade;not null" json:"grade"`
	Code         string    `gorm:"column:code;type:text;not null;uniqueIndex:uniq_idx_class_major_code,priority:2" json:"code"`
	Name         string    `gorm:"column:name;type:text;not null" json:"name"`
	CounselorID  *int64    `gorm:"column:counselor_id" json:"counselor_id,omitempty"`
	IsDeleted    int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (IdxClass) TableName() string { return "idx_class" }

// IdxStudent 学生主体。docs/03 §4.1.4。
type IdxStudent struct {
	ID              int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	StudentNo       string     `gorm:"column:student_no;type:text;not null;uniqueIndex:uniq_idx_student_no" json:"student_no"`
	Name            string     `gorm:"column:name;type:text;not null;index:idx_idx_student_name" json:"name"`
	IDCardEnc       string     `gorm:"column:id_card_enc;type:text" json:"-"`
	IDCardHash      string     `gorm:"column:id_card_hash;type:text;uniqueIndex:uniq_idx_student_id_card_hash" json:"-"`
	Gender          string     `gorm:"column:gender;type:text;check:gender IN ('M','F','U')" json:"gender"`
	BirthDate       *time.Time `gorm:"column:birth_date;type:date" json:"birth_date,omitempty"`
	Ethnicity       string     `gorm:"column:ethnicity;type:text" json:"ethnicity"`
	PoliticalStatus string     `gorm:"column:political_status;type:text;index:idx_idx_student_political_status" json:"political_status"`
	JoinAt          *time.Time `gorm:"column:join_at;type:date" json:"join_at,omitempty"`
	MemberCardNo    string     `gorm:"column:member_card_no;type:text" json:"member_card_no"`
	CollegeID       *int64     `gorm:"column:college_id;index:idx_idx_student_college_id" json:"college_id,omitempty"`
	MajorID         *int64     `gorm:"column:major_id" json:"major_id,omitempty"`
	ClassID         *int64     `gorm:"column:class_id;index:idx_idx_student_class_id" json:"class_id,omitempty"`
	Grade           *int       `gorm:"column:grade" json:"grade,omitempty"`
	PhoneEnc        string     `gorm:"column:phone_enc;type:text" json:"-"`
	PhoneHash       string     `gorm:"column:phone_hash;type:text" json:"-"`
	Email           string     `gorm:"column:email;type:text" json:"email"`
	EnrollmentAt    *time.Time `gorm:"column:enrollment_at;type:date" json:"enrollment_at,omitempty"`
	GraduationAt    *time.Time `gorm:"column:graduation_at;type:date" json:"graduation_at,omitempty"`
	Status          string     `gorm:"column:status;type:text;not null;default:enrolled;check:status IN ('enrolled','suspended','graduated','withdrawn')" json:"status"`
	IsDifficulty    int        `gorm:"column:is_difficulty;not null;default:0" json:"is_difficulty"`
	DifficultyLevel string     `gorm:"column:difficulty_level;type:text" json:"difficulty_level"`
	IsDeleted       int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (IdxStudent) TableName() string { return "idx_student" }

// IdxDormBuilding 楼栋。docs/03 §4.1.5。
type IdxDormBuilding struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Code        string    `gorm:"column:code;type:text;not null;uniqueIndex:uniq_idx_dorm_building_code" json:"code"`
	Name        string    `gorm:"column:name;type:text;not null" json:"name"`
	FloorCount  int       `gorm:"column:floor_count;not null;default:0" json:"floor_count"`
	TutorUserID *int64    `gorm:"column:tutor_user_id" json:"tutor_user_id,omitempty"`
	IsDeleted   int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (IdxDormBuilding) TableName() string { return "idx_dorm_building" }

// IdxDormFloor 楼层。docs/03 §4.1.5。
type IdxDormFloor struct {
	ID                   int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BuildingID           int64     `gorm:"column:building_id;not null;uniqueIndex:uniq_idx_dorm_floor_building_no,priority:1" json:"building_id"`
	FloorNo              int       `gorm:"column:floor_no;not null;uniqueIndex:uniq_idx_dorm_floor_building_no,priority:2" json:"floor_no"`
	FloorLeaderStudentID *int64    `gorm:"column:floor_leader_student_id" json:"floor_leader_student_id,omitempty"`
	IsDeleted            int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt            time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (IdxDormFloor) TableName() string { return "idx_dorm_floor" }

// IdxDormRoom 寝室。docs/03 §4.1.5。
type IdxDormRoom struct {
	ID                  int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BuildingID          int64     `gorm:"column:building_id;not null;uniqueIndex:uniq_idx_dorm_room_bno_no,priority:1" json:"building_id"`
	FloorID             int64     `gorm:"column:floor_id;not null" json:"floor_id"`
	RoomNo              string    `gorm:"column:room_no;type:text;not null;uniqueIndex:uniq_idx_dorm_room_bno_no,priority:2" json:"room_no"`
	BedCount            int       `gorm:"column:bed_count;not null;default:4" json:"bed_count"`
	RoomLeaderStudentID *int64    `gorm:"column:room_leader_student_id" json:"room_leader_student_id,omitempty"`
	IsDeleted           int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt           time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (IdxDormRoom) TableName() string { return "idx_dorm_room" }

// IdxDormBed 床位。docs/03 §4.1.5。
type IdxDormBed struct {
	ID                int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	RoomID            int64      `gorm:"column:room_id;not null;uniqueIndex:uniq_idx_dorm_bed_room_no,priority:1" json:"room_id"`
	BedNo             string     `gorm:"column:bed_no;type:text;not null;uniqueIndex:uniq_idx_dorm_bed_room_no,priority:2" json:"bed_no"`
	OccupantStudentID *int64     `gorm:"column:occupant_student_id" json:"occupant_student_id,omitempty"`
	MoveInAt          *time.Time `gorm:"column:move_in_at;type:date" json:"move_in_at,omitempty"`
	MoveOutAt         *time.Time `gorm:"column:move_out_at;type:date" json:"move_out_at,omitempty"`
	IsDeleted         int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (IdxDormBed) TableName() string { return "idx_dorm_bed" }
