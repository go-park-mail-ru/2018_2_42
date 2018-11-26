# Для сборки
# sudo docker build . --tag chat_server ;
# потом для запуска
# sudo docker run --network="host" --detach chat_server ;

FROM golang:1.11
# GOPATH='/go'
# копируем исходники

# скачиваем все зависимости
RUN go get \
    'github.com/jackc/pgx' \
    'github.com/spf13/pflag' \
    'gopkg.in/inconshreveable/log15.v2' \
    'github.com/gorilla/websocket' \
    'github.com/mailru/easyjson' ;

COPY '.' "${GOPATH}/src/github.com/go-park-mail-ru/2018_2_42/chat_server"

# компилируем сервер
RUN go build -o '/go/bin/chat_server' 'github.com/go-park-mail-ru/2018_2_42/chat_server' ;
# При запуске контейнера запустить сервер
CMD '/go/bin/chat_server' --port=8001
# сделать порт доступным.
EXPOSE 8001
