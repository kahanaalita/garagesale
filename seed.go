package schema

import (
	"github.com/jmoiron/sqlx"
)

// seeds is a string constant containing all of the queries needed to get the
// db seeded to a useful state for development.
//
// Использование константы в Go файле — это простой способ убедиться, что запросы
// являются частью скомпилированного исполняемого файла и избежать проблем с путями
// рабочей директории. Это имеет недостаток в том, что синтаксис может не подсвечиваться
// и для некоторых случаев может быть труднее читать по сравнению с использованием
// SQL файлов. Также можно рассмотреть комбинированный подход, используя инструмент
// вроде packr или go-bindata.
//
// Обратите внимание, что серверы баз данных, кроме PostgreSQL, могут не поддерживать
// выполнение нескольких запросов в рамках одного выполнения, поэтому эту большую
// константу может потребоваться разбить на части.
const seeds = `
INSERT INTO your_table_name (id, name, quantity, price, created_at) VALUES
('72f8b983-3eb4-48db-9ed0-e45cc6bd716b', 'McDonalds Toys', 75, 120, '2019-01-01 00:00:02.000001+00')
ON CONFLICT DO NOTHING;
`

// Seed runs the set of seed-data queries against db. The queries are
// ran in a transaction and rolled back if any fail.
//
// Seed запускает набор запросов для инициализации данных в базе данных. Запросы
// выполняются в рамках транзакции и откатываются в случае любой ошибки.
func Seed(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(seeds); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}
