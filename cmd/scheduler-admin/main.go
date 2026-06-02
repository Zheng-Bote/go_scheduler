/**
 * SPDX-FileComment: Scheduler Admin
 * SPDX-FileType: SOURCE
 * SPDX-FileContributor: ZHENG Robert
 * SPDX-FileCopyrightText: 2026 ZHENG Robert
 * SPDX-License-Identifier: Apache-2.0
 *
 * @file main.go
 * @brief Cross-platform GUI admin tool for job management via Fyne
 * @version 1.0.0
 * @date 2026-06-02
 *
 * @author ZHENG Robert (robert@hase-zheng.net)
 * @copyright Copyright (c) 2026 ZHENG Robert
 * @LICENSE Apache-2.0
 */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/user"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	fynetooltip "github.com/dweymouth/fyne-tooltip"
	ttwidget "github.com/dweymouth/fyne-tooltip/widget"
)

// ScheduledProgram mirrors the server-side database structure for job
// configuration. It is used to decode the JSON response from the scheduler
// admin API and to construct payloads for create/update requests.
type ScheduledProgram struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	Command       string   `json:"command"`
	Args          []string `json:"args"`
	CronExpr      string   `json:"cron_expr"`
	Enabled       bool     `json:"enabled"`
	RestartOnExit bool     `json:"restart_on_exit"`
}

// AdminUI holds the state and Fyne widgets for the scheduler administration
// GUI. It manages the connection form, job table, and the list of programs
// retrieved from the scheduler API.
type AdminUI struct {
	Window   fyne.Window
	URL      *widget.Entry
	User     *widget.Entry
	Token    *widget.Entry
	JobTable *widget.Table
	Programs []ScheduledProgram
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Scheduler Admin")
	myWindow.Resize(fyne.NewSize(800, 600))

	ui := &AdminUI{
		Window: myWindow,
		URL:    widget.NewEntry(),
		User:   widget.NewEntry(),
		Token:  widget.NewPasswordEntry(),
	}

	ui.URL.SetPlaceHolder("http://localhost:8080")
	ui.User.SetText(getHeloUser())
	ui.User.Disable()
	ui.Token.SetPlaceHolder("auth token")

	// Form for connection settings
	connForm := widget.NewForm(
		widget.NewFormItem("Scheduler URL", ui.URL),
		widget.NewFormItem("Admin User", ui.User),
		widget.NewFormItem("Auth Token", ui.Token),
	)

	// Table to display programs
	ui.JobTable = widget.NewTable(
		func() (int, int) {
			return len(ui.Programs) + 1, 2
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Wide Job Name Placeholder")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id.Row == 0 {
				if id.Col == 0 {
					label.SetText("Job Name")
					label.TextStyle.Bold = true
				} else {
					label.SetText("Status")
					label.TextStyle.Bold = true
				}
				return
			}

			label.TextStyle.Bold = false
			p := ui.Programs[id.Row-1]
			if id.Col == 0 {
				label.SetText(p.Name)
			} else {
				status := "Enabled"
				if !p.Enabled {
					status = "Disabled"
				}
				label.SetText(status)
			}
		},
	)

	// Adjust column widths
	ui.JobTable.SetColumnWidth(0, 400)
	ui.JobTable.SetColumnWidth(1, 150)

	ui.JobTable.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 {
			ui.showJobEditor(ui.Programs[id.Row-1])
		}
	}

	// Toolbar with tooltips using fyne-tooltip library
	toolbar := widget.NewToolbar(
		&toolbarTooltipAction{theme.ListIcon(), "load Jobs", func() { ui.loadJobs() }},
		&toolbarTooltipAction{theme.ContentAddIcon(), "new Job", func() { ui.showJobEditor(ScheduledProgram{Enabled: true}) }},
		widget.NewToolbarSeparator(),
		&toolbarTooltipAction{theme.FileIcon(), "download System-Logs", func() { ui.downloadSystemLogs() }},
		&toolbarTooltipAction{theme.FileTextIcon(), "download Job-Audit-Logs", func() { ui.downloadJobAuditLogs() }},
		&toolbarTooltipAction{theme.FileApplicationIcon(), "download Admin-Audit-Logs", func() { ui.downloadAdminAuditLogs() }},
	)

	topContainer := container.NewVBox(connForm, toolbar)
	content := container.NewBorder(topContainer, nil, nil, nil, ui.JobTable)

	// Wrap content in tooltip layer
	myWindow.SetContent(fynetooltip.AddWindowToolTipLayer(content, myWindow.Canvas()))
	myWindow.ShowAndRun()
}

