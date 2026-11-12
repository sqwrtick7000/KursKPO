// Package ui содержит все экраны и компоненты графического интерфейса
package ui

import (
	"fmt"
	"strconv"
	"strings"

	"salary_app/db"
	"salary_app/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// employeeForm отображает диалог добавления/редактирования сотрудника
func employeeForm(title string, emp *models.Employee, database *db.Database, onSave func(), w fyne.Window) {
	fullNameEntry := widget.NewEntry()
	fullNameEntry.SetPlaceHolder("Иванов Иван Иванович")

	tabEntry := widget.NewEntry()
	tabEntry.SetPlaceHolder("001")

	yearEntry := widget.NewEntry()
	yearEntry.SetPlaceHolder("2026")

	monthEntry := widget.NewEntry()
	monthEntry.SetPlaceHolder("1-12")

	hoursEntry := widget.NewEntry()
	hoursEntry.SetPlaceHolder("160.0")

	rateEntry := widget.NewEntry()
	rateEntry.SetPlaceHolder("20.00")

	if emp != nil {
		fullNameEntry.SetText(emp.FullName)
		tabEntry.SetText(emp.TabNumber)
		yearEntry.SetText(strconv.Itoa(emp.Year))
		monthEntry.SetText(strconv.Itoa(emp.Month))
		hoursEntry.SetText(fmt.Sprintf("%.2f", emp.HoursWorked))
		rateEntry.SetText(fmt.Sprintf("%.2f", emp.HourlyRate))
	}

	form := widget.NewForm(
		widget.NewFormItem("Ф.И.О. сотрудника *", fullNameEntry),
		widget.NewFormItem("Табельный номер *", tabEntry),
		widget.NewFormItem("Год *", yearEntry),
		widget.NewFormItem("Месяц (1-12) *", monthEntry),
		widget.NewFormItem("Часов отработано *", hoursEntry),
		widget.NewFormItem("Почасовой тариф ($) *", rateEntry),
	)

	dlg := dialog.NewCustomConfirm(title, "Сохранить", "Отмена", form, func(ok bool) {
		if !ok {
			return
		}
		// Валидация
		if strings.TrimSpace(fullNameEntry.Text) == "" || strings.TrimSpace(tabEntry.Text) == "" {
			showError("Ф.И.О. и табельный номер обязательны", w)
			return
		}
		year, err1 := strconv.Atoi(strings.TrimSpace(yearEntry.Text))
		month, err2 := strconv.Atoi(strings.TrimSpace(monthEntry.Text))
		hours, err3 := strconv.ParseFloat(strings.TrimSpace(hoursEntry.Text), 64)
		rate, err4 := strconv.ParseFloat(strings.TrimSpace(rateEntry.Text), 64)

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			showError("Введите корректные числовые значения", w)
			return
		}
		if month < 1 || month > 12 {
			showError("Месяц должен быть от 1 до 12", w)
			return
		}
		if hours < 0 || hours > 730 {
			showError("Часы не могут быть отрицательными и быть больше 730", w)
			return
		}
		if rate < 0 {
			showError("Тариф не может быть отрицательными", w)
			return
		}

		e := models.Employee{
			FullName:    strings.TrimSpace(fullNameEntry.Text),
			TabNumber:   strings.TrimSpace(tabEntry.Text),
			Year:        year,
			Month:       month,
			HoursWorked: hours,
			HourlyRate:  rate,
		}
		if emp != nil {
			e.ID = emp.ID
			if err := database.UpdateEmployee(e); err != nil {
				showError("Ошибка обновления: "+err.Error(), w)
				return
			}
			showInfo("Запись успешно обновлена", w)
		} else {
			if err := database.AddEmployee(e); err != nil {
				showError("Ошибка добавления: "+err.Error(), w)
				return
			}
			showInfo("Запись успешно добавлена", w)
		}
		onSave()
	}, w)
	dlg.Resize(fyne.NewSize(500, 420))
	dlg.Show()
}

// buildEmployeeTable строит таблицу сотрудников
func buildEmployeeTable(employees []models.Employee, onSelect func(int)) *widget.Table {
	headers := []string{"ID", "Ф.И.О.", "Таб. №", "Год", "Мес.", "Часы", "Тариф", "Брутто (₽)", "Налог (₽)", "Нетто (₽)"}

	table := widget.NewTable(
		func() (int, int) { return len(employees) + 1, len(headers) },
		func() fyne.CanvasObject {
			return widget.NewLabel("placeholder")
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			label.Wrapping = fyne.TextTruncate
			if id.Row == 0 {
				label.TextStyle = fyne.TextStyle{Bold: true}
				label.SetText(headers[id.Col])
				return
			}
			label.TextStyle = fyne.TextStyle{}
			e := employees[id.Row-1]
			switch id.Col {
			case 0:
				label.SetText(strconv.Itoa(e.ID))
			case 1:
				label.SetText(e.FullName)
			case 2:
				label.SetText(e.TabNumber)
			case 3:
				label.SetText(strconv.Itoa(e.Year))
			case 4:
				label.SetText(models.MonthName(e.Month))
			case 5:
				label.SetText(fmt.Sprintf("%.1f", e.HoursWorked))
			case 6:
				label.SetText(fmt.Sprintf("%.2f", e.HourlyRate))
			case 7:
				label.SetText(fmt.Sprintf("%.2f", e.GrossSalary()))
			case 8:
				label.SetText(fmt.Sprintf("%.2f", e.Tax()))
			case 9:
				label.SetText(fmt.Sprintf("%.2f", e.NetSalary()))
			}
		},
	)

	colWidths := []float32{40, 220, 80, 60, 100, 70, 80, 110, 100, 110}
	for i, w := range colWidths {
		table.SetColumnWidth(i, w)
	}
	table.SetRowHeight(0, 36)

	table.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 && onSelect != nil {
			onSelect(id.Row - 1)
		}
	}
	return table
}

// showError показывает диалог ошибки
func showError(msg string, w fyne.Window) {
	dialog.ShowError(fmt.Errorf(msg), w)
}

// showInfo показывает информационный диалог
func showInfo(msg string, w fyne.Window) {
	dialog.ShowInformation("Успех", msg, w)
}

// confirmDelete показывает диалог подтверждения удаления
func confirmDelete(what string, onConfirm func(), w fyne.Window) {
	dialog.ShowConfirm(
		"Подтверждение удаления",
		fmt.Sprintf("Вы действительно хотите удалить %s?", what),
		func(ok bool) {
			if ok {
				onConfirm()
			}
		},
		w,
	)
}

// makeSortBar создаёт панель сортировки
func makeSortBar(fields []string, fieldKeys []string, onSort func(key, dir string)) *fyne.Container {
	fieldSel := widget.NewSelect(fields, nil)
	fieldSel.SetSelectedIndex(0)
	dirSel := widget.NewSelect([]string{"По возрастанию", "По убыванию"}, nil)
	dirSel.SetSelectedIndex(0)

	btn := widget.NewButtonWithIcon("Сортировать", theme.ListIcon(), func() {
		idx := fieldSel.SelectedIndex()
		if idx < 0 {
			idx = 0
		}
		dir := "ASC"
		if dirSel.SelectedIndex() == 1 {
			dir = "DESC"
		}
		onSort(fieldKeys[idx], dir)
	})
	return container.NewHBox(
		widget.NewLabel("Поле:"), fieldSel,
		widget.NewLabel("Порядок:"), dirSel,
		btn,
	)
}
