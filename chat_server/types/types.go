package types

import "github.com/mailru/easyjson"

// все типы, в формате которых нужно пересылать сообщения с клиентом.

// первый уровень
//easyjson:json
type Event struct {
	Method    string              `json:"method,required"`
	Parameter easyjson.RawMessage `json:"parameter,required"`
}

// "send"
type Message struct {
	From      *string  `json:"from"`          // логин пользователя
	To        *string  `json:"to"`            // логин пользователя
	Text      string   `json:"text,required"` // сообщение пользователя
	Reply     *Message `json:"reply"`         // пересылаемое сообщение, во вторую очередь
	Time      string   `json:"time"`          // в формате iso-8601
	Id        uint     `json:"id"`
	IsHistory bool     `json:"is_history"` // falst если посылает горутина получившая, true если из базы.
}

//easyjson:json
type Messages []Message

// "history"
//easyjson:json
type HistoryRequest struct {
	From   *string `json:"from"`   // логин пользователя
	To     *string `json:"-"`      // логин запрашивающего пользователя, на тороне сервера
	Before uint    `json:"before"` // id последнего сообщения, которое известно.
}

// Работа с "Before" параметром:
// знал  указал  пришло
//                1
//                2
//                3
//                4
// 5       5
// 6
// 7
// 8
// 9
