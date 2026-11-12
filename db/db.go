// Package db реализует слой доступа к данным через SQLite
package db

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"salary_app/models"

	_ "github.com/mattn/go-sqlite3"
)

// Database обёртка над sql.DB с методами приложения
type Database struct {
	conn *sql.DB
}

// Open открывает (или создаёт) базу данных по указанному пути
func Open(path string) (*Database, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть БД: %w", err)
	}
	d := &Database{conn: conn}
	if err := d.migrate(); err != nil {
		return nil, fmt.Errorf("ошибка миграции: %w", err)
	}
	return d, nil
}

// Close закрывает соединение с БД
func (d *Database) Close() {
	d.conn.Close()
}

// migrate создаёт таблицы и начальные данные если их нет
func (d *Database) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id       INTEGER PRIMARY KEY AUTOINCREMENT,
		login    TEXT    NOT NULL UNIQUE,
		password TEXT    NOT NULL,
		role     INTEGER NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS employees (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		full_name    TEXT    NOT NULL,
		tab_number   TEXT    NOT NULL,
		year         INTEGER NOT NULL,
		month        INTEGER NOT NULL,
		hours_worked REAL    NOT NULL,
		hourly_rate  REAL    NOT NULL
	);
	`
	if _, err := d.conn.Exec(schema); err != nil {
		return err
	}
	// Создать учётную запись администратора по умолчанию если нет ни одного
	var count int
	d.conn.QueryRow("SELECT COUNT(*) FROM users WHERE role=1").Scan(&count)
	if count == 0 {
		hash := HashPassword("admin123")
		d.conn.Exec("INSERT INTO users (login, password, role) VALUES (?, ?, 1)", "admin", hash)
	}
	return nil
}

// HashPassword возвращает SHA-256 хеш пароля в hex-формате
func HashPassword(password string) string {
	h := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", h)
}

// ─────────────────────────── ПОЛЬЗОВАТЕЛИ ───────────────────────────────────

// Authenticate проверяет логин/пароль и возвращает пользователя или nil
func (d *Database) Authenticate(login, password string) (*models.User, error) {
	hash := HashPassword(password)
	row := d.conn.QueryRow(
		"SELECT id, login, password, role FROM users WHERE login=? AND password=?",
		login, hash,
	)
	u := &models.User{}
	if err := row.Scan(&u.ID, &u.Login, &u.Password, &u.Role); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

// GetAllUsers возвращает список всех пользователей
func (d *Database) GetAllUsers() ([]models.User, error) {
	rows, err := d.conn.Query("SELECT id, login, password, role FROM users ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []models.User
	for rows.Next() {
		u := models.User{}
		rows.Scan(&u.ID, &u.Login, &u.Password, &u.Role)
		users = append(users, u)
	}
	return users, nil
}

// AddUser добавляет нового пользователя (пароль хешируется внутри)
func (d *Database) AddUser(login, password string, role int) error {
	hash := HashPassword(password)
	_, err := d.conn.Exec(
		"INSERT INTO users (login, password, role) VALUES (?, ?, ?)",
		login, hash, role,
	)
	return err
}

// UpdateUser обновляет логин, роль и (при непустом password) пароль пользователя
func (d *Database) UpdateUser(id int, login, password string, role int) error {
	if password != "" {
		hash := HashPassword(password)
		_, err := d.conn.Exec(
			"UPDATE users SET login=?, password=?, role=? WHERE id=?",
			login, hash, role, id,
		)
		return err
	}
	_, err := d.conn.Exec(
		"UPDATE users SET login=?, role=? WHERE id=?",
		login, role, id,
	)
	return err
}

// DeleteUser удаляет пользователя по ID
func (d *Database) DeleteUser(id int) error {
	_, err := d.conn.Exec("DELETE FROM users WHERE id=?", id)
	return err
}

// ─────────────────────────── СОТРУДНИКИ ─────────────────────────────────────

// GetAllEmployees возвращает все записи о сотрудниках
func (d *Database) GetAllEmployees() ([]models.Employee, error) {
	rows, err := d.conn.Query(
		"SELECT id, full_name, tab_number, year, month, hours_worked, hourly_rate FROM employees ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEmployees(rows)
}

// AddEmployee добавляет запись о сотруднике
func (d *Database) AddEmployee(e models.Employee) error {
	_, err := d.conn.Exec(
		"INSERT INTO employees (full_name, tab_number, year, month, hours_worked, hourly_rate) VALUES (?,?,?,?,?,?)",
		e.FullName, e.TabNumber, e.Year, e.Month, e.HoursWorked, e.HourlyRate,
	)
	return err
}

// UpdateEmployee обновляет запись о сотруднике
func (d *Database) UpdateEmployee(e models.Employee) error {
	_, err := d.conn.Exec(
		"UPDATE employees SET full_name=?, tab_number=?, year=?, month=?, hours_worked=?, hourly_rate=? WHERE id=?",
		e.FullName, e.TabNumber, e.Year, e.Month, e.HoursWorked, e.HourlyRate, e.ID,
	)
	return err
}

// DeleteEmployee удаляет запись по ID
func (d *Database) DeleteEmployee(id int) error {
	_, err := d.conn.Exec("DELETE FROM employees WHERE id=?", id)
	return err
}

// SearchEmployees ищет записи по ФИО или табельному номеру (подстрока)
func (d *Database) SearchEmployees(query string) ([]models.Employee, error) {
	like := "%" + query + "%"
	rows, err := d.conn.Query(
		"SELECT id, full_name, tab_number, year, month, hours_worked, hourly_rate FROM employees WHERE full_name LIKE ? OR tab_number LIKE ? ORDER BY full_name",
		like, like,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEmployees(rows)
}

// SortEmployees возвращает сотрудников, отсортированных по указанному полю
func (d *Database) SortEmployees(field, direction string) ([]models.Employee, error) {
	allowed := map[string]bool{
		"full_name": true, "tab_number": true, "year": true,
		"month": true, "hours_worked": true, "hourly_rate": true,
	}
	if !allowed[field] {
		field = "full_name"
	}
	if direction != "ASC" && direction != "DESC" {
		direction = "ASC"
	}
	rows, err := d.conn.Query(
		fmt.Sprintf("SELECT id, full_name, tab_number, year, month, hours_worked, hourly_rate FROM employees ORDER BY %s %s", field, direction),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEmployees(rows)
}

// GetEmployeePeriodTotal возвращает суммарные выплаты (net) сотруднику за период
func (d *Database) GetEmployeePeriodTotal(tabNumber string, yearFrom, monthFrom, yearTo, monthTo int) ([]models.Employee, float64, error) {
	rows, err := d.conn.Query(
		`SELECT id, full_name, tab_number, year, month, hours_worked, hourly_rate
		 FROM employees
		 WHERE tab_number=?
		   AND (year*12+month) >= ?
		   AND (year*12+month) <= ?
		 ORDER BY year, month`,
		tabNumber,
		yearFrom*12+monthFrom,
		yearTo*12+monthTo,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	emps, err := scanEmployees(rows)
	if err != nil {
		return nil, 0, err
	}
	var total float64
	for _, e := range emps {
		total += e.NetSalary()
	}
	return emps, total, nil
}

// scanEmployees считывает строки в срез Employee
func scanEmployees(rows *sql.Rows) ([]models.Employee, error) {
	var result []models.Employee
	for rows.Next() {
		e := models.Employee{}
		if err := rows.Scan(&e.ID, &e.FullName, &e.TabNumber, &e.Year, &e.Month, &e.HoursWorked, &e.HourlyRate); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

// ClearAllEmployees удаляет все записи о сотрудниках (сброс данных)
func (d *Database) ClearAllEmployees() error {
	_, err := d.conn.Exec("DELETE FROM employees; DELETE FROM sqlite_sequence WHERE name='employees'")
	return err
}

// CountEmployees возвращает количество записей о сотрудниках
func (d *Database) CountEmployees() int {
	var n int
	d.conn.QueryRow("SELECT COUNT(*) FROM employees").Scan(&n)
	return n
}

// SeedDemoData добавляет демонстрационные данные
func (d *Database) SeedDemoData() error {
	demo := []models.Employee{
		{FullName: "Иванов Иван Иванович", TabNumber: "001", Year: 2024, Month: 1, HoursWorked: 160, HourlyRate: 500},
		{FullName: "Иванов Иван Иванович", TabNumber: "001", Year: 2024, Month: 2, HoursWorked: 152, HourlyRate: 500},
		{FullName: "Иванов Иван Иванович", TabNumber: "001", Year: 2024, Month: 3, HoursWorked: 176, HourlyRate: 500},
		{FullName: "Петрова Мария Сергеевна", TabNumber: "002", Year: 2024, Month: 1, HoursWorked: 144, HourlyRate: 650},
		{FullName: "Петрова Мария Сергеевна", TabNumber: "002", Year: 2024, Month: 2, HoursWorked: 168, HourlyRate: 650},
		{FullName: "Сидоров Алексей Петрович", TabNumber: "003", Year: 2024, Month: 1, HoursWorked: 200, HourlyRate: 750},
		{FullName: "Сидоров Алексей Петрович", TabNumber: "003", Year: 2024, Month: 2, HoursWorked: 144, HourlyRate: 750},
		{FullName: "Козлова Анна Николаевна", TabNumber: "004", Year: 2024, Month: 1, HoursWorked: 120, HourlyRate: 480},
		{FullName: "Козлова Анна Николаевна", TabNumber: "004", Year: 2024, Month: 2, HoursWorked: 136, HourlyRate: 480},
		{FullName: "Новиков Дмитрий Олегович", TabNumber: "005", Year: 2024, Month: 1, HoursWorked: 160, HourlyRate: 900},
	}
	for _, e := range demo {
		if err := d.AddEmployee(e); err != nil {
			return err
		}
	}
	return nil
}
