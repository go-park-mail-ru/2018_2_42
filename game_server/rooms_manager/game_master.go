package rooms_manager

import (
	"encoding/json"
	"github.com/go-park-mail-ru/2018_2_42/game_server/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strconv"
)

func (r *Room) GameMaster() {
	var message []byte
	var role RoleId
	for {
		select {
		case message = <-r.User0From:
			role = 0
			log.Printf("message came from the User0: " + string(message))
		case message = <-r.User1From:
			role = 1
			log.Printf("message came from the User1: " + string(message))
		}
		event := types.Event{}
		err := json.Unmarshal(message, &event)
		if err != nil {
			response, _ := json.Marshal(types.ErrorMessage(
				"error while parsing first level: " + err.Error()))
			response, _ = json.Marshal(types.Event{
				Method:    "error_message",
				Parameter: response,
			})
			if role == 0 {
				r.User0To <- response
			} else {
				r.User1To <- response
			}
		}
		if event.Method == "upload_map" {
			err := r.UploadMap(role, event.Parameter)
			if err != nil {
				response, _ := json.Marshal(types.ErrorMessage(
					"error while process 'upload_map': " + err.Error()))
				response, _ = json.Marshal(types.Event{
					Method:    "error_message",
					Parameter: response,
				})
				if role == 0 {
					r.User0To <- response
				} else {
					r.User1To <- response
				}
				if r.User0UploadedCharacters && r.User1UploadedCharacters {
					r.DownloadMap(role)
				}
			}
			continue
		}
		if event.Method == "attempt_go_to_cell" {
			gameover, err := r.AttemptGoToCell(role, event.Parameter)
			if err != nil {
				response, _ := json.Marshal(types.ErrorMessage(
					"error while process 'attempt_go_to_cell': " + err.Error()))
				response, _ = json.Marshal(types.Event{
					Method:    "error_message",
					Parameter: response,
				})
				if role == 0 {
					r.User0To <- response
				} else {
					r.User1To <- response
				}
				if r.User0UploadedCharacters && r.User1UploadedCharacters {
					r.DownloadMap(role)
				}
			}
			if gameover {
				// к этому моменту эже все данные должны быть отправлены. только сетевые вопросы и остановка всех 5-и горутин.
				r.StopRoom()
				// TODO: отрегистировать в Rooms.
				break
			}
		}
	}
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
func (r *Room) AttemptGoToCell(role RoleId, message json.RawMessage) (gameover bool, err error) {
	var attemptGoToCell types.AttemptGoToCell
	err = json.Unmarshal(message, &attemptGoToCell)
	if err != nil {
		err = errors.Wrap(err, "in json.Unmarshal message into types.attemptGoToCell: ")
		return
	}
	if role == 0 {
		attemptGoToCell.From = 41 - attemptGoToCell.From
		attemptGoToCell.To = 41 - attemptGoToCell.To
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
	// Сервер смотрит, куда перемещается персонаж и, если целевая клетка пуста,
	// перемещает персонажа на сервере и клиентах.
	if r.Map[attemptGoToCell.To] == nil {
		r.Map[attemptGoToCell.To], r.Map[attemptGoToCell.From] = r.Map[attemptGoToCell.From], nil
		r.MoveCharacter(0, attemptGoToCell.From, attemptGoToCell.To)
		r.MoveCharacter(1, attemptGoToCell.From, attemptGoToCell.To)
		r.YourTurn(0)
		r.YourTurn(1)
		return
	}
	// если в целевой клетке враг
	if r.Map[attemptGoToCell.To].Role != role {
		// проверяем, нет ли там флага
		if r.Map[attemptGoToCell.To].Weapon == "flag" {
			r.Gameover(0, role)
			r.Gameover(1, role)
			gameover = true
			// TODO: каскадный деструктор всего.
			// TODO: запись в базу о конце игры.
			return
		}
		// проверяем победу над обычным оружием.
		if r.Map[attemptGoToCell.From].Weapon.IsExceed(r.Map[attemptGoToCell.To].Weapon) {
			winnerWeapon := r.Map[attemptGoToCell.From].Weapon
			loserWeapon := r.Map[attemptGoToCell.To].Weapon
			// двигаем персонажа
			r.Map[attemptGoToCell.To] = r.Map[attemptGoToCell.From]
			// ставим, что оружие победителя спалилось.
			r.Map[attemptGoToCell.To].ShowedWeapon = true
			// меняем ход // TODO: Возможно, стоит использовать bool в качестве роли.
			if r.UserTurnNumber == 0 {
				r.UserTurnNumber = 1
			} else {
				r.UserTurnNumber = 0
			}
			// отсылаем изменения.
			r.Attack(0, attemptGoToCell.From, winnerWeapon, attemptGoToCell.To, loserWeapon)
			r.Attack(1, attemptGoToCell.From, winnerWeapon, attemptGoToCell.To, loserWeapon)
			// отсылаем смену хода
			r.YourTurn(0)
			r.YourTurn(1)
			return
		}
		// проверяем поражение
		if r.Map[attemptGoToCell.To].Weapon.IsExceed(r.Map[attemptGoToCell.From].Weapon) {
			winnerWeapon := r.Map[attemptGoToCell.To].Weapon
			loserWeapon := r.Map[attemptGoToCell.From].Weapon
			// убираем проигравшего нападавшего персонажа, победитель передвигается на клетку проигравшего.
			r.Map[attemptGoToCell.From] = r.Map[attemptGoToCell.To]
			// ставим, что оружие победителя спалилось.
			r.Map[attemptGoToCell.From].ShowedWeapon = true
			// меняем ход
			if r.UserTurnNumber == 0 {
				r.UserTurnNumber = 1
			} else {
				r.UserTurnNumber = 0
			}
			// отсылаем изменения.
			r.Attack(0, attemptGoToCell.From, winnerWeapon, attemptGoToCell.To, loserWeapon)
			r.Attack(1, attemptGoToCell.From, winnerWeapon, attemptGoToCell.To, loserWeapon)
			// отсылаем смену хода
			r.YourTurn(0)
			r.YourTurn(1)
			return
		}
		if r.Map[attemptGoToCell.To].Weapon == (r.Map[attemptGoToCell.From].Weapon) {
			// меняем ход
			if r.UserTurnNumber == 0 {
				r.UserTurnNumber = 1
			} else {
				r.UserTurnNumber = 0
			}
			r.Map[attemptGoToCell.To].ShowedWeapon = true
			r.Map[attemptGoToCell.From].ShowedWeapon = true
			r.AddWeapon(r.Map[attemptGoToCell.To].Role, attemptGoToCell.From, r.Map[attemptGoToCell.From].Weapon)
			r.AddWeapon(r.Map[attemptGoToCell.From].Role, attemptGoToCell.To, r.Map[attemptGoToCell.To].Weapon)
			r.YourTurn(0)
			r.YourTurn(1)
			return
		}
	} else {
		err = errors.New("attempt to attack yourself")
		return
	}
	return
}

// ответственность: сформировать изменение для клиента, не изменяет карту.
// считает, что карта уже изменена. Вращает для нулевого игрока.
func (r *Room) MoveCharacter(role RoleId, from int, to int) {
	if role == 0 {
		responce, _ := json.Marshal(types.MoveCharacter{
			From: 41 - from,
			To:   41 - to,
		})
		responce, _ = json.Marshal(types.Event{
			Method:    "move_character",
			Parameter: responce,
		})
		r.User0To <- responce
	} else {
		responce, _ := json.Marshal(types.MoveCharacter{
			From: from,
			To:   to,
		})
		responce, _ = json.Marshal(types.Event{
			Method:    "move_character",
			Parameter: responce,
		})
		r.User1To <- responce
	}
	return
}

// ответственность: сборка изменения для клиента, не изменяет карту.
// считает, что карта уже изменена. Вращает для нулевого игрока.
func (r *Room) Attack(role RoleId, winner int, winnerWeapon Weapon, loser int, loserWeapon Weapon) {
	if role == 0 {
		responce, _ := json.Marshal(types.Attack{
			Winner: types.AttackingСharacter{
				Coordinates: 41 - winner,
				Weapon:      string(winnerWeapon),
			},
			Loser: types.AttackingСharacter{
				Coordinates: 41 - loser,
				Weapon:      string(loserWeapon),
			},
		})
		response, _ := json.Marshal(types.Event{
			Method:    "attack",
			Parameter: responce,
		})
		r.User0To <- response
	} else {
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
		r.User1To <- response
	}
	return
}

// ответственность: сборка изменения для клиента, не изменяет карту.
// считает, что карта уже изменена. вращает для нулевого
func (r *Room) AddWeapon(role RoleId, coordinates int, weapon Weapon) {
	if role == 0 {
		response, _ := json.Marshal(types.AddWeapon{
			Coordinates: 41 - coordinates,
			Weapon:      string(weapon),
		})
		response, _ = json.Marshal(types.Event{
			Method:    "add_weapon",
			Parameter: response,
		})
		r.User0To <- response
	} else {
		response, _ := json.Marshal(types.AddWeapon{
			Coordinates: coordinates,
			Weapon:      string(weapon),
		})
		response, _ = json.Marshal(types.Event{
			Method:    "add_weapon",
			Parameter: response,
		})
		r.User1To <- response
	}
	return
}

// ответственность: сборка изменения для клиента, не изменяет карту и не прекращает игру.
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
