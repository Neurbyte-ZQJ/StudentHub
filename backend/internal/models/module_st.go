package models

import "time"

// StAssociation 社团主数据。docs/03 §6.2.1。
type StAssociation struct {
	ID                  int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo               string     `gorm:"column:biz_no;type:text;uniqueIndex:uniq_st_association_biz_no" json:"biz_no"`
	Name                string     `gorm:"column:name;type:text;not null" json:"name"`
	CollegeID           int64      `gorm:"column:college_id;not null;index:idx_st_assoc_college,priority:1" json:"college_id"`
	TutorUserID         *int64     `gorm:"column:tutor_user_id" json:"tutor_user_id,omitempty"`
	PresidentStudentID  *int64     `gorm:"column:president_student_id" json:"president_student_id,omitempty"`
	BusinessScope       string     `gorm:"column:business_scope;type:text;not null" json:"business_scope"`
	Status              string     `gorm:"column:status;type:text;not null;default:preparing;check:status IN ('preparing','trial','registered','rectifying','cancelled');index:idx_st_assoc_college,priority:2" json:"status"`
	TrialStartedAt      *time.Time `gorm:"column:trial_started_at;type:date" json:"trial_started_at,omitempty"`
	RegisteredAt        *time.Time `gorm:"column:registered_at;type:date" json:"registered_at,omitempty"`
	StarRating          *int       `gorm:"column:star_rating;check:star_rating IS NULL OR (star_rating BETWEEN 1 AND 5)" json:"star_rating,omitempty"`
	FoundedAt           *time.Time `gorm:"column:founded_at;type:date" json:"founded_at,omitempty"`
	IsDeleted           int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt           time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy           *int64     `gorm:"column:created_by" json:"created_by,omitempty"`
	UpdatedBy           *int64     `gorm:"column:updated_by" json:"updated_by,omitempty"`
}

func (StAssociation) TableName() string { return "st_association" }

