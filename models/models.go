// Package models содержит структуры данных приложения
package models

// User представляет учётную запись пользователя системы
type User struct {
	ID       int
	Login    string
	Password string // хеш SHA-256
	Role     int    // 1 - администратор, 0 - пользователь
}

// Employee представляет запись о сотруднике и его рабочем времени
type Employee struct {
	ID          int
	FullName    string  // Ф.И.О. сотрудника
	TabNumber   string  // Табельный номер
	Year        int     // Год
	Month       int     // Месяц (1-12)
	HoursWorked float64 // Количество проработанных часов за месяц (0-730)
	HourlyRate  float64 // Почасовой тариф (руб./час)
}

const (
	OvertimeThreshold  = 144.0 // порог сверхурочных часов
	OvertimeMultiplier = 2.0   // коэффициент оплаты сверхурочных
	TaxRate            = 0.12  // ставка подоходного налога (12%)
	RoleAdmin          = 1
	RoleUser           = 0
)

// GrossSalary рассчитывает зарплату до вычета налога
func (e *Employee) GrossSalary() float64 {
	//проверка на то чтобы работники за месяц не отрабатывали больше 730 часов(столько всего в месяце)
	if e.HoursWorked > 730 {
		return -1
	}
	if e.HoursWorked <= OvertimeThreshold {
		return e.HoursWorked * e.HourlyRate
	}
	normal := OvertimeThreshold * e.HourlyRate
	overtime := (e.HoursWorked - OvertimeThreshold) * e.HourlyRate * OvertimeMultiplier
	return normal + overtime
}

// NetSalary рассчитывает зарплату после вычета подоходного налога (12%)
func (e *Employee) NetSalary() float64 {
	return e.GrossSalary() * (1 - TaxRate)
}

// Tax возвращает сумму подоходного налога
func (e *Employee) Tax() float64 {
	return e.GrossSalary() * TaxRate
}

// MonthName возвращает название месяца на русском
func MonthName(m int) string {
	months := []string{
		"", "Январь", "Февраль", "Март", "Апрель",
		"Май", "Июнь", "Июль", "Август",
		"Сентябрь", "Октябрь", "Ноябрь", "Декабрь",
	}
	if m < 1 || m > 12 {
		return "Неизвестно"
	}
	return months[m]
}
