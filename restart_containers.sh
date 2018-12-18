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
