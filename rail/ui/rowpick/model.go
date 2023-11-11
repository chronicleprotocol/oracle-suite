//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package rowpick

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chronicleprotocol/oracle-suite/rail/ui/queue"
)

type DoneMapper func(table.Model) tea.Cmd
type KeyMapper func(table.Model) func(string)
type Data struct {
	Cols   []table.Column
	Rows   []table.Row
	Mapper DoneMapper
	Keys   KeyMapper
}

type Done struct {
	Idx int
	Row table.Row
}

type Model struct {
	table table.Model
	dm    DoneMapper
	km    KeyMapper
}

func NewModel() Model {
	return Model{
		table: table.New(
			table.WithFocused(true),
			table.WithStyles(styles()),
		),
	}
}
func (m Model) Reset() Model {
	m.table.SetCursor(0)
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, queue.Msg(Done{
				Idx: m.table.Cursor(),
				Row: m.table.SelectedRow(),
			}).Bat()
		case "enter":
			if m.dm != nil {
				return m, m.dm(m.table)
			}
			return m, queue.Msg(Done{
				Idx: m.table.Cursor(),
				Row: m.table.SelectedRow(),
			}).Bat()
			// default:
			// 	if m.km != nil {
			// 		m.km(m.table)(msg.String())
			// 	}
		}

	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width)
		m.table.SetHeight(msg.Height - 4)
		return m, nil

	case Data:
		// m.dm = msg.Mapper
		// m.km = msg.Keys
		m.table.SetRows([]table.Row{}) //  because of the way table.Model works - len(rows) must be < len(cols)
		for _, r := range msg.Rows {
			for x, c := range r {
				msg.Cols[x].Width = max(msg.Cols[x].Width, len(c))
			}
		}
		m.table.SetColumns(msg.Cols)
		m.table.SetRows(msg.Rows)
		return m, nil
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Render(m.table.View())
}

func styles() table.Styles {
	tableStyles := table.DefaultStyles()
	tableStyles.Header = tableStyles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	tableStyles.Selected = tableStyles.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)
	return tableStyles
}
