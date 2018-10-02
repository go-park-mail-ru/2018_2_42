FROM golang:1.11

# GOPATH="/go"

# копируем исходники
COPY "." "/go/src/"

# скачиваем все зависимости
RUN ["go", "get", "-v", "./..."]

# компилируем сервер
RUN ["go", "build", "-o", "/go/bin/authorization_server", "authorization_server"]

# При запуске контейнера запустить сервер
CMD "/go/bin/authorization_server"

# сделать порт доступным.
EXPOSE 8080