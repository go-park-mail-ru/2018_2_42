# Ручная сборка на случай отладки:
# sudo docker build . --tag 'game_server' && \
# sudo docker push 'olegschwann/game_server';

# Ручной запуск игры:
# sudo docker run \
# --name 'game' \
# --network 'rpsarena-net' \
# --detach \
# --rm \
# 'olegschwann/game_server':latest;

FROM golang:1.11

# GOPATH='/go'
# скачиваем все зависимости
RUN go get -v \
    'github.com/spf13/pflag'\
    'github.com/rs/zerolog/log'\
    'github.com/pkg/errors'\
    'github.com/mailru/easyjson'\
    'github.com/gorilla/websocket';

# копируем исходники
COPY '.' "${GOPATH}/src/github.com/go-park-mail-ru/2018_2_42/game_server/"

# компилируем сервер
RUN go build -o '/go/bin/game_server' 'github.com/go-park-mail-ru/2018_2_42/game_server';

# сделать порт доступным.
EXPOSE 8080

# При запуске контейнера запустить сервер

CMD ["/go/bin/game_server", "--listen-port", "8080"]