// getHeloUser returns the current OS username stripped of any domain prefix
// (e.g. "DOMAIN\Username" becomes "Username"). The result is used as the
// default admin user for HTTP basic authentication.
func getHeloUser() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	// On Windows, Username can be in the format "DOMAIN\Username". We want just "Username".
	parts := strings.Split(u.Username, "\\")
	return parts[len(parts)-1]
}

// loadJobs fetches all scheduled programs from the scheduler admin API and
// refreshes the job table. It first runs Windows Hello verification (if
// available) before making the HTTP request.
func (ui *AdminUI) loadJobs() {
	go func() {
		verified, err := verifyWindowsHello(ui.Window)
		if err != nil || !verified {
			if err != nil {
				dialog.ShowError(err, ui.Window)
			}
			return
		}

		client := &http.Client{}
		req, err := http.NewRequest("GET", ui.URL.Text+"/admin/jobs", nil)
		if err != nil {
			dialog.ShowError(err, ui.Window)
			return
		}
		req.SetBasicAuth(ui.User.Text, ui.Token.Text)

		resp, err := client.Do(req)
		if err != nil {
			dialog.ShowError(err, ui.Window)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			dialog.ShowInformation("Error", fmt.Sprintf("Failed to load jobs: %s", string(body)), ui.Window)
			return
		}

		var programs []ScheduledProgram
		if err := json.NewDecoder(resp.Body).Decode(&programs); err != nil {
			dialog.ShowError(err, ui.Window)
			return
		}
		ui.Programs = programs
		ui.JobTable.Refresh()
	}()
}

// showJobEditor opens a modal dialog for creating or editing a scheduled
// program.
func (ui *AdminUI) showJobEditor(p ScheduledProgram) {
	nameEntry := widget.NewEntry()
	nameEntry.SetText(p.Name)
	cmdEntry := widget.NewEntry()
	cmdEntry.SetText(p.Command)
	argsEntry := widget.NewEntry()
	argsEntry.SetText(strings.Join(p.Args, ","))
	cronEntry := widget.NewEntry()
	cronEntry.SetText(p.CronExpr)
	enabledCheck := widget.NewCheck("Enabled", func(bool) {})
	enabledCheck.Checked = p.Enabled
	restartCheck := widget.NewCheck("Restart on Exit", func(bool) {})
	restartCheck.Checked = p.RestartOnExit

	form := widget.NewForm(
		widget.NewFormItem("Job Name", nameEntry),
		widget.NewFormItem("Command", cmdEntry),
		widget.NewFormItem("Args (comma sep)", argsEntry),
		widget.NewFormItem("Cron Expr", cronEntry),
		widget.NewFormItem("", enabledCheck),
		widget.NewFormItem("", restartCheck),
	)

	var items []fyne.CanvasObject
	items = append(items, form)
	if p.Name != "" {
		delBtn := widget.NewButton("Delete Job", func() {
			dialog.ShowConfirm("Confirm", "Delete this job?", func(ok bool) {
				if ok {
					ui.deleteJob(p.Name)
				}
			}, ui.Window)
		})
		items = append(items, delBtn)
	}

	dialog.ShowCustomConfirm("Edit Job", "Save", "Cancel", container.NewVBox(items...), func(save bool) {
		if !save {
			return
		}
		p.Name = nameEntry.Text
		p.Command = cmdEntry.Text
		if argsEntry.Text != "" {
			p.Args = strings.Split(argsEntry.Text, ",")
		} else {
			p.Args = []string{}
		}
		p.CronExpr = cronEntry.Text
		p.Enabled = enabledCheck.Checked
		p.RestartOnExit = restartCheck.Checked
		ui.saveJob(p)
	}, ui.Window)
}

