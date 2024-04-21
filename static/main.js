const setUpWebsocket = (chatText) => {
    const socket = new WebSocket('ws://localhost:8081/ws')

    socket.onopen = () => {
        console.info('Websocket opened')
    }

    socket.onmessage = (event) => {
        console.log('On message data: ', event.data)
        const { username, text } = JSON.parse(event.data)
        const p = document.createElement('p')
        p.innerHTML = `<strong>${username}</strong>: ${text}`

        chatText.append(p)
        chatText.scrollTop = chatText.scrollHeight
    }

    socket.onerror = (error) => {
        console.error('Websocket error: ', error)
    }

    socket.onclose = (event) => {
        console.info('Websocket closed')
    }

    return socket
}

(function main() {
    const form = document.getElementById('input-form')
    const chatText = document.getElementById('chat-text')
    const socket = setUpWebsocket(chatText)

    form.onsubmit = (event) => {
        event.preventDefault();
        const username = document.getElementById("input-username");
        const text = document.getElementById("input-text");
        
        const chatMessage = JSON.stringify({ username: username.value, text: text.value })
        socket.send(chatMessage)

        text.value = ''
    }
})()
