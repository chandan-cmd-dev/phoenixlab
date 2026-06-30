package controllers

import (
	"PhoenixLab/models"
	"PhoenixLab/services"
	"fmt"
	"strconv"
	"strings"
)

type SheetController struct {
	BaseController
}

func (c *SheetController) loadConn() *models.SheetConnection {
	id, err := c.GetIntParam(":id")
	if err != nil {
		c.Abort("404")
		return nil
	}
	conn, err := (&services.SheetConnectionService{}).GetByID(id)
	if err != nil {
		c.Abort("404")
		return nil
	}
	if !c.GetCurrentUser().IsSuperAdmin() && conn.BranchId != c.GetCurrentUser().BranchId {
		c.Abort("403")
		return nil
	}
	return conn
}

type connListView struct {
	*models.SheetConnection
	OpenConflicts int64
	OpenAdoptions int64
}

func (c *SheetController) List() {
	c.RequireRole("admin", "super_admin")
	oauth := services.OAuthService{}
	connSvc := services.SheetConnectionService{}
	conflictSvc := services.SheetConflictService{}
	adoptionSvc := services.SheetAdoptionService{}

	conns, _ := connSvc.List(c.GetBranchScope())
	views := make([]connListView, 0, len(conns))
	for _, cn := range conns {
		views = append(views, connListView{
			SheetConnection: cn,
			OpenConflicts:   conflictSvc.CountOpen(cn.Id),
			OpenAdoptions:   adoptionSvc.CountOpen(cn.Id),
		})
	}

	c.Data["connections"] = views
	c.Data["googleConnected"] = oauth.HasToken()
	c.Data["oauthConfigured"] = oauth.Configured()
	if tok, err := oauth.LoadToken(); err == nil {
		c.Data["googleEmail"] = tok.AccountEmail
	}
	c.Data["title"] = "Google Sheets Sync"
	c.SetActivePage("sheets")
	c.GetFlashMessages()
	c.TplName = "sheets/list.html"
}

func (c *SheetController) Connect() {
	c.RequireRole("admin", "super_admin")
	if !(&services.OAuthService{}).HasToken() {
		c.FlashError("Connect a Google account first")
		c.Redirect("/sheets", 302)
		return
	}
	var branches []*models.Branch
	if c.GetCurrentUser().IsSuperAdmin() {
		branches, _ = models.GetAllBranches()
	}
	c.Data["branches"] = branches
	c.Data["title"] = "Connect a Spreadsheet"
	c.SetActivePage("sheets")
	c.GetFlashMessages()
	c.TplName = "sheets/connect.html"
}

func (c *SheetController) DoConnect() {
	c.RequireRole("admin", "super_admin")
	input := strings.TrimSpace(c.GetString("spreadsheet_id"))
	if input == "" {
		c.FlashError("Please enter a spreadsheet URL or ID")
		c.Redirect("/sheets/connect", 302)
		return
	}
	branchID := c.GetCurrentUser().BranchId
	if c.GetCurrentUser().IsSuperAdmin() {
		if bid, err := c.GetInt("branch_id"); err == nil && bid > 0 {
			branchID = bid
		}
	}
	if branchID == 0 {
		c.FlashError("Please select a branch")
		c.Redirect("/sheets/connect", 302)
		return
	}
	conn, err := (&services.SheetConnectionService{}).Create(input, branchID, c.GetCurrentUser().Id)
	if err != nil {
		c.FlashError("Could not connect: " + err.Error())
		c.Redirect("/sheets/connect", 302)
		return
	}
	c.FlashSuccess("Spreadsheet connected — configure the column mapping")
	c.Redirect(fmt.Sprintf("/sheets/%d/mapping", conn.Id), 302)
}

