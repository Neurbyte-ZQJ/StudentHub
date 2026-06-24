// 一次性脚本：排查 / 清理 sys_menu 表中的孤儿菜单。
//
// 用途：当 sys_menu.Code 唯一索引只防同 code 重复时，重命名 code
// （如 mine-ty-development → mine-development）后再跑 SeedMenus
// 不会覆盖旧记录，会出现两条「title 相同 code 不同」的菜单同时展示。
//
// 用法：
//   go run cmd/dbcheckmenu/main.go                          # 列出可疑重复（默认扫描 title='我的团员发展'）
//   go run cmd/dbcheckmenu/main.go --title "我的团员发展"    # 自定义扫描 title
//   go run cmd/dbcheckmenu/main.go --cleanup --code mine-development  # 软删指定 code
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type row struct {
	ID        int64
	Code      string
	Title     string
	Path      string
	IsDeleted int
	Sort      int
}

func main() {
	title := flag.String("title", "我的团员发展", "按 title 扫描可疑重复")
	cleanup := flag.Bool("cleanup", false, "进入清理模式：需配合 --code 使用")
	code := flag.String("code", "", "清理模式下要软删的 code（可重复传入）")
	flag.Parse()

	// 解析 db path（兼容 main 进程与 go run 临时目录）
	exe, _ := os.Executable()
	root := filepath.Dir(filepath.Dir(exe))
	dbPath := filepath.Join(root, "data", "studenthub.db")
	if _, err := os.Stat(dbPath); err != nil {
		dbPath = "./data/studenthub.db"
	}
	if _, err := os.Stat(dbPath); err != nil {
		log.Fatalf("db not found: %s", dbPath)
	}

	db, err := gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	if *cleanup {
		if *code == "" {
			log.Fatalf("--cleanup 必须配合 --code 使用，例: --code mine-development")
		}
		res := db.Exec("UPDATE sys_menu SET is_deleted = 1 WHERE code = ? AND is_deleted = 0", *code)
		if res.Error != nil {
			log.Fatalf("cleanup: %v", res.Error)
		}
		fmt.Printf("✓ 已软删 code=%q 共 %d 行\n", *code, res.RowsAffected)
		fmt.Println("提示：重启后端让 SeedMenus 重新跑一次，确认孤儿已被清掉。")
		return
	}

	// 扫描模式
	var rows []row
	if err := db.Raw("SELECT id, code, title, path, is_deleted, sort FROM sys_menu WHERE title = ? ORDER BY id ASC", *title).Scan(&rows).Error; err != nil {
		log.Fatalf("query: %v", err)
	}

	fmt.Printf("扫到 %d 条 title=%q 记录：\n", len(rows), *title)
	for _, m := range rows {
		fmt.Printf("  id=%-3d code=%-22s path=%-28s is_deleted=%d sort=%d\n",
			m.ID, m.Code, m.Path, m.IsDeleted, m.Sort)
	}

	alive := 0
	for _, m := range rows {
		if m.IsDeleted == 0 {
			alive++
		}
	}
	if alive > 1 {
		fmt.Printf("\n⚠ 检测到 %d 条「有效」重复菜单，请确认要保留哪条。\n", alive)
		fmt.Println("示例：go run cmd/dbcheckmenu/main.go --cleanup --code mine-development")
	}
}
