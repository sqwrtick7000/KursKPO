package ui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"salary_app/db"
	"salary_app/models"
)

// ShowAdminWindow открывает главное окно администратора
func ShowAdminWindow(app fyne.App, database *db.Database, currentUser *models.User, onLogout func()) fyne.Window {
	w := app.NewWindow("Администратор: " + currentUser.Login + " — Расчёт ЗП")
	w.Resize(fyne.NewSize(1100, 700))
	w.CenterOnScreen()

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Данные о сотрудниках", theme.HomeIcon(),
			buildAdminEmployeesTab(database, w)),
		container.NewTabItemWithIcon("Пользователи", theme.AccountIcon(),
			buildUsersTab(database, w)),
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
			"👤 Роль: Администратор  |  Логин: "+currentUser.Login,
			fyne.TextAlignLeading, fyne.TextStyle{Bold: true},
		),
	)

	w.SetContent(container.NewBorder(container.NewPadded(top), nil, nil, nil, tabs))
	w.Show()
	return w
}

// ─────────────────── ВКЛАДКА: Данные о сотрудниках ──────────────────────────

func buildAdminEmployeesTab(database *db.Database, w fyne.Window) fyne.CanvasObject {
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

	// Панель кнопок
	addBtn := widget.NewButtonWithIcon("Добавить", theme.ContentAddIcon(), func() {
		employeeForm("Добавить сотрудника", nil, database, refreshTable, w)
	})
	addBtn.Importance = widget.SuccessImportance

	editBtn := widget.NewButtonWithIcon("Редактировать", theme.DocumentCreateIcon(), func() {
		if selectedIdx < 0 || selectedIdx >= len(employees) {
			showError("Выберите строку для редактирования", w)
			return
		}
		e := employees[selectedIdx]
		employeeForm("Редактировать запись", &e, database, refreshTable, w)
	})

	delBtn := widget.NewButtonWithIcon("Удалить", theme.DeleteIcon(), func() {
		if selectedIdx < 0 || selectedIdx >= len(employees) {
			showError("Выберите строку для удаления", w)
			return
		}
		e := employees[selectedIdx]
		confirmDelete(fmt.Sprintf("запись сотрудника «%s»?", e.FullName), func() {
			if err := database.DeleteEmployee(e.ID); err != nil {
				showError("Ошибка удаления: "+err.Error(), w)
				return
			}
			showInfo("Запись успешно удалена", w)
			refreshTable()
		}, w)
	})
	delBtn.Importance = widget.DangerImportance

	refreshBtn := widget.NewButtonWithIcon("Обновить", theme.ViewRefreshIcon(), refreshTable)

	btnBar := container.NewHBox(addBtn, editBtn, delBtn, refreshBtn)

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

	top := container.NewVBox(btnBar, searchBar, sortBar, widget.NewSeparator())
	return container.NewBorder(top, nil, nil, nil, tableHolder)
}

// ─────────────────── ВКЛАДКА: Управление пользователями ─────────────────────

