# Для сборки
# sudo docker build . --tag authorization_server
# потом для запуска
# sudo docker run --network="host" --detach --volume "/var/www/media/images":"/var/www/media/images" authorization_server

FROM golang:1.11
# GOPATH="/go"
# копируем исходники
COPY '.' '/go/src/authorization_server'
# скачиваем все зависимости
RUN \
go get -v ./... &&  \
# компилируем сервер
go build -o '/go/bin/authorization_server' authorization_server && \
# папки под аватарки создаём
mkdir --parents --mode=a+rwx '/var/www/media' ;
# При запуске контейнера запустить сервер
CMD "/go/bin/authorization_server"
# сделать порт доступным.
EXPOSE 8080
