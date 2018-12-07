package rooms_manager

import (
	"errors"
	"fmt"
	"github.com/go-park-mail-ru/2018_2_42/game_server/user_connection"
	"github.com/gorilla/websocket"
	"log"
)

// адрес игровой комнаты, уникальный для данного сервера.
type RoomId uint

// роль персонажа. при передаче состояния пользователю, если роли равны, персонаж
// называется синим, не равны - красным.
type RoleId uint8 // ∈ [0, 1]

// описание принадлежности к игре. Номер игровой комнаты и номер в игре,
// певый или второй игрок. Второй хранится на сервере в перевёрнутом состоянии.
type GameToConnect struct {
	Room RoomId
	Role RoleId
}

// Оружие персонажа. Нападение на персонажа со флагом вызывает конец игры. Флаг не может нападать.
type Weapon string // ∈ ["stone", "scissors", "paper", "flag"]

func NewWeapon(key string) (weapon Weapon, err error) {
	switch key {
	case "rock":
		fallthrough
	case "scissors":
		fallthrough
	case "paper":
		fallthrough
	case "flag":
		weapon = Weapon(key)
	default:
		err = errors.New("'" + key + "' ∉ ['rock', 'scissors', 'paper', 'flag']")
	}
	return
}

// true если превосходит передаваемое значение, false
func (w *Weapon) IsExceed(rival Weapon) (exceed bool) {
	switch *w {
	case "rock":
		exceed = rival == "scissors"
	case "scissors":
		exceed = rival == "paper"
	case "paper":
		exceed = rival == "rock"
	}
	return
}

// Персонаж в представлении сервера.
type Сharacter struct {
	Role         RoleId
	Weapon       Weapon
	ShowedWeapon bool
}

func (c *Сharacter) String() (str string) {
	if c == nil {
		str = "            "
	} else {
		if c.Role == 0 {
			str += "0 "
		} else {
			str += "1 "
		}
		switch c.Weapon {
		case "rock":
			str += "rock     "
		case "scissors":
			str += "scissors "
		case "paper":
			str += "paper    "
		case "flag":
			str += "flag     "
		}
		if c.ShowedWeapon {
			str += "+"
		} else {
			str += "-"
		}
	}
	return
}

// Карта в представлении сервера, координаты клеток 0 <= x <= 41, для пустых клеток nil.
//[ 0,  1,  2,  3,  4,  5,  6,
//  7,  8,  9, 10, 11, 12, 13,
// 14, 15, 16, 17, 18, 19, 20,
// 21, 22, 23, 24, 25, 26, 27,
// 28, 29, 30, 31, 32, 33, 34,
// 35, 36, 37, 38, 39, 40, 41]
type Map [42]*Сharacter

func (m Map) String() (str string) { // implement fmt.Stringer interface, called fmt.Print()
	separator := "├────────────┼────────────┼────────────┼────────────┼────────────┼────────────┼────────────┤\n"
	row := func(i int) string {
		return fmt.Sprint("│", m[i], "│", m[i+1], "│", m[i+2], "│", m[i+3], "│", m[i+4], "│", m[i+5], "│", m[i+6], "│\n")
	}
	str = "┌────────────┬────────────┬────────────┬────────────┬────────────┬────────────┬────────────┐\n" +
		row(0) + separator + row(7) + separator + row(14) + separator + row(21) + separator + row(28) + separator + row(35) +
		"└────────────┴────────────┴────────────┴────────────┴────────────┴────────────┴────────────┘\n"
	return
}

type Room struct {
	// соединения с пользователями, могут подменятся во время игры
	User0 *user_connection.UserConnection // array index == RoleId
	User1 *user_connection.UserConnection
	// основные состояния игры.
	Map                     Map
	User0UploadedCharacters bool
	User1UploadedCharacters bool
	UserTurnNumber          RoleId

	// необходимые для перевыбора оружия состояния:
	WeaponReElection struct {
		// находится в состоянии перевыбора, ожидает пока оба клиента
		// пришлют валидные данные на метод "reassign_weapons".
		// можно прислать свой перевыбор только 1 раз
		WaitingForIt bool
		// нулевой пользователь перевыбрал
		User0ReElect bool
		// первый пользователь перевыбрал
		User1ReElect bool
		// атакующий персонаж
		AttackingCharacter int
		// атакуемый персонаж
		AttackedCharacter int
	}

	// переменные для синхронизации мастера игры и читающих/пишуших в Websocket горутин:
	User0From             chan []byte
	User0To               chan []byte
	User0IsAvailableRead  chan struct{}
	User0IsAvailableWrite chan struct{}
	User1From             chan []byte
	User1To               chan []byte
	User1IsAvailableRead  chan struct{}
	User1IsAvailableWrite chan struct{}
}