func buildUsersTab(database *db.Database, w fyne.Window) fyne.CanvasObject {
	var users []models.User
	var selectedIdx int = -1

	tableHolder := container.NewMax()

	refreshUsers := func() {
		us, err := database.GetAllUsers()
		if err != nil {
			showError("Ошибка загрузки пользователей: "+err.Error(), w)
			return
		}
		users = us
		selectedIdx = -1

		tbl := widget.NewTable(
			func() (int, int) { return len(users) + 1, 3 },
			func() fyne.CanvasObject { return widget.NewLabel("") },
			func(id widget.TableCellID, o fyne.CanvasObject) {
				lbl := o.(*widget.Label)
				if id.Row == 0 {
					lbl.TextStyle = fyne.TextStyle{Bold: true}
					switch id.Col {
					case 0:
						lbl.SetText("ID")
					case 1:
						lbl.SetText("Логин")
					case 2:
						lbl.SetText("Роль")
					}
					return
				}
				lbl.TextStyle = fyne.TextStyle{}
				u := users[id.Row-1]
				switch id.Col {
				case 0:
					lbl.SetText(strconv.Itoa(u.ID))
				case 1:
					lbl.SetText(u.Login)
				case 2:
					if u.Role == models.RoleAdmin {
						lbl.SetText("Администратор")
					} else {
						lbl.SetText("Пользователь")
					}
				}
			},
		)
		tbl.SetColumnWidth(0, 50)
		tbl.SetColumnWidth(1, 200)
		tbl.SetColumnWidth(2, 150)
		tbl.OnSelected = func(id widget.TableCellID) {
			if id.Row > 0 {
				selectedIdx = id.Row - 1
			}
		}
		tableHolder.Objects = []fyne.CanvasObject{container.NewVScroll(tbl)}
		tableHolder.Refresh()
	}
	refreshUsers()

	showUserForm := func(title string, user *models.User) {
		loginEntry := widget.NewEntry()
		passEntry := widget.NewPasswordEntry()
		roleCheck := widget.NewCheck("Администратор", nil)

		if user != nil {
			loginEntry.SetText(user.Login)
			roleCheck.SetChecked(user.Role == models.RoleAdmin)
			passEntry.SetPlaceHolder("Оставьте пустым — без изменений")
		} else {
			passEntry.SetPlaceHolder("Пароль (обязателен)")
		}

		form := widget.NewForm(
			widget.NewFormItem("Логин *", loginEntry),
			widget.NewFormItem("Пароль", passEntry),
			widget.NewFormItem("Роль", roleCheck),
		)

		dlg := dialog.NewCustomConfirm(title, "Сохранить", "Отмена", form, func(ok bool) {
			if !ok {
				return
			}
			login := strings.TrimSpace(loginEntry.Text)
			if login == "" {
				showError("Логин обязателен", w)
				return
			}
			role := models.RoleUser
			if roleCheck.Checked {
				role = models.RoleAdmin
			}
			if user == nil {
				pass := passEntry.Text
				if pass == "" {
					showError("Пароль обязателен при создании", w)
					return
				}
				if err := database.AddUser(login, pass, role); err != nil {
					showError("Ошибка создания: "+err.Error(), w)
					return
				}
				showInfo("Пользователь «"+login+"» создан", w)
			} else {
				if err := database.UpdateUser(user.ID, login, passEntry.Text, role); err != nil {
					showError("Ошибка обновления: "+err.Error(), w)
					return
				}
				showInfo("Пользователь обновлён", w)
			}
			refreshUsers()
		}, w)
		dlg.Resize(fyne.NewSize(400, 280))
		dlg.Show()
	}

	addBtn := widget.NewButtonWithIcon("Добавить", theme.ContentAddIcon(), func() {
		showUserForm("Добавить пользователя", nil)
	})
	addBtn.Importance = widget.SuccessImportance

	editBtn := widget.NewButtonWithIcon("Редактировать", theme.DocumentCreateIcon(), func() {
		if selectedIdx < 0 || selectedIdx >= len(users) {
			showError("Выберите пользователя", w)
			return
		}
		u := users[selectedIdx]
		showUserForm("Редактировать пользователя", &u)
	})

	delBtn := widget.NewButtonWithIcon("Удалить", theme.DeleteIcon(), func() {
		if selectedIdx < 0 || selectedIdx >= len(users) {
			showError("Выберите пользователя", w)
			return
		}
		u := users[selectedIdx]
		confirmDelete(fmt.Sprintf("пользователя «%s»?", u.Login), func() {
			if err := database.DeleteUser(u.ID); err != nil {
				showError("Ошибка: "+err.Error(), w)
				return
			}
			showInfo("Пользователь удалён", w)
			refreshUsers()
		}, w)
	})
	delBtn.Importance = widget.DangerImportance

	top := container.NewVBox(
		container.NewHBox(addBtn, editBtn, delBtn),
		widget.NewSeparator(),
	)
	return container.NewBorder(top, nil, nil, nil, tableHolder)
}

