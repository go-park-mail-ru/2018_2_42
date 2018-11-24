// типы, c помощью которых ведётся работа с клиентом.
// полностю повторяет описание взаимодействия в
// 'github.com/go-park-mail-ru/2018_2_42/doc/Interaction inside WebSocket.txt'.

package types

import "encoding/json"

// первый уровень парсинга
type Event struct {
	// Строка с именем вызываемого метода.
	Method string `json:"method"`
	// Параметры. Так как неизвестен формат, всё парсится в 2 этапа.
	// сначала только эта структура, потом по названию выбирается функция, в
	// надстройку к ней передаётся RawMessage, который парсится в конкретную структуру.
	Parameter json.RawMessage `json:"parameter"`
}

type UploadMap struct {
	Color   string     `json:"color"`
	Weapons [14]string `json:"weapons"`
}

type AttemptGoToCell struct {
	From int `json:"from"`
	To   int `json:"to"`
}

// for DownloadMap only.
type MapCell struct {
	Color  string  `json:"color"`
	Weapon *string `json:"weapon"`
}

type DownloadMap [42]*MapCell

type YourRival []byte

type YourTurn bool

type MoveCharacter struct {
	From int `json:"from"`
	To   int `json:"to"`
}

// for Attack only.
type AttackingСharacter struct {
	Coordinates int    `json:"coordinates"`
	Weapon      string `json:"weapon"`
}

type Attack struct {
	Winner AttackingСharacter `json:"winner"`
	Loser  AttackingСharacter `json:"loser"`
}

type AddWeapon struct {
	Coordinates int    `json:"coordinates"`
	Weapon      string `json:"weapon"`
}

type Gameover struct {
	WinnerColor string `json:"winner_color"`
}

type ErrorMessage string
