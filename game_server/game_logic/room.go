package game_logic

import (
	"github.com/go-park-mail-ru/2018_2_42/game_server/user_connection"
	"github.com/gorilla/websocket"
	"log"
)

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

	// для того, что бы отрегистрировать комнату надо отправить RoomId в этот канал.
	Completed chan RoomId
	OwnNumber RoomId
}

func NewRoom(player0, player1 *user_connection.UserConnection, completedRooms chan RoomId, ownNumber RoomId) (room *Room) {
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

		Completed: completedRooms,
		OwnNumber: ownNumber,
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
	// r.User0.Connection.Close() сделает WebSocketWriter(0) после отправки последнего сообщения
	close(r.User1IsAvailableRead)
	close(r.User1IsAvailableWrite)
	// r.User1.Connection.Close() сделает WebSocketWriter(1) после отправки последнего сообщения
	close(r.User0To)
	close(r.User1To)
	log.Print("room with User0.Token='" + r.User0.Token + "', r.User1.Token='" + r.User1.Token + "' closed")
	return
}

// удаляет комнату из списка комнат.
// Удаляет соединения пользователей из списка обрабатываемых соединений.
func (r *Room) RemoveRoom() {
	r.Completed <- r.OwnNumber
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
	log.Print("WebSocketReader room = " + r.OwnNumber.String() + ", role = " + role.String() + " correctly completed.")
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
	log.Print("WebSocketWriter room = " + r.OwnNumber.String() + ", role = " + role.String() + " correctly completed.")
	return
}