func NewRoom(player0, player1 *user_connection.UserConnection) (room *Room) {
	room = &Room{
		User0: player0,
		User1: player1,

		Map: [42]*Сharacter{},

		User0From:             make(chan []byte, 5),
		User0To:               make(chan []byte, 5),
		User0IsAvailableRead:  make(chan struct{}, 1),
		User0IsAvailableWrite: make(chan struct{}, 1),
		User1From:             make(chan []byte, 5),
		User1To:               make(chan []byte, 5),
		User1IsAvailableRead:  make(chan struct{}, 1),
		User1IsAvailableWrite: make(chan struct{}, 1),
	}
	go room.WebSocketReader(0)
	go room.WebSocketWriter(0)
	go room.WebSocketReader(1)
	go room.WebSocketWriter(1)
	go room.GameMaster()

	log.Printf("Room created with User0 = '%s', User1 = '%s'", room.User0.Token, room.User1.Token)
	return
}

// Деструктор комнаты.
// отключение горутин, должно вызываться из game master.
//    ╭─User0From─▶─╮      ╭─◀─User1From─╮
// User0           GameMaster            User1
//    ╰─User0To───◀─╯      ╰─▶─User1To───╯
func (r *Room) StopRoom() {
	close(r.User0IsAvailableRead)
	close(r.User0IsAvailableWrite)
	_ = r.User0.Connection.Close()
	close(r.User1IsAvailableRead)
	close(r.User1IsAvailableWrite)
	_ = r.User1.Connection.Close()
	close(r.User0To)
	close(r.User1To)
	log.Print("room with User0.Token='" + r.User0.Token + "', r.User1.Token='" + r.User1.Token + "' closed")
	return
}

// восстанавливает соединение и перезавускает
func (r *Room) Reconnect(user *user_connection.UserConnection, role RoleId) {
	log.Printf("Reconnect sessioni = '%s' as role %d", user.Token, role)

	if role == 0 {
		if r.User0 != nil {
			_ = r.User0.Connection.Close()
		}
		r.User0 = user
		// если в канале уже есть сигнал, пропускаем.
		select {
		case r.User0IsAvailableRead <- struct{}{}:
		default:
		}
		select {
		case r.User0IsAvailableWrite <- struct{}{}:
		default:
		}
	} else {
		if r.User1 != nil {
			_ = r.User1.Connection.Close()
		}
		r.User1 = user
		// если в канале уже есть сигнал, пропускаем.
		select {
		case r.User1IsAvailableRead <- struct{}{}:
		default:
		}
		select {
		case r.User1IsAvailableWrite <- struct{}{}:
		default:
		}
	}
	return
}

func (r *Room) WebSocketReader(role RoleId) {
	// TODO: Add correct timeout, on move and reconnect.
	// c.conn.SetReadLimit(maxMessageSize)
	// c.conn.SetReadDeadline(time.Now().Add(pongWait))
	// c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	if role == 0 {
		for {
			_, message, err := r.User0.Connection.ReadMessage()
			if err != nil {
				log.Print("Error from user role 0 with Token '" + r.User0.Token + "': '" + err.Error() + "'.")
				_, stillOpen := <-r.User0IsAvailableRead
				if !stillOpen {
					close(r.User0From)
					break
				}
			} else {
				log.Print("message from user role 0 with Token '" + r.User0.Token + "': '" + string(message) + "'.")
				r.User0From <- message
			}
		}
	} else {
		for {
			_, message, err := r.User1.Connection.ReadMessage()
			if err != nil {
				log.Print("Error from user role 1 with Token '" + r.User0.Token + "': '" + err.Error() + "'.")
				_, stillOpen := <-r.User1IsAvailableRead
				if !stillOpen {
					close(r.User1From)
					break
				}
			} else {
				log.Print("message from user role 1 with Token '" + r.User0.Token + "': '" + string(message) + "'.")
				r.User1From <- message
			}
		}
	}
	return
}

