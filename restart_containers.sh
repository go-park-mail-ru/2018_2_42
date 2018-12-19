  go get github.com/fzipp/gocyclo &&
  go build -o "${GOPATH}/bin/gocyclo" 'github.com/fzipp/gocyclo' &&
  find -name *.go -exec "${GOPATH}/bin/gocyclo" -over 25  {} + &&
  sudo docker pull olegschwann/authorization_server:latest &&
  sudo docker kill authorization
  sudo docker run \
  --name 'authorization' \
  --network 'rpsarena-net' \
  --volume "/var/www/media/images":"/var/www/media/images" \
  --detach \
  --rm olegschwann/authorization_server:latest &&
  sudo docker pull olegschwann/game_server:latest && 
  sudo docker kill game
  sudo docker run \
  --name 'game' \
  --network 'rpsarena-net' \
  --detach \
  --rm olegschwann/game_server:latest 
