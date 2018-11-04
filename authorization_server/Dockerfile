# Для сборки
# sudo docker build . --tag authorization_server
# потом для запуска
# sudo docker run --network="host" --detach --volume "/var/www/media/images":"/var/www/media/images" authorization_server

FROM golang:1.11
# GOPATH='/go'
# копируем исходники
COPY '.' 'authorization_server'
# копируем конфиг в папку с исполняемым файлом.
COPY './config/main.json5' 'authorization_server/main.json5'
# папки под аватарки создаём
RUN \
mkdir --parents --mode=a+rwx '/var/www/media' ;
# компилируем сервер
RUN \
cd authorization_server/ && \
go build -mod=vendor
# При запуске контейнера запустить сервер
CMD '/authorization_server/authorization_server'
# сделать порт доступным.
EXPOSE 8080