func (r *Room) WebSocketWriter(role RoleId) {
	if role == 0 {
	consistentMessageSending0:
		for message := range r.User0To {
			for {
				err := r.User0.Connection.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					_, stillOpen := <-r.User0IsAvailableWrite
					if !stillOpen {
						break consistentMessageSending0
					}
				} else {
					break
				}
			}
		}
		_ = r.User0.Connection.Close()
	} else {
	consistentMessageSending1:
		for message := range r.User1To {
			for {
				err := r.User1.Connection.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					_, stillOpen := <-r.User1IsAvailableWrite
					if !stillOpen {
						break consistentMessageSending1
					}
				} else {
					break
				}
			}
		}
		_ = r.User1.Connection.Close()
	}
	return
}

// Паттерн актор: горутина, распоряжающаяся этим классом запущена из main,
// живёт всё время работы в единственном экземпляре. Блокирующе читает из
// канала connection_upgrader.ConnectionUpgrader.QueueToGame, берёт пользователей
// по одному, проверяет ProcessedPlayers на наличие комнаты для этого пользователя.
// возвращает соединение в комнату или замещает старое, или, если игрок пришёл первый раз,
// добавляет его в создаваемую комнату, помечая соединения в ProcessedPlayers.
type RoomsManager struct {
	// Список соединений, существующих в данный момент.
	// используется для повторного подключения к той же игре, что и раньше.
	// изменяется из конструктора/деструктора игровой комнаты.
	// Ключ - login пользователя.
	ProcessedPlayers map[string]GameToConnect
	// Игровые комнаты.
	// Внутри каждая обслуживается одним мастерм игры - горутиной.
	// 4 горутины на комнату, что изолируют соединение от игровой логики и подметы соединений менеджером потерь.
	//    ╭─User0From─▶─╮      ╭─◀─User1From─╮
	// User0           GameMaster            User1
	//    ╰─User0To───◀─╯      ╰─▶─User1To───╯
	// 4 обслуживающие соединения горутины создаются в момент старта комнаты и живут, как и
	// GameMaster всё время существования комнаты.
	// func User0From обычно заблокирован на чтение из сокета, при разрыве соединения блокируется на
	// чтение из User0IsAvailableRead. Если получает оттуда сигнал - пытается читать снова,
	// если этот канал закрыт - завершает работу.
	// func User0To обычно заблокирован на чтение из канала User0To, если он взял данные от
	// GameMaster, попытался отправить и не смог, то блокируется на чтение из User0IsAvailableWrite,
	// Если получает оттуда сигнал - пытается отправить снова, если этот канал закрыт - завершает
	// работу.
	// C timeout работает GameMaster: обновляет счётчик на каждое событие прихода данных.
	// GameMaster содержит игровую логику, в один поток принимает/рассылает запросы, работает с
	// картой, содержит JSPN RPC сервер, вызывающий функции объекта комнаты.
	Rooms map[RoomId]*Room // TODO: переделать на sync.map, для того, что бы корректно удалить комнату.
	// последний номер созданной комнаты, что бы поддерживать уникальность номеров
	RoomNumber RoomId
}

func NewRoomsManager() (roomsManager *RoomsManager) {
	roomsManager = &RoomsManager{
		ProcessedPlayers: make(map[string]GameToConnect),
		Rooms:            make(map[RoomId]*Room),
	}
	return
}

func (rm *RoomsManager) MaintainConnections(connectionQueue <-chan *user_connection.UserConnection) {
	waitingConnection := (*user_connection.UserConnection)(nil)
	for connection := range connectionQueue {
		game, ok := rm.ProcessedPlayers[connection.Token]
		if ok {
			// восстановление соединения
			log.Printf("Reconnect user = '%s' in role %d to room %d", connection.Token, game.Role, game.Room)

			rm.Rooms[game.Room].Reconnect(connection, game.Role)
		} else {
			if waitingConnection == nil {
				log.Printf("Set connection user = '%s' as waiting", connection.Token)

				waitingConnection = connection
			} else {
				// добавление в новую комнату 2-х соединений и регистрация пользователей,
				// как находящихся в процессе игры.
				log.Printf("create room %d user0 = '%s', user1 = '%s'", rm.RoomNumber, waitingConnection.Token, connection.Token)

				rm.Rooms[rm.RoomNumber] = NewRoom(waitingConnection, connection)
				rm.ProcessedPlayers[waitingConnection.Token] = GameToConnect{
					Room: rm.RoomNumber,
					Role: 0,
				}
				rm.ProcessedPlayers[connection.Token] = GameToConnect{
					Room: rm.RoomNumber,
					Role: 1,
				}
				waitingConnection = nil
				rm.RoomNumber++
			}
		}
	}
	return
}
