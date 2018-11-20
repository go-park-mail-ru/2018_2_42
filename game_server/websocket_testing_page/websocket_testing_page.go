package websocket_testing_page

import "net/http"

func WebSocketTestPage(w http.ResponseWriter, r *http.Request) {
	// language=HTML
	var html = []byte(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>WebSocket test</title>
    <link rel="icon" href="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAABHNCSVQICAgIfAhkiAAAAAtJREFUCJljYAACAAAFAAFiVTKIAAAAAElFTkSuQmCC">
    <style>
html {
    min-height: 100%;
}
body {
    margin: 0;
    min-height: 100%;
    display: flex;
    flex-grow: 1;
    justify-content: center;
    align-items: center;
}
main {
    width: 40rem;
    border: 1px solid gray;
    display: flex;
    flex-direction: column;
}
button {
    display: flex;
}
textarea {
    display: flex;
    min-height: 10rem;
}
    </style>
</head>
<body>
<main>
    <button id="connect">Connect</button>
    <button id="disconnect">Disconnect</button>
    <textarea id="text"></textarea>
    <button id="send">Send</button>
</main>
<script>
    let socket;

    document.querySelector("#connect").addEventListener("click", (event) => {
        socket = new WebSocket("ws://localhost:8081/game/v1/entrypoint?login=login&token=password");

        socket.addEventListener("open", (event) => {
            console.log("Open socket");
        });

        socket.addEventListener("close", (event) => {
            if (event.wasClean) {
                console.log("Close socket clean")
            } else {
                console.log("Close socket suddenly")
            }
            console.log('Code: ', event.code, ' cause: ', event.reason);
        });

        socket.addEventListener("message", (event) => {
            console.log("Get data: ", event.data);
        });

        socket.addEventListener("error", (error) => {
            console.log("Error: ", error.message);
        });
    });

    document.querySelector("#disconnect").addEventListener("click", (event) => {
        socket.close();
    });

    document.querySelector("#send").addEventListener("click", (event) => {
        let text = document.querySelector("#text").innerText;
        socket.send(text);
        console.log("send socket: ", text)
    });
</script>
</body>
</html>
`)
	w.Header().Add("Content-Type", "text/html;charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(html)
	r.Body.Close()
	return
}
