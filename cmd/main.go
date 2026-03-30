package main

import (
	"PhoenixLab/models"
	_ "PhoenixLab/routers"
	"fmt"
	"log"
	"os"
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

func main() {
	if err := connectWithRetry(10); err != nil {
		log.Fatalf("Fatal: %v", err)
	}

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

	runmode, _ := web.AppConfig.String("runmode")
	if runmode == "dev" {
		web.BConfig.WebConfig.DirectoryIndex = true
		web.BConfig.WebConfig.EnableDocs = true
	}

	web.BConfig.WebConfig.TemplateLeft = "{{"
	web.BConfig.WebConfig.TemplateRight = "}}"

	web.Run()
}
