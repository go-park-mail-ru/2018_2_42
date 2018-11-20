package rooms_manager

import (
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
type GameToСonnect struct {
	Room RoomId
	Role RoleId
}

// Оружие персонажа. Нападение на персонажа со флагом вызывает конец игры. Флаг не может нападать.
type Weapon string // ∈ ["stone", "scissors", "paper", "flag"]

// Персонаж в представлении сервера.
type Сharacter struct {
	Role         RoleId
	Weapon       Weapon
	ShowedWeapon bool
}

// Карта в представлении сервера, координаты клеток 0 <= x <= 41, для пустых клеток nil.
//[ 0,  1,  2,  3,  4,  5,  6,
//  7,  8,  9, 10, 11, 12, 13,
// 14, 15, 16, 17, 18, 19, 20,
// 21, 22, 23, 24, 25, 26, 27,
// 28, 29, 30, 31, 32, 33, 34,
// 35, 36, 37, 38, 39, 40, 41]
type Map [42]*Сharacter

type Room struct {
	User0 *user_connection.UserConnection // array index == RoleId
	User1 *user_connection.UserConnection
	Map   Map

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
		Map: [42]*Сharacter{},

		User0: player0,
		User1: player1,

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

// восстанавливает соединение и перезавускает
func (r *Room) Reconnect(user *user_connection.UserConnection, role RoleId) {
	log.Printf("Reconnect sessioni = '%s' as role %d", user.Token, role)

	if role == 0 {
		if r.User0 != nil {
			r.User0.Connection.Close()
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
			r.User1.Connection.Close()
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
				_, stillOpen := <-r.User0IsAvailableRead
				if !stillOpen {
					close(r.User0From)
					break
				}
			} else {
				r.User0From <- message
			}
		}
	} else {
		for {
			_, message, err := r.User1.Connection.ReadMessage()
			if err != nil {
				_, stillOpen := <-r.User1IsAvailableRead
				if !stillOpen {
					close(r.User1From)
					break
				}
			} else {
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
		r.User0.Connection.Close()
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
		r.User1.Connection.Close()
	}
	return
}

func (r *Room) GameMaster() {
	// TODO: вся логика игры тут.
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
	ProcessedPlayers map[string]GameToСonnect
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
	Rooms map[RoomId]*Room
	// последний номер созданной комнаты, что бы поддерживать уникальность номеров
	RoomNumber RoomId
}

func NewRoomsManager() (roomsManager *RoomsManager) {
	roomsManager = &RoomsManager{
		ProcessedPlayers: make(map[string]GameToСonnect),
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
				rm.ProcessedPlayers[waitingConnection.Token] = GameToСonnect{
					Room: rm.RoomNumber,
					Role: 0,
				}
				rm.ProcessedPlayers[connection.Token] = GameToСonnect{
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
