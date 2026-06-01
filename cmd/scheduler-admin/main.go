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
	"fyne.io/fyne/v2/widget"
)

// ScheduledProgram matches the server-side DB structure
type ScheduledProgram struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	Command       string   `json:"command"`
	Args          []string `json:"args"`
	CronExpr      string   `json:"cron_expr"`
	Enabled       bool     `json:"enabled"`
	RestartOnExit bool     `json:"restart_on_exit"`
}

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

	// Buttons
	loadBtn := widget.NewButton("Load Jobs", func() { ui.loadJobs() })
	newBtn := widget.NewButton("New Job", func() { ui.showJobEditor(ScheduledProgram{Enabled: true}) })

	topContainer := container.NewVBox(connForm, container.NewHBox(loadBtn, newBtn))
	content := container.NewBorder(topContainer, nil, nil, nil, ui.JobTable)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

func getHeloUser() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	// On Windows, Username can be in the format "DOMAIN\Username". We want just "Username".
	parts := strings.Split(u.Username, "\\")
	return parts[len(parts)-1]
}

func (ui *AdminUI) loadJobs() {
	go func() {
		// Require Windows Hello verification (no-op on non-Windows)
		verified, err := verifyWindowsHello(ui.Window)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Windows Hello failed: %w", err), ui.Window)
			return
		}
		if !verified {
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

	// Add Delete Button if editing existing job
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
		dialog.ShowInformation("Success", "Job saved and scheduler reloaded", ui.Window)
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
