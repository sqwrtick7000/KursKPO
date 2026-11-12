package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"salary_app/db"
	"salary_app/models"
)

// ShowUserWindow открывает главное окно обычного пользователя
func ShowUserWindow(app fyne.App, database *db.Database, currentUser *models.User, onLogout func()) fyne.Window {
	w := app.NewWindow("Пользователь: " + currentUser.Login + " — Расчёт ЗП")
	w.Resize(fyne.NewSize(1050, 650))
	w.CenterOnScreen()

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Просмотр данных", theme.HomeIcon(),
			buildUserViewTab(database, w)),
		container.NewTabItemWithIcon("Расчёт за период", theme.SearchIcon(),
			buildPeriodTab(database, w)),
	)
	tabs.SetTabLocation(container.TabLocationLeading)

	logoutBtn := widget.NewButtonWithIcon("Выйти", theme.LogoutIcon(), func() {
		dialog.ShowConfirm("Выход", "Выйти из системы?", func(ok bool) {
			if ok {
				w.Close()
				onLogout()
			}
		}, w)
	})

	top := container.NewBorder(nil, nil, nil, logoutBtn,
		widget.NewLabelWithStyle(
			"👤 Роль: Пользователь  |  Логин: "+currentUser.Login,
			fyne.TextAlignLeading, fyne.TextStyle{Bold: true},
		),
	)

	w.SetContent(container.NewBorder(container.NewPadded(top), nil, nil, nil, tabs))
	w.Show()
	return w
}

// buildUserViewTab строит вкладку просмотра данных для обычного пользователя
func buildUserViewTab(database *db.Database, w fyne.Window) fyne.CanvasObject {
	var employees []models.Employee
	var selectedIdx int = -1

	tableHolder := container.NewMax()

	refreshTable := func() {
		emps, err := database.GetAllEmployees()
		if err != nil {
			showError("Ошибка загрузки данных: "+err.Error(), w)
			return
		}
		employees = emps
		selectedIdx = -1
		tbl := buildEmployeeTable(employees, func(idx int) { selectedIdx = idx })
		tableHolder.Objects = []fyne.CanvasObject{container.NewVScroll(tbl)}
		tableHolder.Refresh()
	}
	refreshTable()

	// Задача варианта: показать детализацию по выбранному
	taskBtn := widget.NewButtonWithIcon("Задача: ЗП выбранного", theme.ConfirmIcon(), func() {
		if selectedIdx < 0 || selectedIdx >= len(employees) {
			showError("Выберите строку в таблице", w)
			return
		}
		e := employees[selectedIdx]
		msg := fmt.Sprintf(
			"Сотрудник: %s\nТаб. №: %s\nПериод: %s %d\n\n"+
				"Часов отработано: %.1f\n"+
				"Почасовой тариф: %.2f ₽\n\n"+
				"Обычная оплата (до 144 ч): %.2f ₽\n"+
				"Сверхурочных часов: %.1f (×2)\n"+
				"Брутто (до налога): %.2f ₽\n"+
				"Подоходный налог (12%%): %.2f ₽\n"+
				"────────────────────────\n"+
				"К выплате (нетто): %.2f ₽",
			e.FullName, e.TabNumber,
			models.MonthName(e.Month), e.Year,
			e.HoursWorked, e.HourlyRate,
			normalPay(e), overtimeHours(e),
			e.GrossSalary(), e.Tax(), e.NetSalary(),
		)
		dialog.ShowInformation("Расчёт заработной платы", msg, w)
	})
	taskBtn.Importance = widget.HighImportance

	// Поиск
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Поиск по ФИО или табельному номеру...")
	searchBtn := widget.NewButtonWithIcon("Найти", theme.SearchIcon(), func() {
		q := strings.TrimSpace(searchEntry.Text)
		if q == "" {
			refreshTable()
			return
		}
		emps, err := database.SearchEmployees(q)
		if err != nil {
			showError("Ошибка поиска: "+err.Error(), w)
			return
		}
		if len(emps) == 0 {
			showError("По запросу «"+q+"» ничего не найдено", w)
			return
		}
		employees = emps
		selectedIdx = -1
		tbl := buildEmployeeTable(employees, func(idx int) { selectedIdx = idx })
		tableHolder.Objects = []fyne.CanvasObject{container.NewVScroll(tbl)}
		tableHolder.Refresh()
	})
	searchEntry.OnSubmitted = func(_ string) { searchBtn.OnTapped() }
	clearBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		searchEntry.SetText("")
		refreshTable()
	})
	searchBar := container.NewBorder(nil, nil, nil, container.NewHBox(searchBtn, clearBtn), searchEntry)

	// Сортировка
	sortFields := []string{"Ф.И.О.", "Табельный №", "Год", "Месяц", "Часы", "Тариф"}
	sortKeys := []string{"full_name", "tab_number", "year", "month", "hours_worked", "hourly_rate"}
	sortBar := makeSortBar(sortFields, sortKeys, func(key, dir string) {
		emps, err := database.SortEmployees(key, dir)
		if err != nil {
			showError("Ошибка сортировки: "+err.Error(), w)
			return
		}
		employees = emps
		selectedIdx = -1
		tbl := buildEmployeeTable(employees, func(idx int) { selectedIdx = idx })
		tableHolder.Objects = []fyne.CanvasObject{container.NewVScroll(tbl)}
		tableHolder.Refresh()
	})

	top := container.NewVBox(
		container.NewHBox(taskBtn,
			widget.NewButtonWithIcon("Обновить", theme.ViewRefreshIcon(), refreshTable),
		),
		searchBar,
		sortBar,
		widget.NewSeparator(),
	)
	return container.NewBorder(top, nil, nil, nil, tableHolder)
}

// normalPay — нормальная оплата (до порога сверхурочных)
func normalPay(e models.Employee) float64 {
	if e.HoursWorked <= models.OvertimeThreshold {
		return e.HoursWorked * e.HourlyRate
	}
	return models.OvertimeThreshold * e.HourlyRate
}

// overtimeHours — количество сверхурочных часов
func overtimeHours(e models.Employee) float64 {
	if e.HoursWorked <= models.OvertimeThreshold {
		return 0
	}
	return e.HoursWorked - models.OvertimeThreshold
}