func (ui *AdminUI) saveJob(p ScheduledProgram) {
	data, _ := json.Marshal([]ScheduledProgram{p})
	client := &http.Client{}
	req, err := http.NewRequest("POST", ui.URL.Text+"/admin/update-jobs", bytes.NewBuffer(data))
	if err != nil {
		dialog.ShowError(err, ui.Window)
		return
	}
	req.SetBasicAuth(ui.User.Text, ui.Token.Text)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		dialog.ShowError(err, ui.Window)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		dialog.ShowInformation("Success", "Job saved", ui.Window)
		ui.loadJobs()
	} else {
		body, _ := io.ReadAll(resp.Body)
		dialog.ShowInformation("Error", string(body), ui.Window)
	}
}

func (ui *AdminUI) deleteJob(name string) {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", ui.URL.Text+"/admin/delete-job?name="+name, nil)
	if err != nil {
		dialog.ShowError(err, ui.Window)
		return
	}
	req.SetBasicAuth(ui.User.Text, ui.Token.Text)
	resp, err := client.Do(req)
	if err != nil {
		dialog.ShowError(err, ui.Window)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		dialog.ShowInformation("Success", "Job deleted", ui.Window)
		ui.loadJobs()
	} else {
		body, _ := io.ReadAll(resp.Body)
		dialog.ShowInformation("Error", string(body), ui.Window)
	}
}

func (ui *AdminUI) downloadSystemLogs() {
	ui.downloadLogs("/admin/logs/system", "system_logs.json", "System logs saved")
}

func (ui *AdminUI) downloadJobAuditLogs() {
	ui.downloadLogs("/admin/logs/job-audit", "job_audit_logs.json", "Job audit logs saved")
}

func (ui *AdminUI) downloadAdminAuditLogs() {
	ui.downloadLogs("/admin/logs/admin-audit", "admin_audit_logs.json", "Admin audit logs saved")
}

func (ui *AdminUI) downloadLogs(endpoint, fileName, successMsg string) {
	go func() {
		verified, err := verifyWindowsHello(ui.Window)
		if err != nil || !verified {
			if err != nil {
				dialog.ShowError(err, ui.Window)
			}
			return
		}
		client := &http.Client{}
		req, err := http.NewRequest("GET", ui.URL.Text+endpoint, nil)
		if err != nil {
			dialog.ShowError(err, ui.Window)
			return
		}
		req.SetBasicAuth(ui.User.Text, ui.Token.Text)
		resp, err := client.Do(req)
		if err != nil {
			dialog.ShowError(err, ui.Window)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			dialog.ShowInformation("Error", fmt.Sprintf("Failed to load logs: %s", string(body)), ui.Window)
			return
		}
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			dialog.ShowError(err, ui.Window)
			return
		}
		saveDlg := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil || writer == nil {
				return
			}
			defer writer.Close()
			_, _ = writer.Write(data)
			dialog.ShowInformation("Success", successMsg, ui.Window)
		}, ui.Window)
		saveDlg.SetFileName(fileName)
		saveDlg.Show()
	}()
}

// toolbarTooltipAction implements widget.ToolbarItem and creates a button
// with a tooltip for use in a Fyne toolbar using the fyne-tooltip library.
type toolbarTooltipAction struct {
	icon    fyne.Resource
	tooltip string
	action  func()
}

func (t *toolbarTooltipAction) ToolbarObject() fyne.CanvasObject {
	b := ttwidget.NewButtonWithIcon("", t.icon, t.action)
	b.Importance = widget.LowImportance
	b.SetToolTip(t.tooltip)
	return b
}
