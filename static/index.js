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
        clone.querySelector(".routine").innerText = sh.stopped ? "STOPPED" : sh.routine;
        // clone.querySelector(".icon").src = `img/ships/${sh.type}.png`;
        if(sh.waitingForHttp){
            clone.querySelector(".state").innerText = "Waiting for HTTP";
        }
        if(sh.asleepUntil){
            let asleepUntil = new Date(sh.asleepUntil);
            clone.querySelector(".state").innerText = `Waiting for ${Math.floor((asleepUntil-now)/1000)} seconds`;
        }

        if(sh.stoppedReason){
            clone.querySelector(".state").innerText = sh.stoppedReason;
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
    log.innerText = logMessages.join("\n")
}


let mapLocations = [];
let mapXOffset = 0;
let mapYOffset = 0;
let mapScale = 1;
let mapPanSpeed = 1;
let mapZoomSpeed = 0.001;

const mapIcons = {
    "PLANET": "🌍",
    "GAS_GIANT": "☀️",
    "MOON": "🌕",
    "ORBITAL_STATION": "🛰️",
    "JUMP_GATE": "🚪",
    "ASTEROID": "*",
    "ENGINEERED_ASTEROID": "+",
    "ASTEROID_BASE": "🏢",
    "NEBULA": "💨",
    "DEBRIS_FIELD": "::",
    "GRAVITY_WELL": "🔽",
    "ARTIFICIAL_GRAVITY_WELL": "⏬",
    "FUEL_STATION": "⛽",
}


async function populateMap(){
    let req = await fetch("/waypoints")
        .then(res => res.json())
    req.forEach((waypoint) => {
        mapLocations.push({
            x: waypoint.waypointData.x,
            y: waypoint.waypointData.y,
            icon: mapIcons[waypoint.waypointData.type] || "?",
        })
    })
}


function getCanvasCoords(x, y){
    return [(x* mapScale - mapXOffset), (y* mapScale - mapYOffset)];
}

async function initMap(){
    await populateMap()

    let canvasContainer = document.getElementById("map")
    let canvas = document.getElementById("mapCanvas");

    canvas.style.width ='100%';
    canvas.style.height='100%';
    canvas.width  = canvas.offsetWidth;
    canvas.height = canvas.offsetHeight;

    let isDragging = false;
    let lastMouseX = 0;
    let lastMouseY = 0;
    canvas.onmousedown = (e) => {
        e.preventDefault();
        e.stopPropagation();
        lastMouseX = e.clientX;
        lastMouseY = e.clientY;
        isDragging = true;
    }
    canvas.onmouseup = (e) => {
        e.preventDefault();
        e.stopPropagation();
        isDragging = false;
    }
    canvas.onmousemove = (e) => {
        if(!isDragging)return;
        e.preventDefault();
        e.stopPropagation();
        let xDiff = lastMouseX - e.clientX;
        let yDiff = lastMouseY - e.clientY;
        lastMouseX = e.clientX;
        lastMouseY = e.clientY;
        mapXOffset += xDiff * mapPanSpeed;
        mapYOffset += yDiff * mapPanSpeed;

        drawMap();
    }

    canvas.onmousewheel = (e) => {
        e.preventDefault();
        e.stopPropagation();
        mapScale += e.wheelDelta * mapZoomSpeed;
        drawMap();
    }



    drawMap()



}

function drawMap(){
    let canvas = document.getElementById("mapCanvas");
    let ctx = canvas.getContext("2d");

    ctx.fillStyle = "black";
    ctx.fillRect(0, 0, canvas.width, canvas.height);
    ctx.fillStyle = "white";
    ctx.font = `${mapScale*10}px serif`
    for (let i = 0; i < mapLocations.length; i++){
        let [x, y] = getCanvasCoords(mapLocations[i].x, mapLocations[i].y);
        ctx.fillText(mapLocations[i].icon, x, y)//, 10 * mapScale, 10 * mapScale)
    }
}

let viewingMap = false;
function toggleViewMap() {
    viewingMap = !viewingMap;
    document.getElementById("shipState").style.display = viewingMap ? "none" : null;
    document.getElementById("map").style.display = !viewingMap ? "none" : null;
}



initMap().then(toggleViewMap)
