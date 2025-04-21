package entity

type App struct {
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Secret      string `db:"secret"`
}
