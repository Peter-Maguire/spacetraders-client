const ws = new WebSocket(`ws://${location.host}/ws`)

const logMessages = [];

ws.onmessage = function (message) {
    const update = JSON.parse(message.data);
    switch(update.type){
        case "state":
            updateState(update.data);
            break;
        case "log":
            updateLog(update.data)
            break;

    }
}


function updateState({ship, http}){
    const ships = document.getElementById("shipState");
    const requests = document.getElementById("httpCalls");

    const shipTemplate = document.getElementById("shipTemplate")


    const now = new Date();
    ship.forEach((sh)=>{
        const clone = shipTemplate.content.cloneNode(true);
        const container = document.createElement("div")
        container.classList.add("ship", sh.name);
        clone.querySelector(".name").innerText = sh.name;

        // clone.querySelector(".icon").src = `img/ships/${sh.type}.png`;
        if(sh.waitingForHttp){
            clone.querySelector(".state").innerText = "Waiting for HTTP";
        }
        if(sh.asleepUntil){
            let asleepUntil = new Date(sh.asleepUntil);
            clone.querySelector(".state").innerText = `Waiting for ${Math.floor((asleepUntil-now)/1000)} seconds`;
        }

        if(sh.stopped){
            clone.querySelector(".routine").innerText = `STOPPED: ${sh.stoppedReason}`;
        }else{
            clone.querySelector(".routine").innerText = sh.routine;
        }

        container.appendChild(clone);
        const currentShip = ships.querySelector(`.${sh.name}`);
        if(currentShip) {
            currentShip.remove()
        }
        ships.appendChild(container);


    })
    requests.innerText = http.requests.map((req, i)=>{
        return `#${i+1} ${req.receivers}x [${req.priority}] ${req.method} ${req.path}`
    }).join("\n");
}

function updateLog(data){
    const log = document.getElementById("log");
    log.scrollTop = log.scrollHeight;
    logMessages.push(data);
    if(logMessages.length > 50)
        logMessages.shift()
    log.innerText = logMessages.join("")
}