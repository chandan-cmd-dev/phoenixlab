package main

import (
	"PhoenixLab/models"
	_ "PhoenixLab/routers"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/session"
	_ "github.com/lib/pq"
)

func getDBConn() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	if v, err := web.AppConfig.String("dbconn"); err == nil && v != "" {
		return v
	}
	return "host=localhost port=5432 user=tracker password=secret dbname=phoenixlab sslmode=disable"
}

func init() {
	orm.RegisterDriver("postgres", orm.DRPostgres)

	orm.RegisterModel(
		new(models.Branch),
		new(models.User),
		new(models.PartCatalog),
		new(models.Ticket),
		new(models.TicketPart),
		new(models.AuditLog),
		new(models.OdooSyncLog),
		new(models.Comment),
		new(models.TicketWorkflow),
	)
}

func connectWithRetry(maxAttempts int) error {
	dbConn := getDBConn()
	var lastErr error
	for i := 1; i <= maxAttempts; i++ {
		if err := orm.RegisterDataBase("default", "postgres", dbConn); err != nil {
			lastErr = err
			log.Printf("DB connection attempt %d/%d failed: %v — retrying in 3s...", i, maxAttempts, err)
			time.Sleep(3 * time.Second)
			continue
		}
		if err := orm.RunSyncdb("default", false, true); err != nil {
			log.Printf("Schema sync warning: %v", err)
		}
		return nil
	}
	return fmt.Errorf("could not connect to database after %d attempts: %w", maxAttempts, lastErr)
}

func runMigrations() {
	o := orm.NewOrm()

	// Check if schema_migrations table already exists (i.e. migrations ran before)
	var smCount int
	o.Raw(`SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'schema_migrations'`).QueryRow(&smCount)

	freshMigrationTable := smCount == 0

	_, err := o.Raw(`CREATE TABLE IF NOT EXISTS schema_migrations (
		filename TEXT PRIMARY KEY,
		applied_at TIMESTAMPTZ DEFAULT NOW()
	)`).Exec()
	if err != nil {
		log.Printf("Warning: could not create schema_migrations table: %v", err)
		return
	}

	// If schema_migrations was just created, check if this is an EXISTING production DB
	// by looking for the tickets table. If it exists, pre-seed schema_migrations with all
	// migrations that created/altered tables already present so they are SKIPPED.
	if freshMigrationTable {
		var ticketsExist int
		o.Raw(`SELECT COUNT(*) FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'tickets'`).QueryRow(&ticketsExist)

		if ticketsExist > 0 {
			log.Println("Migration: existing production database detected — pre-seeding schema_migrations")
			// These migrations created/modified tables that already exist in production.
			// Mark them as applied so they are never re-run.
			baseline := []string{
				"001_create_branches.sql",
				"002_create_users.sql",
				"003_create_parts_catalog.sql",
				"004_create_tickets.sql",
				"005_create_ticket_parts.sql",
				"006_create_audit_log.sql",
				"007_create_odoo_sync_log.sql",
				"008_add_excel_fields.sql",
				"009_create_comments.sql",
			}
			for _, name := range baseline {
				o.Raw("INSERT INTO schema_migrations (filename) VALUES (?) ON CONFLICT DO NOTHING", name).Exec()
			}
			log.Printf("Migration: marked %d baseline migrations as already applied", len(baseline))
		}
	}

	files, err := filepath.Glob("migrations/*.sql")
	if err != nil || len(files) == 0 {
		log.Printf("No migration files found")
		return
	}
	sort.Strings(files)

	for _, f := range files {
		basename := filepath.Base(f)
		var count int
		o.Raw("SELECT COUNT(*) FROM schema_migrations WHERE filename = ?", basename).QueryRow(&count)
		if count > 0 {
			continue
		}

		content, err := os.ReadFile(f)
		if err != nil {
			log.Printf("Migration error reading %s: %v", basename, err)
			continue
		}

		log.Printf("Migration: applying %s ...", basename)

		db, _ := orm.GetDB("default")
		tx, err := db.Begin()
		if err != nil {
			log.Printf("Migration error starting tx for %s: %v", basename, err)
			continue
		}

		statements := strings.Split(string(content), ";")
		var txErr error
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := tx.Exec(stmt); err != nil {
				txErr = err
				break
			}
		}

		if txErr != nil {
			tx.Rollback()
			log.Printf("Migration %s failed: %v", basename, txErr)
			continue
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (filename) VALUES ($1)", basename); err != nil {
			tx.Rollback()
			log.Printf("Migration %s: failed to record: %v", basename, err)
			continue
		}

		if err := tx.Commit(); err != nil {
			log.Printf("Migration %s: commit failed: %v", basename, err)
			continue
		}
		log.Printf("Migration applied: %s", basename)
	}
}

var i18nData map[string]map[string]string