func (c *SheetController) Mapping() {
	c.RequireRole("admin", "super_admin")
	conn := c.loadConn()
	if conn == nil {
		return
	}

	client, err := services.NewSheetsClient()
	if err != nil {
		c.FlashError("Google client error: " + err.Error())
		c.Redirect("/sheets", 302)
		return
	}
	meta, err := client.GetMeta(conn.SpreadsheetId)
	if err != nil {
		c.FlashError("Could not read spreadsheet: " + err.Error())
		c.Redirect("/sheets", 302)
		return
	}

	mapSvc := services.SheetMappingService{}

	reSuggest := false
	if t := c.GetString("tab"); t != "" && t != conn.TabName {
		conn.TabName = t
		conn.Brand = detectBrandFromName(t)
		reSuggest = true
	}
	if hrStr := c.GetString("header_row"); hrStr != "" {
		if hr, e := strconv.Atoi(hrStr); e == nil {
			conn.HeaderRow = hr
			reSuggest = true
		}
	}

	rows, _ := client.GetRows(conn.SpreadsheetId, conn.TabName)
	existing, _ := mapSvc.LoadMappings(conn.Id)

	headerRow := conn.HeaderRow
	if (len(existing) == 0 || reSuggest) && c.GetString("header_row") == "" && len(rows) > 0 {
		headerRow = mapSvc.DetectHeaderRow(rows)
		conn.HeaderRow = headerRow
	}

	var headers []string
	if headerRow >= 0 && headerRow < len(rows) {
		headers = rows[headerRow]
	}

	var mappings []*models.SheetColumnMapping
	if len(existing) > 0 && !reSuggest {
		mappings = existing
	} else {
		mappings = mapSvc.SuggestMapping(headers, conn.Brand)
	}

	firstSample := make([]string, len(headers))
	if headerRow+1 < len(rows) {
		for i, v := range rows[headerRow+1] {
			if i < len(firstSample) {
				firstSample[i] = v
			}
		}
	}

	c.Data["conn"] = conn
	c.Data["tabs"] = meta.Tabs
	c.Data["headers"] = headers
	c.Data["firstSample"] = firstSample
	c.Data["mappings"] = mappings
	c.Data["headerRow"] = headerRow
	c.Data["fieldCatalog"] = mapSvc.FieldCatalog()
	c.Data["identityOptions"] = mapSvc.IdentityKeyOptions()
	identitySelected := map[string]bool{}
	for _, f := range conn.IdentityFields() {
		identitySelected[f] = true
	}
	c.Data["identitySelectedMap"] = identitySelected
	c.Data["title"] = "Column Mapping"
	c.SetActivePage("sheets")
	c.GetFlashMessages()
	c.TplName = "sheets/mapping.html"
}

func (c *SheetController) SaveMapping() {
	c.RequireRole("admin", "super_admin")
	conn := c.loadConn()
	if conn == nil {
		return
	}
	connSvc := services.SheetConnectionService{}

	if tab := c.GetString("tab_name"); tab != "" {
		conn.TabName = tab
	}
	if hr, err := c.GetInt("header_row"); err == nil {
		conn.HeaderRow = hr
	}
	if dir := c.GetString("sync_direction"); dir != "" {
		conn.SyncDirection = dir
	}
	if brand := c.GetString("brand"); brand != "" {
		conn.Brand = brand
	}
	if idKeys := c.GetStrings("identity_key"); len(idKeys) > 0 {
		conn.IdentityKey = strings.Join(idKeys, ",")
	}
	autoVal := c.GetString("auto_sync_enabled")
	conn.AutoSyncEnabled = autoVal == "on" || autoVal == "1" || autoVal == "true"
	if iv, err := c.GetInt("sync_interval_minutes"); err == nil && iv > 0 {
		conn.SyncIntervalMinutes = iv
	}
	conn.Status = "mapped"
	if err := connSvc.Save(conn); err != nil {
		c.FlashError("Could not save connection: " + err.Error())
		c.Redirect(fmt.Sprintf("/sheets/%d/mapping", conn.Id), 302)
		return
	}

	colCount, _ := c.GetInt("col_count")
	var mappings []*models.SheetColumnMapping
	for i := 0; i < colCount; i++ {
		target := c.GetString(fmt.Sprintf("target_%d", i))
		if target == "" {
			target = "ignore"
		}
		mappings = append(mappings, &models.SheetColumnMapping{
			ColumnIndex: i,
			Header:      c.GetString(fmt.Sprintf("header_%d", i)),
			TargetField: target,
			Transform:   c.GetString(fmt.Sprintf("transform_%d", i)),
		})
	}
	if err := (&services.SheetMappingService{}).SaveMapping(conn.Id, mappings); err != nil {
		c.FlashError("Could not save mapping: " + err.Error())
		c.Redirect(fmt.Sprintf("/sheets/%d/mapping", conn.Id), 302)
		return
	}
	c.FlashSuccess("Mapping saved")
	c.Redirect(fmt.Sprintf("/sheets/%d/preview", conn.Id), 302)
}