// ─────────────────── ВКЛАДКА: Расчёт за период ───────────────────────────────

func buildPeriodTab(database *db.Database, w fyne.Window) fyne.CanvasObject {
	tabEntry := widget.NewEntry()
	tabEntry.SetPlaceHolder("Табельный номер сотрудника")

	yearFromEntry := widget.NewEntry()
	yearFromEntry.SetPlaceHolder("2024")
	monthFromEntry := widget.NewEntry()
	monthFromEntry.SetPlaceHolder("1")

	yearToEntry := widget.NewEntry()
	yearToEntry.SetPlaceHolder("2024")
	monthToEntry := widget.NewEntry()
	monthToEntry.SetPlaceHolder("12")

	resultText := widget.NewRichTextFromMarkdown("")
	resultScroll := container.NewVScroll(resultText)

	calcBtn := widget.NewButtonWithIcon("Рассчитать", theme.ConfirmIcon(), func() {
		tab := strings.TrimSpace(tabEntry.Text)
		if tab == "" {
			showError("Введите табельный номер", w)
			return
		}
		yf, e1 := strconv.Atoi(strings.TrimSpace(yearFromEntry.Text))
		mf, e2 := strconv.Atoi(strings.TrimSpace(monthFromEntry.Text))
		yt, e3 := strconv.Atoi(strings.TrimSpace(yearToEntry.Text))
		mt, e4 := strconv.Atoi(strings.TrimSpace(monthToEntry.Text))
		if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
			showError("Введите корректные значения года и месяца", w)
			return
		}
		if mf < 1 || mf > 12 || mt < 1 || mt > 12 {
			showError("Месяц должен быть от 1 до 12", w)
			return
		}

		emps, total, err := database.GetEmployeePeriodTotal(tab, yf, mf, yt, mt)
		if err != nil {
			showError("Ошибка запроса: "+err.Error(), w)
			return
		}
		if len(emps) == 0 {
			resultText.ParseMarkdown("**Данных по указанному сотруднику за этот период не найдено.**")
			return
		}

		var sb strings.Builder
		name := emps[0].FullName
		sb.WriteString(fmt.Sprintf("## %s (таб. №: %s)\n\n", name, tab))
		sb.WriteString(fmt.Sprintf("**Период:** %s %d — %s %d\n\n",
			models.MonthName(mf), yf, models.MonthName(mt), yt))
		sb.WriteString("| Год | Месяц | Часов | Тариф | Брутто | Налог 12%% | Нетто |\n")
		sb.WriteString("|-----|-------|-------|-------|--------|-----------|-------|\n")
		for _, e := range emps {
			sb.WriteString(fmt.Sprintf("| %d | %s | %.1f | %.2f | %.2f | %.2f | **%.2f** |\n",
				e.Year, models.MonthName(e.Month),
				e.HoursWorked, e.HourlyRate,
				e.GrossSalary(), e.Tax(), e.NetSalary()))
		}
		sb.WriteString(fmt.Sprintf("\n---\n**Итого выплачено (нетто): %.2f ₽**\n", total))
		resultText.ParseMarkdown(sb.String())
	})
	calcBtn.Importance = widget.HighImportance

	form := container.NewVBox(
		widget.NewLabelWithStyle("Расчёт выплат за период", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewForm(
			widget.NewFormItem("Табельный номер", tabEntry),
			widget.NewFormItem("С (год)", yearFromEntry),
			widget.NewFormItem("С (месяц)", monthFromEntry),
			widget.NewFormItem("По (год)", yearToEntry),
			widget.NewFormItem("По (месяц)", monthToEntry),
		),
		calcBtn,
		widget.NewSeparator(),
	)

	return container.NewBorder(form, nil, nil, nil, resultScroll)
}