func loadI18n() {
	i18nData = make(map[string]map[string]string)
	langs := []string{"en", "zh", "es"}
	for _, lang := range langs {
		path := fmt.Sprintf("static/i18n/%s.json", lang)
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("i18n: could not load %s: %v", path, err)
			continue
		}
		var m map[string]string
		if err := json.Unmarshal(data, &m); err != nil {
			log.Printf("i18n: could not parse %s: %v", path, err)
			continue
		}
		i18nData[lang] = m
	}
}

func t(lang, key string) string {
	if m, ok := i18nData[lang]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	if m, ok := i18nData["en"]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	return key
}

func main() {
	if err := connectWithRetry(10); err != nil {
		log.Fatalf("Fatal: %v", err)
	}

	runMigrations()
	loadI18n()

	if web.GlobalSessions == nil {
		var err error
		web.GlobalSessions, err = session.NewManager("memory", &session.ManagerConfig{
			CookieName:      "phoenixlab_session",
			Gclifetime:      86400,
			Maxlifetime:     86400,
			CookieLifeTime:  86400,
			EnableSetCookie: true,
		})
		if err != nil {
			log.Fatalf("Session init error: %v", err)
		}
		go web.GlobalSessions.GC()
	}

	web.AddFuncMap("list", func(args ...string) []string {
		return args
	})
	web.AddFuncMap("auditIcon", func(action string) string {
		icons := map[string]string{
			"create": "✨", "update": "✏️", "delete": "🗑️",
			"login": "👁️", "push": "🔗", "pull": "⬇️",
		}
		if v, ok := icons[action]; ok {
			return v
		}
		return "•"
	})

	web.AddFuncMap("statusLabel", func(s string) string {
		return strings.ReplaceAll(s, "_", " ")
	})

	web.AddFuncMap("formatCurrency", func(f float64) string {
		return fmt.Sprintf("$%.2f", f)
	})

	web.AddFuncMap("formatDate", func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format("02 Jan 2006")
	})

	web.AddFuncMap("formatDateTime", func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format("02 Jan 2006 15:04")
	})

	web.AddFuncMap("formatDateInput", func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format("2006-01-02")
	})

	web.AddFuncMap("formatTAT", func(hours float64) string {
		if hours == 0 {
			return "N/A"
		}
		if hours < 24 {
			return fmt.Sprintf("%.1fh", hours)
		}
		days := hours / 24
		return fmt.Sprintf("%.1fd", days)
	})

	web.AddFuncMap("truncate", func(text string, length int) string {
		if len(text) <= length {
			return text
		}
		return text[:length] + "..."
	})

	web.AddFuncMap("eq", func(a, b interface{}) bool {
		return a == b
	})

	web.AddFuncMap("add", func(a, b int) int {
		return a + b
	})

	web.AddFuncMap("minus", func(a, b int) int {
		return a - b
	})

	web.AddFuncMap("pct", func(part, total int) int {
		if total == 0 {
			return 0
		}
		return part * 100 / total
	})

	web.AddFuncMap("substr", func(s string, start, end int) string {
		if start >= len(s) {
			return ""
		}
		if end > len(s) {
			end = len(s)
		}
		return s[start:end]
	})

	web.AddFuncMap("seq", func(start, end int) []int {
		var result []int
		for i := start; i <= end; i++ {
			result = append(result, i)
		}
		return result
	})

	web.AddFuncMap("t", t)

	web.AddFuncMap("wfStepLabel", func(step string) string {
		if label, ok := models.WorkflowStepLabels[step]; ok {
			return label
		}
		return strings.ReplaceAll(step, "_", " ")
	})

	web.AddFuncMap("wfTypeLabel", func(wfType string) string {
		if label, ok := models.WorkflowTypeLabels[wfType]; ok {
			return label
		}
		return strings.ReplaceAll(wfType, "_", " ")
	})

	web.AddFuncMap("formatDateTZ", func(t time.Time, tz string) string {
		if t.IsZero() {
			return ""
		}
		if loc, err := time.LoadLocation(tz); err == nil {
			t = t.In(loc)
		}
		return t.Format("02 Jan 2006")
	})

	web.AddFuncMap("formatDateTimeTZ", func(t time.Time, tz string) string {
		if t.IsZero() {
			return ""
		}
		if loc, err := time.LoadLocation(tz); err == nil {
			t = t.In(loc)
		}
		return t.Format("02 Jan 2006 15:04")
	})

	runmode, _ := web.AppConfig.String("runmode")
	if runmode == "dev" {
		web.BConfig.WebConfig.DirectoryIndex = true
		web.BConfig.WebConfig.EnableDocs = true
	}

	web.BConfig.WebConfig.TemplateLeft = "{{"
	web.BConfig.WebConfig.TemplateRight = "}}"

	web.Run()
}