func (c *SheetController) Preview() {
	c.RequireRole("admin", "super_admin")
	conn := c.loadConn()
	if conn == nil {
		return
	}
	p, err := (&services.SheetSyncService{}).Preview(conn)
	if err != nil {
		c.FlashError("Preview failed: " + err.Error())
		c.Redirect("/sheets", 302)
		return
	}
	c.Data["conn"] = conn
	c.Data["preview"] = p
	c.Data["openConflicts"] = (&services.SheetConflictService{}).CountOpen(conn.Id)
	c.Data["openAdoptions"] = (&services.SheetAdoptionService{}).CountOpen(conn.Id)
	c.Data["title"] = "Sync Preview"
	c.SetActivePage("sheets")
	c.GetFlashMessages()
	c.TplName = "sheets/preview.html"
}

func (c *SheetController) DoImport() {
	c.RequireRole("admin", "super_admin")
	conn := c.loadConn()
	if conn == nil {
		return
	}
	res, err := (&services.SheetSyncService{}).Import(conn, c.GetCurrentUser().Id)
	if err != nil {
		c.FlashError("Import failed: " + err.Error())
		c.Redirect(fmt.Sprintf("/sheets/%d/preview", conn.Id), 302)
		return
	}
	c.FlashSuccess("Import complete — " + res.Summary())
	c.Redirect(fmt.Sprintf("/sheets/%d/preview", conn.Id), 302)
}

func (c *SheetController) PushPreview() {
	c.RequireRole("admin", "super_admin")
	conn := c.loadConn()
	if conn == nil {
		return
	}
	p, err := (&services.SheetSyncService{}).PushPreview(conn)
	if err != nil {
		c.FlashError("Push preview failed: " + err.Error())
		c.Redirect(fmt.Sprintf("/sheets/%d/preview", conn.Id), 302)
		return
	}
	c.Data["conn"] = conn
	c.Data["preview"] = p
	c.Data["title"] = "Push Preview"
	c.SetActivePage("sheets")
	c.GetFlashMessages()
	c.TplName = "sheets/push.html"
}

func (c *SheetController) DoPush() {
	c.RequireRole("admin", "super_admin")
	conn := c.loadConn()
	if conn == nil {
		return
	}
	res, err := (&services.SheetSyncService{}).Push(conn, c.GetCurrentUser().Id)
	if err != nil {
		c.FlashError("Push failed: " + err.Error())
		c.Redirect(fmt.Sprintf("/sheets/%d/push", conn.Id), 302)
		return
	}
	c.FlashSuccess("Push complete — " + res.Summary())
	c.Redirect(fmt.Sprintf("/sheets/%d/preview", conn.Id), 302)
}

func (c *SheetController) DoReconcile() {
	c.RequireRole("admin", "super_admin")
	conn := c.loadConn()
	if conn == nil {
		return
	}
	res, err := (&services.SheetSyncService{}).Reconcile(conn, c.GetCurrentUser().Id)
	if err != nil {
		c.FlashError("Reconcile failed: " + err.Error())
		c.Redirect(fmt.Sprintf("/sheets/%d/preview", conn.Id), 302)
		return
	}
	c.FlashSuccess("Reconcile complete — " + res.Summary())
	c.Redirect(fmt.Sprintf("/sheets/%d/preview", conn.Id), 302)
}

