package accessor

import (
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/go-park-mail-ru/2018_2_42/chat_server/types"
)

// глобальный объект пакета с пулом соединений.
// Вся SQL логика прикрепляется к нему как методы.
type ConnPool struct {
	pgx.ConnPool
}

// Глобальный регистратор для функций подготовки,
// собирает все добавленные функции подготовк и в одну.
// непотокобезопасно, все init() функции собираются в один исполняемый поток.
type Preparer struct {
	functionsToPrepare []func(*pgx.Conn) error
}

func (p *Preparer) add(function func(*pgx.Conn) error) {
	p.functionsToPrepare = append(p.functionsToPrepare, function)
}

// Добавление привелегированной функции, которая точно будет вызвана первой.
// Используется для создания таблиц и идексов, до подготовки всяких 'select'.
// Вызывать 1 раз.
func (p *Preparer) addFirst(function func(*pgx.Conn) error) {
	if len(p.functionsToPrepare) == 0 {
		p.functionsToPrepare = append(p.functionsToPrepare, function)
	} else {
		p.functionsToPrepare = append([]func(*pgx.Conn) error{function}, p.functionsToPrepare...) // ... - операция распаковки массива как **[] в python3.
	}
}

func (p *Preparer) Execute(conn *pgx.Conn) (err error) {
	for _, function := range p.functionsToPrepare {
		if err := function(conn); err != nil {
			return errors.New("error on execute function '" +
				runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name() +
				"' :" + err.Error())
		}
	}
	return nil
}

// Это статический объект, так же как и init функции, что добавляют функции подготовки сюда.
var Prep Preparer

type Error struct {
	Code            int // http коды из "${GOROOT}/src/net/http/status.go"
	UnderlyingError error
}

func (e *Error) Error() string {
	return "Error '" + http.StatusText(e.Code) + "': " + e.UnderlyingError.Error()
}

// создание таблиц
func init() {
	Prep.addFirst(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
create table if not exists "massages"(
	"to" text null,
	"from" text null,
	"id" serial4 primary key,
	
	"text" text not null,
	"time" timestamp not null,
	"reply" integer references "massages"("id") null
);

-- -- для личных чатов
-- create unique index if not exists "massages_private_chat_to_me" on "massages"(
--     "to", "from", "id", concat(select "s" from values "from", "to" order by "s")
-- ) where "to" is not null; 
-- 
-- create unique index if not exists "massages_private_chat_from_me" on "massages"(
--     "from", "to", "id"
-- ) where "from" is not null and "to" is not null;
-- 
-- -- для общего чата
-- create unique index if not exists "massages_general_chat" on "massages"(
-- 	"to", "from", "id"
-- ) where "to" is null; 
`
		_, err = conn.Exec(sql)
		return
	})
}

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
insert into "massages"(
	"to",
	"from",
	"text",
	"time",
	"reply"
) values (
	$1, $2, $3, $4, $5
) returning "id";
`
		_, err = conn.Prepare("massages_insert", sql)
		return
	})
}

func (cp *ConnPool) MassagesInsert(to *string, from *string, text string, time time.Time, reply uint) (id uint, err error) {
	err = cp.QueryRow("massages_insert", &to, &from, &text, time, &reply).Scan(&id)
	if err != nil {
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
	}
	return
}

// запрос для приватных чатов
func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// TODO: необходим идентификатор чата из отсортированных id по нему.
		// language=PostgreSQL
		sql := `
select
	"to", "from", "id", "text", "time",	"reply"
from 
    "massages"
where 
	("to" = $1 and "from" = $2) or  
	("to" = $1 and "from" = $2) and 
    "id" < $3
order by "id" desc
limit 50
;`
		_, err = conn.Prepare("massages_select", sql)
		return
	})
}

func (cp *ConnPool) MassagesSelect(to *string, from *string, id uint) (messages types.Messages, err error) {
	rows, err := cp.Query("massages_insert", &to, &from, &id)
	if err != nil {
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	defer rows.Close()
	for rows.Next() {
		var message types.Message
		err = rows.Scan(&message.To, &message.From, &message.Id, &message.Text, &message.Time, &message.Reply)
		if err != nil {
			err = &Error{
				Code:            http.StatusInternalServerError,
				UnderlyingError: err,
			}
			return
		}
		messages = append(messages, message)
	}
	return
}
