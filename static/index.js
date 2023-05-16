const ws = new WebSocket("ws://192.168.1.11:8080/ws")



const logMessages = [];


ws.onmessage = function (message) {
    const log = document.getElementById("log");
    const ships = document.getElementById("shipState");
    const requests = document.getElementById("httpCalls");
    const update = JSON.parse(message.data);
    console.log(update);
    switch(update.type){
        case "state":
            ships.innerText = update.data[0];
            requests.innerText = update.data[1];
            break;
        case "log":
            logMessages.push(update.data);
            if(logMessages.length > 50)
                logMessages.shift()
            log.innerText = logMessages.join("")

    }
}