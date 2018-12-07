# Для сборки
# sudo docker build . --tag authorization_server
# потом для запуска
# sudo docker run --network="host" --detach --volume "/var/www/media/images":"/var/www/media/images" authorization_server

FROM golang:1.11
# GOPATH='/go'
# скачиваем все зависимости
RUN go get -v \
    'github.com/spf13/pflag'\
    'github.com/rs/zerolog/log'\
    'github.com/pkg/errors'\
    'github.com/mailru/easyjson'\
    'github.com/gorilla/websocket'\
    'google.golang.org/grpc'\
    'github.com/golang/protobuf/proto';
# копируем исходники
COPY '.' "${GOPATH}/src/github.com/go-park-mail-ru/2018_2_42/game_server/"
# компилируем сервер
RUN go build -o '/go/bin/game_server' 'github.com/go-park-mail-ru/2018_2_42/game_server';
# сделать порт доступным.
EXPOSE 8080
# При запуске контейнера запустить сервер
CMD '/go/bin/game_server' --listen-port=8080
