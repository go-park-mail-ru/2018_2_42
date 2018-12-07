// типы, c помощью которых ведётся работа с клиентом.
// полностю повторяет описание взаимодействия в
// 'github.com/go-park-mail-ru/2018_2_42/doc/Interaction inside WebSocket.txt'.

package types

import (
	"github.com/mailru/easyjson"
	"github.com/pkg/errors"
	"strconv"
)

// первый уровень парсинга
//easyjson:json
type Event struct {
	// Строка с именем вызываемого метода.
	Method string `json:"method,required"`
	// Параметры. Так как неизвестен формат, всё парсится в 2 этапа.
	// сначала только эта структура, потом по названию выбирается функция, в
	// надстройку к ней передаётся RawMessage, который парсится в конкретную структуру.
	Parameter easyjson.RawMessage `json:"parameter,required"`
}

//easyjson:json
type UploadMap struct {
	Weapons [14]string `json:"weapons,required"`
}

//easyjson:json
type AttemptGoToCell struct {
	From int `json:"from,required"`
	To   int `json:"to,required"`
}

func (a *AttemptGoToCell) Check() (err error) {
	if a.From < 0 && 41 < a.From && a.To < 0 && 41 < a.To {
		err = errors.New(strconv.Itoa(a.From) + " or " + strconv.Itoa(a.To) + " out of range.")
	}
	switch a.From - a.To {
	case -7: // ⍗
	case -1: // ⍈
	case +1: // ⍇
	case +7: // ⍐
	default:
		err = errors.New(strconv.Itoa(a.From) + " and " + strconv.Itoa(a.To) + " not in adjacent cells.")
	}
	return
}

//easyjson:json
type ReassignWeapons struct {
	NewWeapon         string `json:"new_weapon,required"`
	CharacterPosition int    `json:"character_position,required"`
}

//easyjson:json
type MapCell struct {
	// type for DownloadMap only.
	User   bool    `json:"user,required"`
	Weapon *string `json:"weapon,required"`
}

//easyjson:json
type DownloadMap [42]*MapCell

type YourRival string

func (yr YourRival) MarshalJSON() ([]byte, error) { // easyjson не захотел работать со string
	return []byte("\"" + yr + "\""), nil
}

type YourTurn bool

//easyjson:json
type MoveCharacter struct {
	From int `json:"from,required"`
	To   int `json:"to,required"`
}

// type for struct Attack only
//easyjson:json
type AttackingСharacter struct {
	Coordinates int    `json:"coordinates,required"`
	Weapon      string `json:"weapon,required"`
}

//easyjson:json
type Attack struct {
	Winner AttackingСharacter `json:"winner,required"`
	Loser  AttackingСharacter `json:"loser,required"`
}

//easyjson:json
type AddWeapon struct {
	Coordinates int    `json:"coordinates,required"`
	Weapon      string `json:"weapon,required"`
}

//easyjson:json
type WeaponChangeRequest struct {
	CharacterPosition int `json:"character_position,required"`
}

//easyjson:json
type GameOver struct {
	Winner bool `json:"winner,required"` // true - вы, false - ваш соперник
	From   int  `json:"from,required"`
	To     int  `json:"to,required"`
}

type ErrorMessage string

func (em ErrorMessage) MarshalJSON() ([]byte, error) { // easyjson не захотел работать со string
	return []byte("\"" + em + "\""), nil
}

//easyjson:json
type ServerResponse struct {
	Status  string `json:"status,required"`
	Message string `json:"message,required"`
}
