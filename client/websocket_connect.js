const socket = new WebSocket("ws://localhost:3000/websocket")

socket.onopen = function(event) {
    console.log("Websocket connected")
    socket.send("server working")
}

socket.onmessage = function(event) {
    console.log("Received: ", event.data)
}

socket.onclose = function(event) {
    console.log("Websocket is closed")
}

socket.onerror = function(error) {
    console.log("Websocket error: ", error)
}