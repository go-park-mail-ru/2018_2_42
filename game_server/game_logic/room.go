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

	// Каналы, c помощью которых go room.GameMaster() общается с
	// ['go room.WebSocketReader(0)', 'go room.WebSocketWriter(0)', 'go room.WebSocketReader(1)', 'go room.WebSocketWriter(1)']
	Messaging struct{
		User0From             chan []byte
		User0To               chan []byte
		User1From             chan []byte
		User1To               chan []byte
	}

	// Каналы для синхронизации мастера игры и читающих/пишуших в Websocket горутин при разрыве соединения.
	Recovery struct {
		// приход сообщения означает:
		// go room.WebSocketReader(0) может снова заблокироваться на чтение из сокета
		User0IsAvailableRead  chan struct{}
		// go room.WebSocketWriter(0) может снова попытаться отправить сообщение пользователю User0
		User0IsAvailableWrite chan struct{}
		// go room.WebSocketReader(1) может снова заблокироваться на чтение из сокета
		User1IsAvailableRead  chan struct{}
		// go room.WebSocketWriter(1) может снова попытаться отправить сообщение пользователю User1
		User1IsAvailableWrite chan struct{}
	}

	// Что бы отрегистрировать комнату, надо отправить RoomId в канал:
	Completed chan RoomId
	OwnNumber RoomId
}

func NewRoom(player0, player1 *user_connection.UserConnection, completedRooms chan RoomId, ownNumber RoomId) (room *Room) {
	room = &Room{
		User0: player0,
		User1: player1,
		Map: [42]*Сharacter{},
		Completed: completedRooms,
		OwnNumber: ownNumber,
	}
	room.Messaging.User0From = make(chan []byte, 5)
	room.Messaging.User0To = make(chan []byte, 5)
	room.Messaging.User1From = make(chan []byte, 5)
	room.Messaging.User1To = make(chan []byte, 5)
	room.Recovery.User0IsAvailableRead = make(chan struct{}, 1)
	room.Recovery.User0IsAvailableWrite = make(chan struct{}, 1)
	room.Recovery.User1IsAvailableRead = make(chan struct{}, 1)
	room.Recovery.User1IsAvailableWrite = make(chan struct{}, 1)

	// Внутри каждая комната обслуживается одним мастерм игры - горутиной.
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
	// завершит go room.WebSocketReader(0)
	close(r.Recovery.User0IsAvailableRead)

	// завершит go room.WebSocketWriter(0)
	// go WebSocketWriter(0) закроет tcp соединение r.User0.Connection.Close() после отправки последнего сообщения
	close(r.Messaging.User0To)
	close(r.Recovery.User0IsAvailableWrite)

	// завершит go room.WebSocketReader(1)
	close(r.Recovery.User1IsAvailableRead)

	// завершит go room.WebSocketWriter(1)
	// go WebSocketWriter(1) закроет tcp соединение r.User1.Connection.Close() после отправки последнего сообщения
	close(r.Messaging.User1To)
	close(r.Recovery.User1IsAvailableWrite)

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
		case r.Recovery.User0IsAvailableRead <- struct{}{}:
		default:
		}
		select {
		case r.Recovery.User0IsAvailableWrite <- struct{}{}:
		default:
		}
	} else {
		if r.User1 != nil {
			_ = r.User1.Connection.Close()
		}
		r.User1 = user
		// если в канале уже есть сигнал, пропускаем.
		select {
		case r.Recovery.User1IsAvailableRead <- struct{}{}:
		default:
		}
		select {
		case r.Recovery.User1IsAvailableWrite <- struct{}{}:
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
				_, stillOpen := <-r.Recovery.User0IsAvailableRead
				if !stillOpen {
					close(r.Messaging.User0From)
					break
				}
			} else {
				log.Print("message from user role 0 with Token '" + r.User0.Token + "': '" + string(message) + "'.")
				r.Messaging.User0From <- message
			}
		}
	} else {
		for {
			_, message, err := r.User1.Connection.ReadMessage()
			if err != nil {
				log.Print("Error from user role 1 with Token '" + r.User0.Token + "': '" + err.Error() + "'.")
				_, stillOpen := <-r.Recovery.User1IsAvailableRead
				if !stillOpen {
					close(r.Messaging.User1From)
					break
				}
			} else {
				log.Print("message from user role 1 with Token '" + r.User0.Token + "': '" + string(message) + "'.")
				r.Messaging.User1From <- message
			}
		}
	}
	log.Print("WebSocketReader room = " + r.OwnNumber.String() + ", role = " + role.String() + " correctly completed.")
	return
}

func (r *Room) WebSocketWriter(role RoleId) {
	if role == 0 {
	consistentMessageSending0:
		for message := range r.Messaging.User0To {
			for {
				err := r.User0.Connection.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					_, stillOpen := <-r.Recovery.User0IsAvailableWrite
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
		for message := range r.Messaging.User1To {
			for {
				err := r.User1.Connection.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					_, stillOpen := <-r.Recovery.User1IsAvailableWrite
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
