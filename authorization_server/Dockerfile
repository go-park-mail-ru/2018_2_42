# Ручная сборка на случай отладки:
# sudo docker build . --tag 'authorization_server' && \
# sudo docker push 'olegschwann/authorization_server';

# Ручной запуск сервиса авторизации:
# sudo mkdir --parents --mode=a+rwx '/var/www/media/images' && \
# sudo docker run \
# --name 'authorization' \
# --network 'rpsarena-net' \
# --volume "/var/www/media/images":"/var/www/media/images" \
# --detach \
# --rm \
# 'olegschwann/authorization_server':latest;

FROM golang:1.11
# GOPATH='/go'

# Скачиваем зависимости. В общем случае проще использовать
# go get -v ./...
# но тут удобнее оптимизировать кештрование docker.
RUN go get -v \
    'github.com/pkg/errors' \
    'github.com/spf13/pflag' \
    'github.com/mailru/easyjson' \
    'github.com/lib/pq' ;

# копируем исходники
COPY '.' "${GOPATH}/src/github.com/go-park-mail-ru/2018_2_42/authorization_server"

# Компилируем сервер,
# создаём папку под аватарки.
RUN go build -o '/go/bin/authorization_server' \
    'github.com/go-park-mail-ru/2018_2_42/authorization_server' && \
    mkdir --parents --mode=a+rwx '/var/www/media';

# При запуске контейнера запускаем сервер.
CMD ["/go/bin/authorization_server", "--postgres-path", "postgres://postgres:@database:5432/postgres?sslmode=disable", "--listening-port", "8080"]
