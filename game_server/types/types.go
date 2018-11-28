// типы, c помощью которых ведётся работа с клиентом.
// полностю повторяет описание взаимодействия в
// 'github.com/go-park-mail-ru/2018_2_42/doc/Interaction inside WebSocket.txt'.

package types

import (
	"github.com/mailru/easyjson"
)

// первый уровень парсинга
//easyjson:json
type Event struct {
	// Строка с именем вызываемого метода.
	Method string `json:"method"`
	// Параметры. Так как неизвестен формат, всё парсится в 2 этапа.
	// сначала только эта структура, потом по названию выбирается функция, в
	// надстройку к ней передаётся RawMessage, который парсится в конкретную структуру.
	Parameter easyjson.RawMessage `json:"parameter"`
}

//easyjson:json
type UploadMap struct {
	Color   string     `json:"color"`
	Weapons [14]string `json:"weapons"`
}

//easyjson:json
type AttemptGoToCell struct {
	From int `json:"from"`
	To   int `json:"to"`
}

//easyjson:json
type MapCell struct { // type for DownloadMap only.
	Color  string  `json:"color"`
	Weapon *string `json:"weapon"`
}

//easyjson:json
type DownloadMap [42]*MapCell

type YourRival []byte

type YourTurn bool

//easyjson:json
type MoveCharacter struct {
	From int `json:"from"`
	To   int `json:"to"`
}

//easyjson:json
type AttackingСharacter struct { // type for Attack only
	Coordinates int    `json:"coordinates"`
	Weapon      string `json:"weapon"`
}

//easyjson:json
type Attack struct {
	Winner AttackingСharacter `json:"winner"`
	Loser  AttackingСharacter `json:"loser"`
}

//easyjson:json
type AddWeapon struct {
	Coordinates int    `json:"coordinates"`
	Weapon      string `json:"weapon"`
}

//easyjson:json
type Gameover struct {
	WinnerColor string `json:"winner_color"`
}

type ErrorMessage string

func (em ErrorMessage) MarshalJSON() ([]byte, error) { // easyjson не захотел работать со string
	return []byte("\"" + em + "\""), nil
}

//easyjson:json
type ServerResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
