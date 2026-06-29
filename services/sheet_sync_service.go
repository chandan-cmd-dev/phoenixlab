package services

import (
	"PhoenixLab/models"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

const syncUIDHeader = "PL_SYNC_UID"

type SheetSyncService struct{}

type syncContext struct {
	conn        *models.SheetConnection
	client      *SheetsClient
	identity    []string
	headers     []string
	dataRows    [][]string
	dataStart   int
	fieldCols   map[int]string
	customCols  map[int]string
	transforms  map[int]string
	identityCol map[string]int
	uidCol      int
}

type rowAction struct {
	rowIndex   int
	sheetRow   []string
	naturalKey string
	stampedUid string
	uid        string
	action     string
	link       *models.SheetRowLink
	ticketID   int
	candidates []int
}

type SyncPreview struct {
	Total     int
	Creates   int
	Adopts    int
	Updates   int
	Ambiguous int
	Skipped   int
}

type SyncResult struct {
	Total        int
	Created      int
	Updated      int
	Adopted      int
	Ambiguous    int
	Skipped      int
	Pushed       int
	Conflicts    int
	PulledFields int
	PushedFields int
	Errors       []string
}

func (r *SyncResult) Summary() string {
	return fmt.Sprintf("created=%d updated=%d adopted=%d ambiguous=%d conflicts=%d pushed=%d skipped=%d",
		r.Created, r.Updated, r.Adopted, r.Ambiguous, r.Conflicts, r.Pushed, r.Skipped)
}

func (s *SheetSyncService) loadContext(conn *models.SheetConnection) (*syncContext, error) {
	if conn.TabName == "" {
		return nil, fmt.Errorf("connection has no tab selected")
	}
	client, err := NewSheetsClient()
	if err != nil {
		return nil, err
	}
	rows, err := client.GetRows(conn.SpreadsheetId, conn.TabName)
	if err != nil {
		return nil, err
	}
	if conn.HeaderRow >= len(rows) {
		return nil, fmt.Errorf("header row %d is beyond the sheet length (%d rows)", conn.HeaderRow, len(rows))
	}

	mappingSvc := SheetMappingService{}
	mappings, err := mappingSvc.LoadMappings(conn.Id)
	if err != nil {
		return nil, err
	}
	if len(mappings) == 0 {
		return nil, fmt.Errorf("connection has no column mapping; configure mapping first")
	}

	ctx := &syncContext{
		conn:        conn,
		client:      client,
		identity:    conn.IdentityFields(),
		headers:     rows[conn.HeaderRow],
		dataStart:   conn.HeaderRow + 1,
		fieldCols:   map[int]string{},
		customCols:  map[int]string{},
		transforms:  map[int]string{},
		identityCol: map[string]int{},
		uidCol:      -1,
	}
	if conn.HeaderRow+1 < len(rows) {
		ctx.dataRows = rows[conn.HeaderRow+1:]
	}

	for _, m := range mappings {
		if m.IsIgnored() {
			continue
		}
		ctx.transforms[m.ColumnIndex] = m.Transform
		if m.IsCustom() {
			ctx.customCols[m.ColumnIndex] = m.CustomKey()
		} else {
			ctx.fieldCols[m.ColumnIndex] = m.TargetField
			if _, taken := ctx.identityCol[m.TargetField]; !taken {
				ctx.identityCol[m.TargetField] = m.ColumnIndex
			}
		}
	}

	for i, h := range ctx.headers {
		if strings.TrimSpace(h) == syncUIDHeader {
			ctx.uidCol = i
			break
		}
	}
	return ctx, nil
}

func (ctx *syncContext) naturalKey(row []string) string {
	parts := make([]string, 0, len(ctx.identity))
	for _, field := range ctx.identity {
		val := ""
		if col, ok := ctx.identityCol[field]; ok {
			val = strings.TrimSpace(getCellValue(row, col))
		}
		parts = append(parts, val)
	}
	return strings.Join(parts, "||")
}

func (ctx *syncContext) serialValue(row []string) string {
	if col, ok := ctx.identityCol["SerialNumber"]; ok {
		return strings.TrimSpace(getCellValue(row, col))
	}
	return strings.TrimSpace(strings.ReplaceAll(ctx.naturalKey(row), "||", ""))
}

func (s *SheetSyncService) classifyRow(o orm.Ormer, ctx *syncContext, rowIdx int) rowAction {
	row := ctx.dataRows[rowIdx]
	act := rowAction{rowIndex: rowIdx, sheetRow: row}

	if ctx.serialValue(row) == "" {
		act.action = "skip"
		return act
	}
	act.naturalKey = ctx.naturalKey(row)
	if ctx.uidCol >= 0 {
		act.stampedUid = strings.TrimSpace(getCellValue(row, ctx.uidCol))
	}
	act.uid = act.naturalKey
	if act.stampedUid != "" {
		act.uid = act.stampedUid
	}

	if act.stampedUid != "" {
		var link models.SheetRowLink
		if err := o.QueryTable("sheet_row_links").
			Filter("connection_id", ctx.conn.Id).
			Filter("stamped_uid", act.stampedUid).One(&link); err == nil {
			act.action = "update"
			act.link = &link
			act.ticketID = link.TicketId
			return act
		}
	}
	var link models.SheetRowLink
	if err := o.QueryTable("sheet_row_links").
		Filter("connection_id", ctx.conn.Id).
		Filter("sheet_row_uid", act.uid).One(&link); err == nil {
		act.action = "update"
		act.link = &link
		act.ticketID = link.TicketId
		return act
	}

	ids := s.matchTicketsByIdentity(o, ctx, row)
	switch len(ids) {
	case 0:
		act.action = "create"
	case 1:
		act.action = "adopt"
		act.ticketID = ids[0]
	default:
		act.action = "ambiguous"
		act.candidates = ids
	}
	return act
}

func (s *SheetSyncService) matchTicketsByIdentity(o orm.Ormer, ctx *syncContext, row []string) []int {
	qs := o.QueryTable("tickets")
	if ctx.conn.BranchId > 0 {
		qs = qs.Filter("branch_id", ctx.conn.BranchId)
	}
	matched := 0
	for _, field := range ctx.identity {
		col, ok := ctx.identityCol[field]
		if !ok {
			continue
		}
		qs = qs.Filter(field, strings.TrimSpace(getCellValue(row, col)))
		matched++
	}
	if matched == 0 {
		return nil
	}
	var tickets []*models.Ticket
	qs.All(&tickets, "Id")
	ids := make([]int, 0, len(tickets))
	for _, t := range tickets {
		ids = append(ids, t.Id)
	}
	return ids
}

func (s *SheetSyncService) Preview(conn *models.SheetConnection) (*SyncPreview, error) {
	ctx, err := s.loadContext(conn)
	if err != nil {
		return nil, err
	}
	o := orm.NewOrm()
	p := &SyncPreview{}
	for i := range ctx.dataRows {
		act := s.classifyRow(o, ctx, i)
		p.Total++
		switch act.action {
		case "create":
			p.Creates++
		case "adopt":
			p.Adopts++
		case "update":
			p.Updates++
		case "ambiguous":
			p.Ambiguous++
		default:
			p.Skipped++
		}
	}
	return p, nil
}

func (s *SheetSyncService) Import(conn *models.SheetConnection, userID int) (*SyncResult, error) {
	ctx, err := s.loadContext(conn)
	if err != nil {
		return nil, err
	}
	o := orm.NewOrm()
	audit := AuditService{}
	res := &SyncResult{}

	for i := range ctx.dataRows {
		act := s.classifyRow(o, ctx, i)
		res.Total++
		switch act.action {
		case "skip":
			res.Skipped++
		case "ambiguous":
			s.recordAdoption(o, ctx, act)
			res.Ambiguous++
		case "create":
			if err := s.applyCreate(o, ctx, act, userID, &audit); err != nil {
				res.Errors = append(res.Errors, fmt.Sprintf("row %d: %v", ctx.dataStart+i+1, err))
				res.Skipped++
				continue
			}
			res.Created++
		case "adopt", "update":
			updated, err := s.applyToTicket(o, ctx, act, userID, &audit)
			if err != nil {
				res.Errors = append(res.Errors, fmt.Sprintf("row %d: %v", ctx.dataStart+i+1, err))
				res.Skipped++
				continue
			}
			if act.action == "adopt" {
				res.Adopted++
			} else {
				res.Updated++
			}
			_ = updated
		}
	}

	conn.LastSyncedAt = time.Now()
	(&SheetConnectionService{}).Save(conn)
	return res, nil
}

func (s *SheetSyncService) applyCreate(o orm.Ormer, ctx *syncContext, act rowAction, userID int, audit *AuditService) error {
	t := &models.Ticket{
		BranchId:       ctx.conn.BranchId,
		Brand:          ctx.conn.Brand,
		CreatedBy:      userID,
		WarrantyStatus: "in_warranty",
		Priority:       "normal",
		Status:         "open",
		ReceivedAt:     time.Now(),
	}
	s.applyFields(ctx, act.sheetRow, t)
	if strings.TrimSpace(t.SerialNumber) == "" {
		return fmt.Errorf("missing serial number")
	}
	if _, err := o.Insert(t); err != nil {
		return err
	}
	o.Raw("UPDATE tickets SET assigned_to = NULL WHERE id = ?", t.Id).Exec()
	audit.Log("ticket", t.Id, "create", "", "", "Imported from Google Sheet: "+ctx.conn.TabName, userID, "")
	s.upsertLink(o, ctx, act, t)
	return nil
}

func (s *SheetSyncService) applyToTicket(o orm.Ormer, ctx *syncContext, act rowAction, userID int, audit *AuditService) (bool, error) {
	t := &models.Ticket{Id: act.ticketID}
	if err := o.Read(t); err != nil {
		return false, fmt.Errorf("linked ticket %d not found: %w", act.ticketID, err)
	}
	s.applyFields(ctx, act.sheetRow, t)
	if _, err := o.Update(t); err != nil {
		return false, err
	}
	verb := "update"
	msg := "Updated from Google Sheet: " + ctx.conn.TabName
	if act.action == "adopt" {
		msg = "Adopted from Google Sheet: " + ctx.conn.TabName
	}
	audit.Log("ticket", t.Id, verb, "", "", msg, userID, "")
	s.upsertLink(o, ctx, act, t)
	return true, nil
}

func (s *SheetSyncService) applyFields(ctx *syncContext, row []string, t *models.Ticket) {
	for col, field := range ctx.fieldCols {
		setTicketField(t, field, getCellValue(row, col), ctx.transforms[col])
	}
	for col, key := range ctx.customCols {
		if raw := strings.TrimSpace(getCellValue(row, col)); raw != "" {
			t.SetCustomField(key, raw)
		}
	}
}

func (s *SheetSyncService) upsertLink(o orm.Ormer, ctx *syncContext, act rowAction, t *models.Ticket) {
	snap := ctx.canonicalSnapshot(t)
	snapJSON, _ := json.Marshal(snap)
	link := act.link
	if link == nil {
		link = &models.SheetRowLink{
			ConnectionId: ctx.conn.Id,
			SheetRowUid:  act.uid,
		}
	}
	link.TicketId = t.Id
	link.BaselineSnapshot = string(snapJSON)
	link.ContentHash = hashCanonical(snap)
	if act.stampedUid != "" {
		link.StampedUid = act.stampedUid
	}
	link.LastPulledAt = time.Now()
	if link.Id == 0 {
		o.Insert(link)
	} else {
		o.Update(link)
	}
}

func (s *SheetSyncService) recordAdoption(o orm.Ormer, ctx *syncContext, act rowAction) {
	data := map[string]string{}
	for col, field := range ctx.fieldCols {
		data[field] = getCellValue(act.sheetRow, col)
	}
	for col, key := range ctx.customCols {
		data[models.CustomFieldPrefix+key] = getCellValue(act.sheetRow, col)
	}
	dataJSON, _ := json.Marshal(data)
	candIDs := joinInts(act.candidates)

	var existing models.SheetAdoption
	err := o.QueryTable("sheet_adoptions").
		Filter("connection_id", ctx.conn.Id).
		Filter("sheet_row_uid", act.uid).
		Filter("status", "open").One(&existing)
	if err == nil {
		existing.RowDataJson = string(dataJSON)
		existing.CandidateIds = candIDs
		existing.NaturalKey = act.naturalKey
		o.Update(&existing)
		return
	}
	o.Insert(&models.SheetAdoption{
		ConnectionId: ctx.conn.Id,
		SheetRowUid:  act.uid,
		NaturalKey:   act.naturalKey,
		RowDataJson:  string(dataJSON),
		CandidateIds: candIDs,
		Status:       "open",
	})
}

func (s *SheetSyncService) PushPreview(conn *models.SheetConnection) (*SyncPreview, error) {
	ctx, err := s.loadContext(conn)
	if err != nil {
		return nil, err
	}
	o := orm.NewOrm()
	p := &SyncPreview{}
	for i := range ctx.dataRows {
		p.Total++
		if ctx.serialValue(ctx.dataRows[i]) == "" {
			p.Skipped++
			continue
		}
		if s.resolveLink(o, ctx, ctx.dataRows[i]) == nil {
			p.Skipped++
		} else {
			p.Updates++
		}
	}
	return p, nil
}

func (s *SheetSyncService) Push(conn *models.SheetConnection, userID int) (*SyncResult, error) {
	if conn.SyncDirection == "pull" {
		return nil, fmt.Errorf("connection is pull-only; push is disabled")
	}
	ctx, err := s.loadContext(conn)
	if err != nil {
		return nil, err
	}
	o := orm.NewOrm()
	res := &SyncResult{}
	var updates []CellUpdate

	uidCol := ctx.uidCol
	if uidCol < 0 {
		uidCol = len(ctx.headers)
		updates = append(updates, CellUpdate{Row: conn.HeaderRow, Col: uidCol, Value: syncUIDHeader})
	}

	for i := range ctx.dataRows {
		row := ctx.dataRows[i]
		res.Total++
		sheetRowNum := ctx.dataStart + i
		link := s.resolveLink(o, ctx, row)
		if link == nil {
			res.Skipped++
			continue
		}
		t := &models.Ticket{Id: link.TicketId}
		if err := o.Read(t); err != nil {
			res.Skipped++
			continue
		}

		updates = append(updates, CellUpdate{Row: sheetRowNum, Col: uidCol, Value: link.StampValue()})
		for col, field := range ctx.fieldCols {
			updates = append(updates, CellUpdate{Row: sheetRowNum, Col: col, Value: formatFieldValue(t, field)})
		}
		custom := t.GetCustomFields()
		for col, key := range ctx.customCols {
			updates = append(updates, CellUpdate{Row: sheetRowNum, Col: col, Value: custom[key]})
		}

		snap := ctx.canonicalSnapshot(t)
		snapJSON, _ := json.Marshal(snap)
		link.BaselineSnapshot = string(snapJSON)
		link.ContentHash = hashCanonical(snap)
		link.StampedUid = link.StampValue()
		link.LastPushedAt = time.Now()
		o.Update(link)

		res.Pushed++
		res.PushedFields += len(ctx.fieldCols) + len(ctx.customCols)
	}

	if err := ctx.client.WriteCells(conn.SpreadsheetId, conn.TabName, updates); err != nil {
		return nil, fmt.Errorf("sheet write failed: %w", err)
	}
	conn.LastPushedAt = time.Now()
	(&SheetConnectionService{}).Save(conn)
	return res, nil
}

func (s *SheetSyncService) Reconcile(conn *models.SheetConnection, userID int) (*SyncResult, error) {
	if conn.SyncDirection != "two_way" {
		return nil, fmt.Errorf("reconcile requires a two-way connection")
	}
	ctx, err := s.loadContext(conn)
	if err != nil {
		return nil, err
	}
	o := orm.NewOrm()
	audit := AuditService{}
	res := &SyncResult{}
	var sheetUpdates []CellUpdate

	for i := range ctx.dataRows {
		row := ctx.dataRows[i]
		res.Total++
		sheetRowNum := ctx.dataStart + i
		link := s.resolveLink(o, ctx, row)
		if link == nil {
			res.Skipped++
			continue
		}
		t := &models.Ticket{Id: link.TicketId}
		if err := o.Read(t); err != nil {
			res.Skipped++
			continue
		}

		baseline := decodeSnapshot(link.BaselineSnapshot)
		newBaseline := map[string]string{}
		changed := false

		for col, field := range ctx.fieldCols {
			sheetVal := canonicalValue(getCellValue(row, col), ctx.transforms[col])
			dbVal := fieldCanonical(t, field)
			switch classifyField(baseline[field], sheetVal, dbVal) {
			case "sheetWins":
				setTicketField(t, field, getCellValue(row, col), ctx.transforms[col])
				changed = true
				newBaseline[field] = sheetVal
				res.PulledFields++
			case "dbWins":
				sheetUpdates = append(sheetUpdates, CellUpdate{Row: sheetRowNum, Col: col, Value: formatFieldValue(t, field)})
				newBaseline[field] = dbVal
				res.PushedFields++
			case "conflict":
				s.recordConflict(o, ctx, link, field, baseline[field], sheetVal, dbVal)
				newBaseline[field] = baseline[field]
				res.Conflicts++
			default:
				newBaseline[field] = dbVal
			}
		}

		custom := t.GetCustomFields()
		for col, key := range ctx.customCols {
			snapKey := models.CustomFieldPrefix + key
			sheetVal := canonicalValue(getCellValue(row, col), "text")
			dbVal := canonicalValue(custom[key], "text")
			switch classifyField(baseline[snapKey], sheetVal, dbVal) {
			case "sheetWins":
				t.SetCustomField(key, strings.TrimSpace(getCellValue(row, col)))
				changed = true
				newBaseline[snapKey] = sheetVal
				res.PulledFields++
			case "dbWins":
				sheetUpdates = append(sheetUpdates, CellUpdate{Row: sheetRowNum, Col: col, Value: custom[key]})
				newBaseline[snapKey] = dbVal
				res.PushedFields++
			case "conflict":
				s.recordConflict(o, ctx, link, snapKey, baseline[snapKey], sheetVal, dbVal)
				newBaseline[snapKey] = baseline[snapKey]
				res.Conflicts++
			default:
				newBaseline[snapKey] = dbVal
			}
		}

		if changed {
			o.Update(t)
			audit.Log("ticket", t.Id, "update", "", "", "Reconciled from Google Sheet: "+ctx.conn.TabName, userID, "")
		}
		snapJSON, _ := json.Marshal(newBaseline)
		link.BaselineSnapshot = string(snapJSON)
		link.ContentHash = hashCanonical(newBaseline)
		link.LastPulledAt = time.Now()
		o.Update(link)
		res.Updated++
	}

	if err := ctx.client.WriteCells(conn.SpreadsheetId, conn.TabName, sheetUpdates); err != nil {
		return nil, fmt.Errorf("sheet write failed: %w", err)
	}
	now := time.Now()
	conn.LastSyncedAt = now
	conn.LastPushedAt = now
	(&SheetConnectionService{}).Save(conn)
	return res, nil
}

func (s *SheetSyncService) resolveLink(o orm.Ormer, ctx *syncContext, row []string) *models.SheetRowLink {
	if ctx.uidCol >= 0 {
		if stamp := strings.TrimSpace(getCellValue(row, ctx.uidCol)); stamp != "" {
			var l models.SheetRowLink
			if err := o.QueryTable("sheet_row_links").
				Filter("connection_id", ctx.conn.Id).
				Filter("stamped_uid", stamp).One(&l); err == nil {
				return &l
			}
		}
	}
	var l models.SheetRowLink
	if err := o.QueryTable("sheet_row_links").
		Filter("connection_id", ctx.conn.Id).
		Filter("sheet_row_uid", ctx.naturalKey(row)).One(&l); err == nil {
		return &l
	}
	return nil
}

func (s *SheetSyncService) recordConflict(o orm.Ormer, ctx *syncContext, link *models.SheetRowLink, field, baseVal, sheetVal, dbVal string) {
	var existing models.SheetConflict
	err := o.QueryTable("sheet_conflicts").
		Filter("link_id", link.Id).
		Filter("field_name", field).
		Filter("status", "open").One(&existing)
	if err == nil {
		existing.SheetValue = sheetVal
		existing.DbValue = dbVal
		existing.BaselineValue = baseVal
		o.Update(&existing)
		return
	}
	o.Insert(&models.SheetConflict{
		ConnectionId:  ctx.conn.Id,
		LinkId:        link.Id,
		TicketId:      link.TicketId,
		FieldName:     field,
		BaselineValue: baseVal,
		SheetValue:    sheetVal,
		DbValue:       dbVal,
		Status:        "open",
	})
}

func (ctx *syncContext) canonicalSnapshot(t *models.Ticket) map[string]string {
	snap := map[string]string{}
	for _, field := range ctx.fieldCols {
		snap[field] = fieldCanonical(t, field)
	}
	custom := t.GetCustomFields()
	for _, key := range ctx.customCols {
		snap[models.CustomFieldPrefix+key] = canonicalValue(custom[key], "text")
	}
	return snap
}

func canonicalValue(raw, transform string) string {
	raw = strings.TrimSpace(raw)
	switch transform {
	case "currency":
		return strconv.FormatFloat(parseCurrency(raw), 'f', 2, 64)
	case "bool":
		if isYes(raw) {
			return "true"
		}
		return "false"
	case "date":
		if tm, err := parseFlexDate(raw); err == nil {
			return tm.Format("2006-01-02")
		}
		return ""
	case "status":
		return mapExcelStatus(raw)
	default:
		return raw
	}
}

func fieldCanonical(t *models.Ticket, field string) string {
	v := reflect.ValueOf(t).Elem().FieldByName(field)
	if !v.IsValid() {
		return ""
	}
	switch v.Kind() {
	case reflect.String:
		return strings.TrimSpace(v.String())
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', 2, 64)
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	case reflect.Int, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	default:
		if tm, ok := v.Interface().(time.Time); ok {
			if tm.IsZero() {
				return ""
			}
			return tm.Format("2006-01-02")
		}
	}
	return ""
}

func hashCanonical(snap map[string]string) string {
	keys := make([]string, 0, len(snap))
	for k := range snap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(snap[k])
		b.WriteByte(';')
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func classifyField(baseline, sheet, db string) string {
	sheetChanged := sheet != baseline
	dbChanged := db != baseline
	switch {
	case !sheetChanged && !dbChanged:
		return "agree"
	case sheetChanged && !dbChanged:
		return "sheetWins"
	case !sheetChanged && dbChanged:
		return "dbWins"
	default:
		if sheet == db {
			return "agree"
		}
		return "conflict"
	}
}

func setTicketField(t *models.Ticket, field, raw, transform string) {
	v := reflect.ValueOf(t).Elem().FieldByName(field)
	if !v.IsValid() || !v.CanSet() {
		return
	}
	switch v.Kind() {
	case reflect.String:
		val := strings.TrimSpace(raw)
		if field == "Status" {
			val = mapExcelStatus(val)
		}
		v.SetString(val)
	case reflect.Float64:
		v.SetFloat(parseCurrency(raw))
	case reflect.Bool:
		v.SetBool(isYes(raw))
	default:
		if v.Type() == reflect.TypeOf(time.Time{}) {
			if tm, err := parseFlexDate(raw); err == nil {
				v.Set(reflect.ValueOf(tm))
			}
		}
	}
}

func formatFieldValue(t *models.Ticket, field string) string {
	v := reflect.ValueOf(t).Elem().FieldByName(field)
	if !v.IsValid() {
		return ""
	}
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Float64:
		f := v.Float()
		if f == 0 {
			return ""
		}
		return strconv.FormatFloat(f, 'f', 2, 64)
	case reflect.Bool:
		if v.Bool() {
			return "Yes"
		}
		return "No"
	case reflect.Int, reflect.Int64:
		n := v.Int()
		if n == 0 {
			return ""
		}
		return strconv.FormatInt(n, 10)
	default:
		if tm, ok := v.Interface().(time.Time); ok {
			if tm.IsZero() {
				return ""
			}
			return tm.Format("2006-01-02")
		}
	}
	return ""
}

func applyStashedData(t *models.Ticket, data map[string]string) {
	for k, raw := range data {
		if strings.HasPrefix(k, models.CustomFieldPrefix) {
			if val := strings.TrimSpace(raw); val != "" {
				t.SetCustomField(strings.TrimPrefix(k, models.CustomFieldPrefix), val)
			}
			continue
		}
		setTicketField(t, k, raw, transformFor(k))
	}
}

func decodeSnapshot(s string) map[string]string {
	m := map[string]string{}
	if s == "" {
		return m
	}
	_ = json.Unmarshal([]byte(s), &m)
	if m == nil {
		m = map[string]string{}
	}
	return m
}

func joinInts(ints []int) string {
	parts := make([]string, len(ints))
	for i, n := range ints {
		parts[i] = strconv.Itoa(n)
	}
	return strings.Join(parts, ",")
}

func splitInts(s string) []int {
	var out []int
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if n, err := strconv.Atoi(p); err == nil {
			out = append(out, n)
		}
	}
	return out
}
