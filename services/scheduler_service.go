package services

import (
	"PhoenixLab/models"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

var schedulerInflight sync.Map

func StartScheduler() {
	go func() {
		time.Sleep(15 * time.Second)
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		runDueConnections()
		for range ticker.C {
			runDueConnections()
		}
	}()
	log.Println("Sheet auto-sync scheduler started")
}

func runDueConnections() {
	o := orm.NewOrm()
	var conns []*models.SheetConnection
	_, err := o.QueryTable("sheet_connections").
		Filter("auto_sync_enabled", true).
		Filter("status", "mapped").All(&conns)
	if err != nil {
		return
	}
	for _, c := range conns {
		interval := time.Duration(c.EffectiveIntervalMinutes()) * time.Minute
		if !c.LastAutoRunAt.IsZero() && time.Since(c.LastAutoRunAt) < interval {
			continue
		}
		if _, busy := schedulerInflight.LoadOrStore(c.Id, true); busy {
			continue
		}
		go func(conn *models.SheetConnection) {
			defer schedulerInflight.Delete(conn.Id)
			runScheduledSync(conn)
		}(c)
	}
}

func runScheduledSync(conn *models.SheetConnection) {
	defer func() {
		if r := recover(); r != nil {
			recordAutoRun(conn, "error", fmt.Sprintf("panic: %v", r))
		}
	}()

	sync := SheetSyncService{}
	var res *SyncResult
	var err error
	switch conn.SyncDirection {
	case "pull":
		res, err = sync.Import(conn, conn.CreatedBy)
	case "push":
		res, err = sync.Push(conn, conn.CreatedBy)
	default:
		res, err = sync.Reconcile(conn, conn.CreatedBy)
	}
	if err != nil {
		recordAutoRun(conn, "error", err.Error())
		return
	}
	recordAutoRun(conn, "ok", res.Summary())
}

func recordAutoRun(conn *models.SheetConnection, status, msg string) {
	o := orm.NewOrm()
	conn.LastAutoRunAt = time.Now()
	conn.LastAutoStatus = status
	conn.LastAutoMessage = msg
	o.Update(conn, "LastAutoRunAt", "LastAutoStatus", "LastAutoMessage", "UpdatedAt")
}
