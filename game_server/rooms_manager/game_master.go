package rooms_manager

import (
	"encoding/json"
	"github.com/go-park-mail-ru/2018_2_42/game_server/types"
	"github.com/pkg/errors"
	"strconv"
)

func (r *Room) GameMaster() {
	//TODO слушать UploadMap и attemptGoToCell.
	return
}

// ответственность: загружает данные от пользователя, начинает игру
func (r *Room) UploadMap(role RoleId, message json.RawMessage) (err error) {
	var uploadedMap types.UploadMap
	err = json.Unmarshal(message, &uploadedMap)
	if err != nil {
		err = errors.Wrap(err, "in json.Unmarshal message into types.UploadMap: ")
		return
	}
	if role == 0 {
		if !r.User0UploadedCharacters {
			// uploadedMap.Weapons для клеток 13 12 11 10 9 8 7 6 5 4 3 2 1 0
			var numberOfFlags int
			for i := 0; i <= 13; i++ {
				j := 13 - i
				var weapon *Weapon
				weapon, err = NewWeapon(uploadedMap.Weapons[i])
				if err != nil {
					err = errors.Wrap(err, "in NewWeapon: ")
					return
				}
				if *weapon == "flag" {
					numberOfFlags++
				}
				r.Map[j] = &Сharacter{
					Role:   0,
					Weapon: *weapon,
				}
			}
			if numberOfFlags != 0 {
				err = errors.New("map must contain exactly one flag, but " +
					strconv.Itoa(numberOfFlags) + "found")
				return
			}
			r.User0UploadedCharacters = true
		} else {
			err = errors.New("characters already loaded")
			return
		}
	} else {
		if !r.User1UploadedCharacters {
			// 28 29 30 31 32 33 34 35 36 37 38 39 40 41
			var numberOfFlags int
			for i := 0; i <= 13; i++ {
				j := 28 + i
				var weapon *Weapon
				weapon, err = NewWeapon(uploadedMap.Weapons[i])
				if err != nil {
					err = errors.Wrap(err, "in NewWeapon: ")
					return
				}
				if *weapon == "flag" {
					numberOfFlags++
				}
				r.Map[j] = &Сharacter{
					Role:   0,
					Weapon: *weapon,
				}
			}
			if numberOfFlags != 0 {
				err = errors.New("map must contain exactly one flag, but " +
					strconv.Itoa(int(numberOfFlags)) + "found")
				return
			}
			r.User1UploadedCharacters = true
		} else {
			err = errors.New("characters already loaded")
			return
		}
	}
	if r.User0UploadedCharacters && r.User1UploadedCharacters {
		// Отсылает карту
		r.DownloadMap(0)
		r.DownloadMap(1)
		// Отсылает логин соперника
		r.YourRival(0)
		r.YourRival(1)
		// Отправляет чей ход
		r.YourTurn(0)
		r.YourTurn(1)
	}
	return
}

// ответственность: отправляет карту на клиент, не изменяет карту.
func (r *Room) DownloadMap(role RoleId) {
	// нужно ли переворачивать текст
	if role == 0 {
		downloadMap := types.DownloadMap{}
		for i := 0; i <= 41; i++ {
			j := 41 - i
			if r.Map[j] == nil {
				continue
			}
			var cell = &types.MapCell{}
			// Собственные персонажи всегда синие.
			if r.Map[j].Role == role {
				cell.Color = "blue"
			} else {
				cell.Color = "red"
			}
			// оружие видно только если это собственный игрок или противник показал оружие.
			if r.Map[j].Role == role || r.Map[j].ShowedWeapon {
				weapon := string(r.Map[j].Weapon)
				cell.Weapon = &weapon
			}
			downloadMap[j] = cell
		}
		parameter, _ := json.Marshal(downloadMap)
		response, _ := json.Marshal(types.Event{
			Method:    "download_map",
			Parameter: parameter,
		})
		r.User0To <- response
	} else {
		downloadMap := types.DownloadMap{}
		for i := 0; i <= 41; i++ {
			if r.Map[i] == nil {
				continue
			}
			var cell = &types.MapCell{}
			// Собственные персонажи всегда синие.
			if r.Map[i].Role == role {
				cell.Color = "blue"
			} else {
				cell.Color = "red"
			}
			// оружие видно только если это собственный игрок или противник показал оружие.
			if r.Map[i].Role == role || r.Map[i].ShowedWeapon {
				weapon := string(r.Map[i].Weapon)
				cell.Weapon = &weapon
			}
			downloadMap[i] = cell
		}
		parameter, _ := json.Marshal(downloadMap)
		response, _ := json.Marshal(types.Event{
			Method:    "download_map",
			Parameter: parameter,
		})
		r.User1To <- response
	}
	return
}