// StCharter 社团章程。docs/03 §6.2.2。
type StCharter struct {
	ID            int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	AssociationID int64     `gorm:"column:association_id;not null;uniqueIndex:uniq_st_charter_assoc_ver,priority:1" json:"association_id"`
	Version       int       `gorm:"column:version;not null;default:1;uniqueIndex:uniq_st_charter_assoc_ver,priority:2" json:"version"`
	ChapterCount  int       `gorm:"column:chapter_count;not null;check:chapter_count >= 10" json:"chapter_count"`
	FileID        int64     `gorm:"column:file_id;not null" json:"file_id"`
	EffectiveAt   time.Time `gorm:"column:effective_at;type:date;not null" json:"effective_at"`
	IsCurrent     int       `gorm:"column:is_current;not null;default:1" json:"is_current"`
	IsDeleted     int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (StCharter) TableName() string { return "st_charter" }

// StFounder 社团发起人。docs/03 §6.2.3。
type StFounder struct {
	ID            int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	AssociationID int64     `gorm:"column:association_id;not null;uniqueIndex:uniq_st_founder_assoc_stu,priority:1" json:"association_id"`
	StudentID     int64     `gorm:"column:student_id;not null;uniqueIndex:uniq_st_founder_assoc_stu,priority:2" json:"student_id"`
	JoinedAt      time.Time `gorm:"column:joined_at;not null;default:CURRENT_TIMESTAMP" json:"joined_at"`
	IsDeleted     int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
}

func (StFounder) TableName() string { return "st_founder" }

// StAssocMember 社团成员。docs/03 §6.2.4。
type StAssocMember struct {
	ID            int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	AssociationID int64      `gorm:"column:association_id;not null;index:idx_st_member_assoc_role,priority:1" json:"association_id"`
	StudentID     int64      `gorm:"column:student_id;not null;index" json:"student_id"`
	Role          string     `gorm:"column:role;type:text;not null;default:member;check:role IN ('president','vice_president','director','member');index:idx_st_member_assoc_role,priority:2" json:"role"`
	JoinedAt      time.Time  `gorm:"column:joined_at;type:date;not null" json:"joined_at"`
	LeftAt        *time.Time `gorm:"column:left_at;type:date;index:idx_st_member_assoc_role,priority:3" json:"left_at,omitempty"`
	IsCoreOfficer int        `gorm:"column:is_core_officer;not null;default:0" json:"is_core_officer"`
	IsDeleted     int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (StAssocMember) TableName() string { return "st_assoc_member" }

// StRecruitPlan 招新计划。docs/03 §6.2.5。
type StRecruitPlan struct {
	ID               int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo            string     `gorm:"column:biz_no;type:text" json:"biz_no"`
	AssociationID    int64      `gorm:"column:association_id;not null;index" json:"association_id"`
	Season           string     `gorm:"column:season;type:text;not null;check:season IN ('autumn','spring')" json:"season"`
	AcademicYear     string     `gorm:"column:academic_year;type:text;not null" json:"academic_year"`
	TargetCount      int        `gorm:"column:target_count;not null;check:target_count > 0" json:"target_count"`
	PlanFileID       *int64     `gorm:"column:plan_file_id" json:"plan_file_id,omitempty"`
	AssessmentMethod string     `gorm:"column:assessment_method;type:text" json:"assessment_method"`
	InterviewAt      *time.Time `gorm:"column:interview_at" json:"interview_at,omitempty"`
	Status           string     `gorm:"column:status;type:text;not null;default:S0;check:status IN ('S0','S1','S3','S4')" json:"status"`
	ResultDeadline   *time.Time `gorm:"column:result_deadline;type:date" json:"result_deadline,omitempty"`
	IsFinished       int        `gorm:"column:is_finished;not null;default:0;index:idx_st_recruit_plan_finished,priority:1" json:"is_finished"`
	FinishedAt       *time.Time `gorm:"column:finished_at" json:"finished_at,omitempty"`
	FinishedBy       *int64     `gorm:"column:finished_by" json:"finished_by,omitempty"`
	FinishedReason   string     `gorm:"column:finished_reason;type:text" json:"finished_reason,omitempty"`
	IsDeleted        int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt        time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (StRecruitPlan) TableName() string { return "st_recruit_plan" }

// StRecruitApply 招新申请。docs/03 §6.2.6。
type StRecruitApply struct {
	ID            int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	PlanID        int64      `gorm:"column:plan_id;not null;uniqueIndex:uniq_st_recruit_apply_plan_stu,priority:1" json:"plan_id"`
	StudentID     int64      `gorm:"column:student_id;not null;uniqueIndex:uniq_st_recruit_apply_plan_stu,priority:2;index:idx_st_recruit_apply_student" json:"student_id"`
	ResumeFileID  *int64     `gorm:"column:resume_file_id" json:"resume_file_id,omitempty"`
	Result        string     `gorm:"column:result;type:text;not null;default:pending;check:result IN ('pending','accepted','rejected')" json:"result"`
	ResultAt      *time.Time `gorm:"column:result_at" json:"result_at,omitempty"`
	IsDeleted     int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (StRecruitApply) TableName() string { return "st_recruit_apply" }

// StActivity 活动立项。docs/03 §6.2.7。
type StActivity struct {
	ID                    int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo                 string    `gorm:"column:biz_no;type:text;uniqueIndex:uniq_st_activity_biz_no" json:"biz_no"`
	AssociationID         int64     `gorm:"column:association_id;not null;index:idx_st_activity_assoc_status,priority:1" json:"association_id"`
	Title                 string    `gorm:"column:title;type:text;not null" json:"title"`
	ActivityLevel         string    `gorm:"column:activity_level;type:text;not null;check:activity_level IN ('A','B','C','D');index:idx_st_activity_level,priority:1" json:"activity_level"`
	ExpectedParticipants  int       `gorm:"column:expected_participants;not null;check:expected_participants > 0" json:"expected_participants"`
	BudgetCents           int64     `gorm:"column:budget_cents;not null;default:0;check:budget_cents >= 0" json:"budget_cents"`
	PlanFileID            *int64    `gorm:"column:plan_file_id" json:"plan_file_id,omitempty"`
	EmergencyPlanFileID   *int64    `gorm:"column:emergency_plan_file_id" json:"emergency_plan_file_id,omitempty"`
	SafetyCommitFileID    *int64    `gorm:"column:safety_commit_file_id" json:"safety_commit_file_id,omitempty"`
	Location              string    `gorm:"column:location;type:text;not null" json:"location"`
	StartedAt             time.Time `gorm:"column:started_at;not null;index:idx_st_activity_started" json:"started_at"`
	EndedAt               time.Time `gorm:"column:ended_at;not null" json:"ended_at"`
	ExpectedCount         *int      `gorm:"column:expected_count" json:"expected_count,omitempty"`
	Status                string    `gorm:"column:status;type:text;not null;default:S0;check:status IN ('S0','S1','S2','S3','S4','cancelled');index:idx_st_activity_assoc_status,priority:2;index:idx_st_activity_level,priority:2" json:"status"`
	RejectCount           int       `gorm:"column:reject_count;not null;default:0" json:"reject_count"`
	LastAction            string    `gorm:"column:last_action;type:text" json:"last_action"`
	IsDeleted             int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt             time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt             time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (StActivity) TableName() string { return "st_activity" }

// StActivityApproval 活动审批流。docs/03 §6.2.8。
type StActivityApproval struct {
	ID             int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ActivityID     int64      `gorm:"column:activity_id;not null;uniqueIndex:uniq_st_act_approval_act_step,priority:1" json:"activity_id"`
	StepNo         int        `gorm:"column:step_no;not null;uniqueIndex:uniq_st_act_approval_act_step,priority:2" json:"step_no"`
	ApproverRole   string     `gorm:"column:approver_role;type:text;not null" json:"approver_role"`
	ApproverUserID *int64     `gorm:"column:approver_user_id" json:"approver_user_id,omitempty"`
	Decision       string     `gorm:"column:decision;type:text;check:decision IS NULL OR decision IN ('pass','reject')" json:"decision"`
	Opinion        string     `gorm:"column:opinion;type:text;check:opinion IS NULL OR opinion = '' OR length(opinion) >= 30" json:"opinion"`
	DecidedAt      *time.Time `gorm:"column:decided_at" json:"decided_at,omitempty"`
}

func (StActivityApproval) TableName() string { return "st_activity_approval" }

// StActivityCheckin 活动签到。docs/03 §6.2.9。
type StActivityCheckin struct {
	ID           int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ActivityID   int64     `gorm:"column:activity_id;not null;uniqueIndex:uniq_st_checkin_act_stu,priority:1;index:idx_st_checkin_activity" json:"activity_id"`
	StudentID    int64     `gorm:"column:student_id;not null;uniqueIndex:uniq_st_checkin_act_stu,priority:2;index:idx_st_checkin_student,priority:1" json:"student_id"`
	CheckinAt    time.Time `gorm:"column:checkin_at;not null;index:idx_st_checkin_student,priority:2" json:"checkin_at"`
	Method       string    `gorm:"column:method;type:text;not null;check:method IN ('qrcode','gps','manual')" json:"method"`
	IsLate       int `gorm:"column:is_late;not null;default:0" json:"is_late"`
	LateMinutes  int `gorm:"column:late_minutes;not null;default:0;check:late_minutes >= 0" json:"late_minutes"`
	IsPresent    int `gorm:"column:is_present;not null" json:"is_present"`
}

func (StActivityCheckin) TableName() string { return "st_activity_checkin" }

// StActivitySummary 活动总结。docs/03 §6.2.10。
type StActivitySummary struct {
	ID                  int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ActivityID          int64     `gorm:"column:activity_id;not null;uniqueIndex:uniq_st_act_summary_act" json:"activity_id"`
	ActualParticipants  int       `gorm:"column:actual_participants;not null;check:actual_participants >= 0" json:"actual_participants"`
	AchievementScore    *int      `gorm:"column:achievement_score;check:achievement_score IS NULL OR (achievement_score BETWEEN 0 AND 100)" json:"achievement_score,omitempty"`
	Suggestions         string    `gorm:"column:suggestions;type:text" json:"suggestions"`
	SubmittedAt         time.Time `gorm:"column:submitted_at;not null" json:"submitted_at"`
	IsOverdue           int       `gorm:"column:is_overdue;not null;default:0" json:"is_overdue"`
	IsDeleted           int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt           time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (StActivitySummary) TableName() string { return "st_activity_summary" }

// StActivityPhoto 活动照片。docs/03 §6.2.11。
type StActivityPhoto struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ActivityID int64      `gorm:"column:activity_id;not null;index" json:"activity_id"`
	FileID     int64      `gorm:"column:file_id;not null" json:"file_id"`
	Caption    string     `gorm:"column:caption;type:text" json:"caption"`
	TakenAt    *time.Time `gorm:"column:taken_at" json:"taken_at,omitempty"`
	IsDeleted  int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt  time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (StActivityPhoto) TableName() string { return "st_activity_photo" }

// StExpense 经费报销。docs/03 §6.2.12。
type StExpense struct {
	ID            int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo         string     `gorm:"column:biz_no;type:text;uniqueIndex:uniq_st_expense_biz_no" json:"biz_no"`
	ActivityID    int64      `gorm:"column:activity_id;not null;index" json:"activity_id"`
	AmountCents   int64      `gorm:"column:amount_cents;not null;check:amount_cents > 0" json:"amount_cents"`
	InvoiceCount  int        `gorm:"column:invoice_count;not null;default:1" json:"invoice_count"`
	InvoiceFiles  string     `gorm:"column:invoice_files;type:text" json:"invoice_files"`
	Status        string     `gorm:"column:status;type:text;not null;default:S1;check:status IN ('S1','S3','S4')" json:"status"`
	ReviewedBy    *int64     `gorm:"column:reviewed_by" json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time `gorm:"column:reviewed_at" json:"reviewed_at,omitempty"`
	CoSignedBy    *int64     `gorm:"column:co_signed_by" json:"co_signed_by,omitempty"`
	PaidAt        *time.Time `gorm:"column:paid_at" json:"paid_at,omitempty"`
	IsDeleted     int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (StExpense) TableName() string { return "st_expense" }

// StElection 社团换届。docs/03 §6.2.13。
type StElection struct {
	ID                     int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	BizNo                  string     `gorm:"column:biz_no;type:text" json:"biz_no"`
	AssociationID          int64      `gorm:"column:association_id;not null;index" json:"association_id"`
	TermStart              time.Time  `gorm:"column:term_start;type:date;not null" json:"term_start"`
	TermEnd                time.Time  `gorm:"column:term_end;type:date;not null" json:"term_end"`
	OldPresidentStudentID  *int64     `gorm:"column:old_president_student_id" json:"old_president_student_id,omitempty"`
	NewPresidentStudentID  int64      `gorm:"column:new_president_student_id;not null" json:"new_president_student_id"`
	WorkReportFileID       *int64     `gorm:"column:work_report_file_id" json:"work_report_file_id,omitempty"`
	PlanFileID             *int64     `gorm:"column:plan_file_id" json:"plan_file_id,omitempty"`
	PublicStart            *time.Time `gorm:"column:public_start;type:date" json:"public_start,omitempty"`
	PublicEnd              *time.Time `gorm:"column:public_end;type:date" json:"public_end,omitempty"`
	Status                 string     `gorm:"column:status;type:text;not null;default:S1;check:status IN ('S0','S1','S2','S3','S4')" json:"status"`
	IsDeleted              int        `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt              time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt              time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (StElection) TableName() string { return "st_election" }

// StRating 年度评优。docs/03 §6.2.14。
type StRating struct {
	ID                    int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	AssociationID         int64     `gorm:"column:association_id;not null;uniqueIndex:uniq_st_rating_assoc_year,priority:1" json:"association_id"`
	AcademicYear          string    `gorm:"column:academic_year;type:text;not null;uniqueIndex:uniq_st_rating_assoc_year,priority:2" json:"academic_year"`
	DimensionActivity     int       `gorm:"column:dimension_activity;not null;check:dimension_activity BETWEEN 0 AND 100" json:"dimension_activity"`
	DimensionMemberActive int       `gorm:"column:dimension_member_active;not null;check:dimension_member_active BETWEEN 0 AND 100" json:"dimension_member_active"`
	DimensionFinance      int       `gorm:"column:dimension_finance;not null;check:dimension_finance BETWEEN 0 AND 100" json:"dimension_finance"`
	DimensionBrand        int       `gorm:"column:dimension_brand;not null;check:dimension_brand BETWEEN 0 AND 100" json:"dimension_brand"`
	DimensionSatisfaction int       `gorm:"column:dimension_satisfaction;not null;check:dimension_satisfaction BETWEEN 0 AND 100" json:"dimension_satisfaction"`
	WeightedScore         float64   `gorm:"column:weighted_score;not null" json:"weighted_score"`
	Star                  int       `gorm:"column:star;not null;check:star BETWEEN 1 AND 5" json:"star"`
	PublicVoteCount       *int      `gorm:"column:public_vote_count" json:"public_vote_count,omitempty"`
	Status                string    `gorm:"column:status;type:text;not null;default:S1;check:status IN ('S1','S2','S3')" json:"status"`
	IsDeleted             int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt             time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt             time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (StRating) TableName() string { return "st_rating" }

// StBlacklist 黑名单。docs/03 §6.2.15。
type StBlacklist struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID    int64     `gorm:"column:user_id;not null;index:idx_st_blacklist_user,priority:1" json:"user_id"`
	Reason    string    `gorm:"column:reason;type:text;not null" json:"reason"`
	StartedAt time.Time `gorm:"column:started_at;type:date;not null" json:"started_at"`
	EndedAt   time.Time `gorm:"column:ended_at;type:date;not null;index:idx_st_blacklist_user,priority:2" json:"ended_at"`
	IsDeleted int       `gorm:"column:is_deleted;not null;default:0" json:"is_deleted"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (StBlacklist) TableName() string { return "st_blacklist" }
