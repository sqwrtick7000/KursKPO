// Программа расчёта заработной платы сотрудников предприятия.
// Вариант 6 — Курсовая работа по КПО.
// Язык: Go. GUI: Fyne v2. БД: SQLite (go-sqlite3).
package main

import (
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"salary_app/db"
	"salary_app/models"
	"salary_app/ui"
)

func main() {
	// Путь к файлу БД — рядом с исполняемым файлом
	execPath, _ := os.Executable()
	dbPath := filepath.Join(filepath.Dir(execPath), "salary.db")

	// Открываем (или создаём) БД
	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("Не удалось открыть базу данных: %v", err)
	}
	defer database.Close()

	// Создаём приложение Fyne
	a := app.NewWithID("ru.salary.app.v6")
	a.Settings().SetTheme(newSalaryTheme())

	var loginWin fyne.Window
	var mainWin fyne.Window

	// startMainWindow объявляется как var чтобы допускать рекурсивное использование в onLogout
	var startMainWindow func(user *models.User)
	startMainWindow = func(user *models.User) {
		if loginWin != nil {
			loginWin.Close()
		}

		onLogout := func() {
			// После выхода показываем экран логина снова
			loginWin = ui.ShowLogin(a, database, startMainWindow)
		}

		if user.Role == models.RoleAdmin {
			mainWin = ui.ShowAdminWindow(a, database, user, onLogout)
		} else {
			mainWin = ui.ShowUserWindow(a, database, user, onLogout)
		}
		_ = mainWin
	}

	// Первый экран — авторизация
	loginWin = ui.ShowLogin(a, database, startMainWindow)

	a.Run()
}