// ответственность: отправляет описание соперника, не изменяет карту.
func (r *Room) YourRival(role RoleId) {
	if role == 0 {
		response := types.YourRival(r.User1.Login)
		response, _ = json.Marshal(types.Event{
			Method:    "your_rival",
			Parameter: []byte(response),
		})
		r.User1To <- response
	} else {
		response := types.YourRival(r.User0.Login)
		response, _ = json.Marshal(types.Event{
			Method:    "your_rival",
			Parameter: []byte(response),
		})
		r.User0To <- response
	}
	return
}

// ответственность: отправляет стат чей ход, не изменяет карту.
func (r *Room) YourTurn(role RoleId) {
	var response []byte
	if types.YourTurn(r.UserTurnNumber == role) {
		response = []byte("true")
	} else {
		response = []byte("false")
	}
	response, _ = json.Marshal(types.Event{
		Method:    "your_turn",
		Parameter: response,
	})
	if role == 0 {
		r.User0To <- response
	} else {
		r.User1To <- response
	}
	return
}

// ответственность: принимает данные от пользователя, обрабатывает с учётом состояния,
// изменяет согласно игровой механике карту (фактически содержит всю игру в себе 😮)
// вызывает функции, отправляющие запросы.
func (r *Room) attemptGoToCell(role RoleId, message json.RawMessage) (err error) {
	var attemptGoToCell types.AttemptGoToCell
	err = json.Unmarshal(message, &attemptGoToCell)
	if err != nil {
		err = errors.Wrap(err, "in json.Unmarshal message into types.attemptGoToCell: ")
		return
	}
	if r.UserTurnNumber == role {
		err = errors.New("it's not your turn now")
		return
	}
	if r.Map[attemptGoToCell.From] == nil {
		err = errors.New("there is no character at " + strconv.Itoa(attemptGoToCell.From))
		return
	}
	if r.Map[attemptGoToCell.From].Role != role {
		err = errors.New("this is not your character at " + strconv.Itoa(attemptGoToCell.From))
		return
	}
	// Тут точно существующий персонаж, принадлежащий игроку.
	// TODO: игровая механика

	return
}

// ответственность: изменение, не изменяет карту.
// считает, что карта уже изменена.
func (r *Room) MoveCharacter(role RoleId, from int, to int) {
	responce, _ := json.Marshal(types.MoveCharacter{
		From: from,
		To:   to,
	})
	responce, _ = json.Marshal(types.Event{
		Method:    "move_character",
		Parameter: responce,
	})
	if role == 1 {
		r.User0To <- responce
	} else {
		r.User1To <- responce
	}
	return
}

// ответственность: изменение, не изменяет карту.
// считает, что карта уже изменена.
func (r *Room) Attack(role RoleId, winner int, winnerWeapon Weapon, loser int, loserWeapon Weapon) {
	responce, _ := json.Marshal(types.Attack{
		Winner: types.AttackingСharacter{
			Coordinates: winner,
			Weapon:      string(winnerWeapon),
		},
		Loser: types.AttackingСharacter{
			Coordinates: loser,
			Weapon:      string(loserWeapon),
		},
	})
	response, _ := json.Marshal(types.Event{
		Method:    "attack",
		Parameter: responce,
	})
	if role == 0 {
		r.User0To <- response
	} else {
		r.User1To <- response
	}
	return
}

// ответственность: изменение, не изменяет карту.
// считает, что карта уже изменена.
func (r *Room) AddWeapon(role RoleId, coordinates int, weapon Weapon) {
	response, _ := json.Marshal(types.AddWeapon{
		Coordinates: coordinates,
		Weapon:      string(weapon),
	})
	response, _ = json.Marshal(types.Event{
		Method:    "add_weapon",
		Parameter: response,
	})
	if role == 0 {
		r.User0To <- response
	} else {
		r.User1To <- response
	}
	return
}

// ответственность: изменение, не изменяет карту и не прекращает игру.
// считает, что карта уже изменена.
func (r *Room) Gameover(role RoleId, winnerRole RoleId) {
	var gameover types.Gameover
	if role == winnerRole {
		gameover.WinnerColor = "blue"
	} else {
		gameover.WinnerColor = "red"
	}
	response, _ := json.Marshal(gameover)
	response, _ = json.Marshal(types.Event{
		Method:    "gameover",
		Parameter: response,
	})
	if role == 0 {
		r.User0To <- response
	} else {
		r.User1To <- response
	}
	return
}

// функции, которые можно вызывать с клиента.
// var availableFunctions = map[string]func(r *Room, role RoleId, message json.RawMessage) (err error){
// 	"upload_map": UploadMap,
//	"attempt_go_to_cell":
//}

// проблемы, почему не используются библиотеки:
// Stateful сервер: необходимо помнить роль, в которой работает пользователь,
// комнату, в которой присутствует пользователь.
// решено делать всё на событиях - клиет пересылает действия пользвателя,
// сервер декларативно присылает изменения, в такой форме, что бы они прямо вызывали анимации.

// сервер получает из одного из двух каналов запись.
// добавляет номер игрока.
// парсит первый уровень.
// находит функцию вызываемую и к ней привязаный тип.
// разворачивает в этот тип пришедшие данные.
