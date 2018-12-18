package accessor

import (
	"database/sql"
	"time"
)

type UserID int32

type User struct {
	Id         UserID // первичный ключ, через который связаны остальные поля.
	Disposable bool   /* Играет ли пользователь просто так, без sms и регистрации (и попадания
	   в таблицу рекордов). Такие пользователи создаются, когда входят в
	   игру с одним только именем, и удаляются при выходе из партии. */
	Login         string    // видимое другим игрокам имя пользователя
	AvatarAddress string    // адрес относительно корня сайта: '/media/name-src32.ext'
	LastLoginTime time.Time // timestamp
}

type RegularLoginInformation struct {
	// Id           int32
	UserID       UserID
	PasswordHash string // по алгоритму sha3
}

type GameStatistics struct {
	// Id          int32
	UserId      UserID
	GamesPlayed int32 // количество начатых игр
	Wins        int   // количество доведённых до победного конца
}

// текущая принадлежность к игре.
// допущение - только одна игра в один момент времени.
type CurrentLogin struct {
	// Id int32
	UserId UserID
	// токен авторицации, ставящийся как cookie пользователю
	AuthorizationToken sql.NullString
	// csrf token, проверяемый при приёме html форм при загрузке пользовательских данных.
	CSRFToken sql.NullString
}
