package controllers

import (
	"PhoenixLab/models"
	"PhoenixLab/services"
	"os"
	"path/filepath"
	"time"
)

type ImportController struct {
	BaseController
}

func (c *ImportController) ShowImport() {
	c.RequireRole("admin", "super_admin")

	var branches []*models.Branch
	if c.GetCurrentUser().IsAdmin() {
		branches, _ = models.GetAllBranches()
	}

	c.Data["branches"] = branches
	c.Data["title"] = "Import Excel Data"
	c.SetActivePage("import")
	c.GetFlashMessages()
	c.TplName = "import/index.html"
}

func (c *ImportController) DoImport() {
	c.RequireRole("admin", "super_admin")

	file, header, err := c.GetFile("excel_file")
	if err != nil {
		c.FlashError("Please select an Excel file to upload")
		c.Redirect("/import", 302)
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	if ext != ".xlsx" && ext != ".xls" {
		c.FlashError("Only .xlsx files are supported")
		c.Redirect("/import", 302)
		return
	}

	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "import_"+time.Now().Format("20060102150405")+ext)
	if err := c.SaveToFile("excel_file", tmpFile); err != nil {
		c.FlashError("Failed to save uploaded file: " + err.Error())
		c.Redirect("/import", 302)
		return
	}
	defer os.Remove(tmpFile)

	branchID := c.GetCurrentUser().BranchId
	if c.GetCurrentUser().IsSuperAdmin() {
		if bid, err := c.GetInt("branch_id"); err == nil && bid > 0 {
			branchID = bid
		}
	}

	if branchID == 0 {
		c.FlashError("Please select a branch for the imported tickets")
		c.Redirect("/import", 302)
		return
	}

	importService := services.ImportService{}
	results, err := importService.ImportExcel(tmpFile, branchID, c.GetCurrentUser().Id)
	if err != nil {
		c.FlashError("Import failed: " + err.Error())
		c.Redirect("/import", 302)
		return
	}

	c.Data["results"] = results
	c.Data["title"] = "Import Results"
	c.SetActivePage("import")
	c.GetFlashMessages()
	c.TplName = "import/results.html"
}
