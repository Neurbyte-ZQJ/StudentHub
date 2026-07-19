// Package boot 负责装配并启动 HTTP 服务：配置 / 日志 / DB / 路由。
package boot

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"student-system/internal/eventx"
	"student-system/internal/middleware"
	"student-system/internal/models"
	authapi "student-system/internal/modules/auth/api"
	authjwt "student-system/internal/modules/auth/jwt"
	authrepo "student-system/internal/modules/auth/repository"
	authservice "student-system/internal/modules/auth/service"
	cmpapi "student-system/internal/modules/cmp/api"
	cmpservice "student-system/internal/modules/cmp/event"
	cmprepo "student-system/internal/modules/cmp/repository"
	cmpsvc "student-system/internal/modules/cmp/service"
	dashboardapi "student-system/internal/modules/dashboard/api"
	dashboardsvc "student-system/internal/modules/dashboard/service"
	fileapi "student-system/internal/modules/file/api"
	filerepo "student-system/internal/modules/file/repository"
	fileservice "student-system/internal/modules/file/service"
	idxapi "student-system/internal/modules/idx/api"
	idxrepo "student-system/internal/modules/idx/repository"
	idxservice "student-system/internal/modules/idx/service"
	notiapi "student-system/internal/modules/noti/api"
	notirepo "student-system/internal/modules/noti/repository"
	notiservice "student-system/internal/modules/noti/service"
	qgapi "student-system/internal/modules/qg/api"
	qgrepo "student-system/internal/modules/qg/repository"
	qgservice "student-system/internal/modules/qg/service"
	sqapi "student-system/internal/modules/sq/api"
	sqrepo "student-system/internal/modules/sq/repository"
	sqservice "student-system/internal/modules/sq/service"
	stapi "student-system/internal/modules/st/api"
	strepo "student-system/internal/modules/st/repository"
	stservice "student-system/internal/modules/st/service"
	sysapi "student-system/internal/modules/sys/api"
	sysrepo "student-system/internal/modules/sys/repository"
	sysservice "student-system/internal/modules/sys/service"
	tyapi "student-system/internal/modules/ty/api"
	tyrepo "student-system/internal/modules/ty/repository"
	tyservice "student-system/internal/modules/ty/service"
	"student-system/internal/scheduler"
	"student-system/pkg/cachex"
	"student-system/pkg/logger"
	"student-system/pkg/response"
	"student-system/pkg/revokex"
)

// Config 简易配置。
type Config struct {
	Env    string
	Port   int
	DBPath string
	JWT    JWTConfig
}

// JWTConfig JWT 配置。
type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	Issuer     string
}

func defaultConfig() *Config {
	return &Config{
		Env:    getEnv("APP_ENV", "dev"),
		Port:   getEnvInt("APP_PORT", 8080),
		DBPath: filepath.Join("data", "studenthub.db"),
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "studenthub-dev-jwt-secret-change-in-prod"),
			AccessTTL:  15 * time.Minute,
			RefreshTTL: 168 * time.Hour, // 7 天
			Issuer:     "studenthub",
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	var n int
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
		return fallback
	}
	return n
}

// Run 启动入口：装配并阻塞监听。
func Run() error {
	cfg := defaultConfig()

	zlog, err := logger.New(cfg.Env)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer func() { _ = zlog.Sync() }()

	db, err := initDB(cfg, zlog)
	if err != nil {
		return fmt.Errorf("init db: %w", err)
	}

	if err := autoMigrate(db, zlog); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	// 种子数据：admin + 角色 + 字典 + 菜单 + 院系 + 学生 + 团支部 + 学生用户 + 业务演示数据
	DeduplicateExistingData(db, zlog) // 先清理重复数据（入团申请、团课记录按学号去重）
	SeedAdmin(db, zlog)
	SeedDicts(db, zlog)
	SeedMenus(db, zlog)
	SeedColleges(db, zlog)
	SeedStudents(db, zlog)
	SeedTyBranches(db, zlog)
	SeedStudentUser(db, zlog)
	SeedApprovalUsers(db, zlog)
	SeedSQData(db, zlog)
	SeedQGData(db, zlog)
	SeedCmpRuleVersion(db, zlog)
	SeedTyApplicationForLisi(db, zlog)
	SeedOtherBusinessData(db, zlog)
	SeedExtraTestData(db, zlog) // 批量补充各模块测试数据

	// 初始化 LRU 缓存（ADR-017：5min TTL，512 条目）
	cache := cachex.New(512, 5*time.Minute)

	router := buildRouter(cfg, db, zlog, cache)

	addr := fmt.Sprintf(":%d", cfg.Port)
	zlog.Info("server listening", zap.String("addr", addr), zap.String("env", cfg.Env))
	return router.Run(addr)
}

