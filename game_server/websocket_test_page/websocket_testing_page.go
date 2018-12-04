package websocket_test_page

import "net/http"

func WebSocketTestPage(w http.ResponseWriter, r *http.Request) {
	// language=HTML // Enable "IntelliLang" plugin if syntax highlighting did not work.
	html := []byte(`
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
	background-color: #2e3436;
	font-family: "Arial", sans-serif;
	color: #babdb6;
	font-size: 1.3rem;
}
main {
    width: 40rem;
    display: flex;
    flex-direction: column;
}
button {
    display: flex;
	cursor: pointer;
}
button::-moz-focus-inner {
  border: 0;
}
textarea {
    display: flex;
    min-height: 15rem;
	resize: vertical;
	font-family: "DejaVu Sans Mono", monospace;
}
.decoration {
	margin: 0.2rem 0 0.2rem 0;
	background-color: #272c2d;
	border-radius: 0.5rem;
	color: inherit;
	border: none;
	font-size: inherit;
	padding: 0.2rem;
}
    </style>
</head>
<body>
<main>
    <button id="connect" class="decoration">Connect</button>
    <button id="disconnect" class="decoration">Disconnect</button>
    <textarea id="sent-data" class="decoration" placeholder="Отправляемые через websocket данные."></textarea>
	<button id="send" class="decoration">Send</button>
    <textarea id="incoming-data" class="decoration" placeholder="Присылаемые сервером данные."></textarea>
</main>
<script defer type="application/javascript">
"use strict";

let socket;

document.querySelector("#connect").addEventListener("click", (event) => {
    // очистка поля вывода 
    document.querySelector("#incoming-data").value = "";
    // устанавливаем cookie со случайной строкой.
    let d = new Date();
    d.setDate(d.getDate()+1);
    document.cookie = 
        "sessionid="+Math.round(Math.random()*2**32).toString()+"; "+
        "path=/; "+
        "expires="+d.toUTCString()+";"; 

    socket = new WebSocket("ws://"+location.host+"/game/v1/entrypoint");

    socket.addEventListener("open", (event) => {
        document.querySelector("#incoming-data").value += "// Open socket\n";
    });

    socket.addEventListener("close", (event) => {
        document.querySelector("#incoming-data").value += 
        "// Close socket " + (event.wasClean ? "clean" : "suddenly") + ". " + 
        "Code: " + event.code + " cause: " + event.reason + "\n";
    });

    socket.addEventListener("message", (event) => {
        document.querySelector("#incoming-data").value += 
        event.data + "\n";
    });

    socket.addEventListener("error", (error) => {
        document.querySelector("#incoming-data").value += 
        "// Error: " + error.message + "\n";
    });
});

document.querySelector("#disconnect").addEventListener("click", (event) => {
    socket.close();
});

document.querySelector("#send").addEventListener("click", (event) => {
    socket.send(document.querySelector("#sent-data").value);
});
</script>
</body>
</html>
`)
	w.Header().Add("Content-Type", "text/html;charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(html)
	_ = r.Body.Close()
	return
}
