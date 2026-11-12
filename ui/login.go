package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"salary_app/db"
	"salary_app/models"
)

// ShowLogin отображает экран авторизации и вызывает onLogin при успехе
func ShowLogin(app fyne.App, database *db.Database, onLogin func(user *models.User)) fyne.Window {
	w := app.NewWindow("Система расчёта заработной платы — Вход")
	w.Resize(fyne.NewSize(420, 380))
	w.SetFixedSize(true)
	w.CenterOnScreen()

	// Заголовок
	title := canvas.NewText("Расчёт заработной платы", color.NRGBA{R: 30, G: 100, B: 200, A: 255})
	title.TextSize = 20
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	subtitle := canvas.NewText("Вариант 6 — КПО", color.NRGBA{R: 100, G: 100, B: 100, A: 255})
	subtitle.TextSize = 13
	subtitle.Alignment = fyne.TextAlignCenter

	// Поля ввода
	loginEntry := widget.NewEntry()
	loginEntry.SetPlaceHolder("Логин")

	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("Пароль")

	errorLabel := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{})
	errorLabel.Hide()

	loginBtn := widget.NewButtonWithIcon("Войти", theme.LoginIcon(), nil)
	loginBtn.Importance = widget.HighImportance

	doLogin := func() {
		login := loginEntry.Text
		pass := passEntry.Text
		if login == "" || pass == "" {
			errorLabel.SetText("⚠ Введите логин и пароль")
			errorLabel.Show()
			return
		}
		user, err := database.Authenticate(login, pass)
		if err != nil {
			errorLabel.SetText("⚠ Ошибка БД: " + err.Error())
			errorLabel.Show()
			return
		}
		if user == nil {
			errorLabel.SetText("✗ Неверный логин или пароль")
			errorLabel.Show()
			passEntry.SetText("")
			return
		}
		errorLabel.Hide()
		onLogin(user)
	}

	loginBtn.OnTapped = doLogin
	passEntry.OnSubmitted = func(_ string) { doLogin() }

	hintLabel := widget.NewLabelWithStyle(
		"По умолчанию: admin / admin123",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)

	content := container.NewVBox(
		container.NewPadded(title),
		subtitle,
		widget.NewSeparator(),
		widget.NewForm(
			widget.NewFormItem("Логин", loginEntry),
			widget.NewFormItem("Пароль", passEntry),
		),
		errorLabel,
		loginBtn,
		hintLabel,
	)

	w.SetContent(container.NewPadded(content))
	w.Show()
	return w
}