// initDB 打开 SQLite 连接并启用 WAL 等性能开关（ADR-003）。
func initDB(cfg *Config, zlog *zap.Logger) (*gorm.DB, error) {
	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	gormCfg := &gorm.Config{
		Logger: gormlogger.New(
			zap.NewStdLog(zlog),
			gormlogger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  gormlogger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
	}

	db, err := gorm.Open(sqlite.Open(cfg.DBPath+"?_pragma=foreign_keys(1)"), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// 应用 SQLite 性能开关（docs/03 §10.3）。
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA foreign_keys = ON;",
		"PRAGMA busy_timeout = 5000;",
		"PRAGMA temp_store = MEMORY;",
	}
	for _, p := range pragmas {
		if err := db.Exec(p).Error; err != nil {
			zlog.Warn("apply pragma failed", zap.String("pragma", p), zap.Error(err))
		}
	}

	zlog.Info("sqlite connected", zap.String("path", cfg.DBPath))
	return db, nil
}

// autoMigrate 调用 GORM AutoMigrate 自动建表。
func autoMigrate(db *gorm.DB, zlog *zap.Logger) error {
	zlog.Info("auto migrate begin", zap.Int("model_count", len(models.AllModels())))
	if err := db.AutoMigrate(models.AllModels()...); err != nil {
		return err
	}
	zlog.Info("auto migrate done")
	return nil
}

// buildRouter 注册路由。
func buildRouter(cfg *Config, db *gorm.DB, zlog *zap.Logger, cache *cachex.Cache) *gin.Engine {
	if cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())
	r.Use(requestIDMiddleware())
	r.Use(utf8GuardMiddleware())

	// 初始化 JWT 管理器
	jwtManager := authjwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	// 初始化 RT 黑名单（进程内 LRU，ADR-005 决策细化）
	rtRevoke := revokex.NewLRU()

	// 初始化 Auth 模块
	userRepo := authrepo.NewUserRepository(db)
	authSvc := authservice.NewAuthService(userRepo, jwtManager, rtRevoke)
	authHandler := authapi.NewAuthHandler(authSvc, jwtManager)

	// 初始化 Sys 模块
	dictHandler := sysapi.NewDictHandler(db, cache)
	menuHandler := sysapi.NewMenuHandler(db, cache)
	orgHandler := sysapi.NewOrgHandler(db)
	sysUserRepo := sysrepo.NewUserRepository(db)
	sysUserSvc := sysservice.NewUserService(sysUserRepo)
	sysUserHandler := sysapi.NewUserHandler(sysUserSvc)

	// 初始化 IDX 模块
	studentRepo := idxrepo.NewStudentRepository(db)
	studentSvc := idxservice.NewStudentService(studentRepo)
	studentHandler := idxapi.NewStudentHandler(studentSvc)
	profileHandler := idxapi.NewProfileHandler(studentSvc)

	// 初始化 TY 模块
	bus := eventx.NewBus(db)
	tyAppRepo := tyrepo.NewApplicationRepository(db)
	tyAppSvc := tyservice.NewApplicationService(tyAppRepo, db, bus)
	tyAppHandler := tyapi.NewApplicationHandler(tyAppSvc)

	// TY 模块 - 全流程子流程初始化
	// 推优大会
	tyRecRepo := tyrepo.NewRecommendationRepository(db)
	tyRecSvc := tyservice.NewRecommendationService(tyRecRepo, tyAppRepo, db, bus)
	tyRecHandler := tyapi.NewRecommendationHandler(tyRecSvc)

	// 培养考察
	tyCultRepo := tyrepo.NewCultivationRepository(db)
	tyCultSvc := tyservice.NewCultivationService(tyCultRepo, tyAppRepo, db, bus)
	tyCultHandler := tyapi.NewCultivationHandler(tyCultSvc)

	// 发展对象
	tyDevRepo := tyrepo.NewDevelopmentObjectRepository(db)
	tyDevSvc := tyservice.NewDevelopmentObjectService(tyDevRepo, tyAppRepo, db, bus)
	tyDevHandler := tyapi.NewDevelopmentObjectHandler(tyDevSvc)

	// 政审
	tyPolRepo := tyrepo.NewPoliticalReviewRepository(db)
	tyPolSvc := tyservice.NewPoliticalReviewService(tyPolRepo, tyDevRepo, bus)
	tyPolHandler := tyapi.NewPoliticalReviewHandler(tyPolSvc)

	// 发展大会
	tyMeetRepo := tyrepo.NewDevelopmentMeetingRepository(db)
	tyMeetSvc := tyservice.NewDevelopmentMeetingService(tyMeetRepo, tyDevRepo, tyPolRepo, tyAppRepo, db, bus)
	tyMeetHandler := tyapi.NewDevelopmentMeetingHandler(tyMeetSvc)

	// 转正流程
	tyProbRepo := tyrepo.NewProbationaryRepository(db)
	tyProbSvc := tyservice.NewProbationaryService(tyProbRepo, tyAppRepo, db, bus)
	tyProbHandler := tyapi.NewProbationaryHandler(tyProbSvc)

	// 团员花名册
	tyRosterRepo := tyrepo.NewRosterRepository(db)
	tyRosterSvc := tyservice.NewRosterService(tyRosterRepo, db, bus)
	tyRosterHandler := tyapi.NewRosterHandler(tyRosterSvc)

	// 初始化 ST 模块
	stAssocRepo := strepo.NewAssociationRepository(db)
	stActRepo := strepo.NewActivityRepository(db)
	stRecRepo := strepo.NewRecruitRepository(db)
	stAssocSvc := stservice.NewAssociationService(stAssocRepo, db, bus)
	stActSvc := stservice.NewActivityService(stActRepo, db, bus)
	stRecSvc := stservice.NewRecruitService(stRecRepo, db, bus)
	stAssocHandler := stapi.NewAssociationHandler(stAssocSvc)
	stActHandler := stapi.NewActivityHandler(stActSvc)
	stRecHandler := stapi.NewRecruitHandler(stRecSvc)

	// 初始化 SQ 模块
	sqBuildingRepo := sqrepo.NewBuildingRepository(db)
	sqInspectionRepo := sqrepo.NewInspectionRepository(db)
	sqIncidentRepo := sqrepo.NewIncidentRepository(db)
	sqBuildingSvc := sqservice.NewBuildingService(sqBuildingRepo, db)
	sqInspectionSvc := sqservice.NewInspectionService(sqInspectionRepo, db)
	sqIncidentSvc := sqservice.NewIncidentService(sqIncidentRepo, db, bus)
	sqBuildingHandler := sqapi.NewBuildingHandler(sqBuildingSvc)
	sqInspectionHandler := sqapi.NewInspectionHandler(sqInspectionSvc)
	sqIncidentHandler := sqapi.NewIncidentHandler(sqIncidentSvc)

	// 初始化 QG 模块
	qgDiffRepo := qgrepo.NewDifficultyRepository(db)
	qgPosRepo := qgrepo.NewPositionRepository(db)
	qgAttendRepo := qgrepo.NewAttendanceRepository(db)
	qgAssessRepo := qgrepo.NewAssessmentRepository(db)
	qgDiffSvc := qgservice.NewDifficultyService(qgDiffRepo, db, bus)
	qgPosSvc := qgservice.NewPositionService(qgPosRepo, db, bus)
	qgAttendSvc := qgservice.NewAttendanceService(qgAttendRepo, db)
	qgAssessSvc := qgservice.NewAssessmentService(qgAssessRepo, db, bus)
	qgDiffHandler := qgapi.NewDifficultyHandler(qgDiffSvc)
	qgPosHandler := qgapi.NewPositionHandler(qgPosSvc)
	qgAttendHandler := qgapi.NewAttendanceHandler(qgAttendSvc)
	qgAssessHandler := qgapi.NewAssessmentHandler(qgAssessSvc)

	// 初始化 File 模块（S11）
	fileRepo := filerepo.NewFileRepository(db)
	localStorage := fileservice.NewLocalStorage("storage")
	fileSvc := fileservice.NewFileService(fileRepo, localStorage, db)
	fileHandler := fileapi.NewFileHandler(fileSvc)

	// 初始化 Notification 模块（S11）
	notiRepo := notirepo.NewNotificationRepository(db)
	notiSvc := notiservice.NewNotificationService(notiRepo)
	notiHandler := notiapi.NewNotificationHandler(notiSvc)

	// 初始化 CMP 模块（S12）
	cmpRepo := cmprepo.NewScoreRepository(db)
	cmpCalc := cmpsvc.NewCalculator(db, cmpRepo, bus)
	cmpSvc := cmpsvc.NewScoreService(db, cmpRepo, cmpCalc, bus)
	cmpDashboardSvc := cmpsvc.NewDashboardService(db)
	cmpRuleSvc := cmpsvc.NewRuleVersionService(db, cmpRepo)
	cmpScoreHandler := cmpapi.NewScoreHandler(cmpSvc)
	cmpDashboardHandler := cmpapi.NewDashboardHandler(cmpDashboardSvc)
	cmpRuleHandler := cmpapi.NewRuleVersionHandler(cmpRuleSvc)

	// 初始化 Dashboard 模块
	dashboardSvc := dashboardsvc.NewDashboardService(db)
	dashboardHandler := dashboardapi.NewDashboardHandler(dashboardSvc)

	// 注册 CMP 事件订阅器（订阅 TY/ST/SQ/QG 关键事件 → 触发增量重算）
	cmpSubscriber := cmpservice.NewEventSubscriber(db, cmpSvc, zlog)
	cmpSubscriber.RegisterBusSubscriptions(bus)

	// 初始化 Scheduler 定时任务（S11）
	sched := scheduler.NewScheduler(db, zlog)
	sched.Start()
	jobHandler := scheduler.NewJobHandler(sched, db)

	api := r.Group("/api/v1")
	{
		api.GET("/healthz", func(c *gin.Context) {
			response.OK(c, gin.H{"status": "healthy"})
		})
		api.GET("/readyz", func(c *gin.Context) {
			sqlDB, err := db.DB()
			if err != nil {
				response.Fail(c, 1500, "db handle unavailable")
				return
			}
			if err := sqlDB.Ping(); err != nil {
				response.Fail(c, 1500, "db ping failed")
				return
			}
			response.OK(c, gin.H{"db": "ok"})
		})
	}

	// Auth 路由（公开）
	authHandler.RegisterRoutes(api)

	// 受保护路由（需要 JWT 认证）
	protected := api.Group("")
	protected.Use(middleware.JWTAuth(jwtManager, userRepo))
	{
		authHandler.RegisterProtectedRoutes(protected)
	}

	// Sys 路由（受保护）
	adminOnly := middleware.RequireRoles("R-SY-ADMIN")
	dictHandler.RegisterRoutes(protected, adminOnly)
	menuHandler.RegisterRoutes(protected)
	orgHandler.RegisterRoutes(protected, adminOnly)
	sysUserHandler.RegisterRoutes(protected, adminOnly)

	// IDX 路由（受保护）
	studentHandler.RegisterRoutes(protected, adminOnly)
	profileHandler.RegisterRoutes(protected)

	// TY 路由（受保护）
	tyAppHandler.RegisterRoutes(protected, adminOnly)

	// TY 全流程路由
	tyRecHandler.RegisterRoutes(protected, adminOnly)
	tyCultHandler.RegisterRoutes(protected, adminOnly)
	tyDevHandler.RegisterRoutes(protected, adminOnly)
	tyPolHandler.RegisterRoutes(protected, adminOnly)
	tyMeetHandler.RegisterRoutes(protected, adminOnly)
	tyProbHandler.RegisterRoutes(protected, adminOnly)
	tyRosterHandler.RegisterRoutes(protected, adminOnly)

	// ST 路由（受保护）
	stAssocHandler.RegisterRoutes(protected, adminOnly)
	stActHandler.RegisterRoutes(protected, adminOnly)
	stRecHandler.RegisterRoutes(protected, adminOnly)

	// SQ 路由（受保护）
	sqBuildingHandler.RegisterRoutes(protected, adminOnly)
	sqInspectionHandler.RegisterRoutes(protected, adminOnly)
	sqIncidentHandler.RegisterRoutes(protected, adminOnly)

	// QG 路由（受保护）
	qgDiffHandler.RegisterRoutes(protected, adminOnly)
	qgPosHandler.RegisterRoutes(protected, adminOnly)
	qgAttendHandler.RegisterRoutes(protected, adminOnly)
	qgAssessHandler.RegisterRoutes(protected, adminOnly)

	// File 路由（S11，受保护）
	fileHandler.RegisterRoutes(protected, adminOnly)

	// Notification 路由（S11，受保护，所有登录用户可访问）
	notiHandler.RegisterRoutes(protected)

	// Job 定时任务路由（S11，仅 admin）
	jobHandler.RegisterRoutes(protected, adminOnly)

	// CMP 路由（S12）
	cmpScoreHandler.RegisterRoutes(protected)
	cmpDashboardHandler.RegisterRoutes(protected)
	cmpRuleHandler.RegisterRoutes(protected, adminOnly)

	// Dashboard 路由（受保护，所有登录用户可访问）
	dashboardHandler.RegisterRoutes(protected)

	// SPA 静态资源服务（必须在 NoRoute 之前注册）
	// 使用自定义 handler 确保正确的 MIME 类型（避免 JS 被当作 text/html）
	frontendDist := filepath.Join("frontend", "dist")
	r.GET("/assets/*filepath", func(c *gin.Context) {
		c.File(filepath.Join(frontendDist, "assets", c.Param("filepath")))
	})
	r.GET("/images/*filepath", func(c *gin.Context) {
		c.File(filepath.Join(frontendDist, "images", c.Param("filepath")))
	})

	r.NoRoute(func(c *gin.Context) {
		// SPA fallback：非 /api 开头的路径回退到 index.html（前端路由接管）
		path := c.Request.URL.Path
		if !strings.HasPrefix(path, "/api") {
			c.File(filepath.Join(frontendDist, "index.html"))
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"code": 1404, "message": "not found"})
	})

	return r
}
