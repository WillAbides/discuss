package main

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"os/exec"
)

func loadDiscussions(table *tview.Table, app *tview.Application, discFunc func() ([]discussion, error)) error {
	teamDiscussions, err := discFunc()
	if err != nil {
		return err
	}
	table.
		SetCell(0, 0, tview.NewTableCell("Team").SetSelectable(false)).
		SetCell(0, 1, tview.NewTableCell("Author").SetSelectable(false)).
		SetCell(0, 2, tview.NewTableCell("Created At").SetSelectable(false)).
		SetCell(0, 3, tview.NewTableCell("Title").SetSelectable(false).SetExpansion(1))

	for i, discussion := range teamDiscussions {
		j := i + 1
		lt := discussion.CreatedAt.Local()
		table.
			SetCell(j, 0, tview.NewTableCell(discussion.Team.Name)).
			SetCell(j, 1, tview.NewTableCell(discussion.Author.Login)).
			SetCell(j, 2, tview.NewTableCell(lt.Format(" Mon _2 Jan 15:04 "))).
			SetCell(j, 3, tview.NewTableCell(discussion.Title))
	}
	table.SetFixed(1, 0)
	app.QueueUpdateDraw(func() {
		app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEnter:
				row, _ := table.GetSelection()
				if row > 0 {
					disc := teamDiscussions[row-1]
					_ = exec.Command("open", disc.URL).Run()
				}
			default:
			}
			return event
		})
		app.SetRoot(table, true)
	})
	return nil
}

func loading(modal *tview.Modal, app *tview.Application, loadingCh, killCh chan struct{}) {
	i := 0
	foreverMsg := []string{"Loading ", "discussions ", "takes ", "F", "O", "R", "E", "V", "E", "R"}
	output := ""
	for {
		select {
		case <-killCh:
			return
		case _, ok := <-loadingCh:
			if !ok {
				return
			}
			if len(foreverMsg) > i {
				output = output + foreverMsg[i]
			} else {
				output = output + "."
			}
			i++
			app.QueueUpdateDraw(func() {
				modal.SetText(output)
			})
		}

	}
}

func runUI(loadingChan chan struct{}, discFunc func() ([]discussion, error)) error {
	app := tview.NewApplication()
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)
	table.SetDoneFunc(func(key tcell.Key) {
		app.Stop()
	})
	modal := tview.NewModal()

	killLoading := make(chan struct{})

	go func() {
		loading(modal, app, loadingChan, killLoading)
	}()

	go func() {
		err := loadDiscussions(table, app, discFunc)
		if err != nil {
			close(killLoading)
			app.QueueUpdateDraw(func() {
				modal.SetText("error:\n\n" + err.Error() + "\n\nctrl-c to exit")
			})
		}
	}()

	return app.SetRoot(modal, true).SetFocus(modal).Run()
}