func (c *SheetController) Conflicts() {
	c.RequireRole("admin", "super_admin")
	conn := c.loadConn()
	if conn == nil {
		return
	}
	conflicts, _ := (&services.SheetConflictService{}).ListOpen(conn.Id)
	c.Data["conn"] = conn
	c.Data["conflicts"] = conflicts
	c.Data["title"] = "Conflicts"
	c.SetActivePage("sheets")
	c.GetFlashMessages()
	c.TplName = "sheets/conflicts.html"
}

func (c *SheetController) ResolveConflicts() {
	c.RequireRole("admin", "super_admin")
	conn := c.loadConn()
	if conn == nil {
		return
	}
	decisions := collectDecisions(c)
	if err := (&services.SheetConflictService{}).Resolve(conn.Id, decisions, c.GetCurrentUser().Id); err != nil {
		c.FlashError("Could not resolve conflicts: " + err.Error())
	} else {
		c.FlashSuccess(fmt.Sprintf("Resolved %d conflict(s)", len(decisions)))
	}
	c.Redirect(fmt.Sprintf("/sheets/%d/conflicts", conn.Id), 302)
}

func (c *SheetController) Adoptions() {
	c.RequireRole("admin", "super_admin")
	conn := c.loadConn()
	if conn == nil {
		return
	}
	adoptions, _ := (&services.SheetAdoptionService{}).ListOpenReviews(conn.Id)
	c.Data["conn"] = conn
	c.Data["adoptions"] = adoptions
	c.Data["title"] = "Adoption Review"
	c.SetActivePage("sheets")
	c.GetFlashMessages()
	c.TplName = "sheets/adoptions.html"
}

func (c *SheetController) ResolveAdoptions() {
	c.RequireRole("admin", "super_admin")
	conn := c.loadConn()
	if conn == nil {
		return
	}
	decisions := collectDecisions(c)
	if err := (&services.SheetAdoptionService{}).Resolve(conn.Id, decisions, c.GetCurrentUser().Id); err != nil {
		c.FlashError("Could not resolve adoptions: " + err.Error())
	} else {
		c.FlashSuccess(fmt.Sprintf("Resolved %d adoption(s)", len(decisions)))
	}
	c.Redirect(fmt.Sprintf("/sheets/%d/adoptions", conn.Id), 302)
}

func (c *SheetController) ToggleAutoSync() {
	c.RequireRole("admin", "super_admin")
	conn := c.loadConn()
	if conn == nil {
		return
	}
	conn.AutoSyncEnabled = !conn.AutoSyncEnabled
	(&services.SheetConnectionService{}).Save(conn)
	state := "disabled"
	if conn.AutoSyncEnabled {
		state = "enabled"
	}
	c.FlashSuccess("Auto-sync " + state)
	c.Redirect("/sheets", 302)
}

func (c *SheetController) Delete() {
	c.RequireRole("admin", "super_admin")
	conn := c.loadConn()
	if conn == nil {
		return
	}
	(&services.SheetConnectionService{}).Delete(conn.Id)
	c.FlashSuccess("Connection removed")
	c.Redirect("/sheets", 302)
}

func collectDecisions(c *SheetController) map[int]string {
	decisions := map[int]string{}
	_ = c.Ctx.Request.ParseForm()
	for key, vals := range c.Ctx.Request.Form {
		if !strings.HasPrefix(key, "decision_") || len(vals) == 0 || vals[0] == "" {
			continue
		}
		if id, err := strconv.Atoi(strings.TrimPrefix(key, "decision_")); err == nil {
			decisions[id] = vals[0]
		}
	}
	return decisions
}

func detectBrandFromName(tab string) string {
	lower := strings.ToLower(tab)
	switch {
	case strings.Contains(lower, "hp"):
		return "HP"
	case strings.Contains(lower, "lenovo"):
		return "Lenovo"
	case strings.Contains(lower, "dell"):
		return "Dell"
	default:
		return "Other"
	}
}
