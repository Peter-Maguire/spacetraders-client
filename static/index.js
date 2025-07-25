const logMessages = [];

let headerUpdateTimeout;

function connect() {

    const ws = new WebSocket(location.protocol === 'https:' ? `wss://${location.host}/ws` : `ws://${location.host}/ws`)

    ws.onmessage = function (message) {
        const update = JSON.parse(message.data);
        switch (update.type) {
            case "state":
                updateState(update.data);
                break;
            case "log":
                updateLog(update.data)
                break;

        }
    }

    clearTimeout(headerUpdateTimeout);
    headerUpdateTimeout = setInterval(updateHeader, 10000)
    updateHeader();

    ws.onerror = function (error) {
        ws.close();
    }

    ws.onclose = function () {
        updateLog("Connection lost... Reconnecting...")
        setTimeout(connect, 1000)
    }
}

connect();


let shipStates = [];
let waypoints = [];


function updateState({ship, http}){
    const ships = document.getElementById("shipState");
    const requests = document.getElementById("httpCalls");

    const shipTemplate = document.getElementById("shipTemplate")

    shipStates = ship;
    drawMap();

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
            clone.querySelector(".state").innerText = `Waiting for ${parseTime(asleepUntil)}`;
        }

        if(sh.stoppedReason){
            clone.querySelector(".state").innerText = sh.stoppedReason;
        }

        if(sh.fuel) {
            clone.querySelector(".fuel").innerText = `Fuel: ${sh.fuel.current}/${sh.fuel.capacity}\n`
        }
        if(sh.cargo){
            clone.querySelector(".inventory").innerHTML = `Cargo: ${sh.cargo.units}/${sh.cargo.capacity}</br>`
            clone.querySelector(".inventory").innerHTML += sh.cargo.inventory.map((cargo)=>{
                return `${cargo.symbol} x${cargo.units}`
            }).join("</br>");
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


function parseTime(asleepUntil){
    const now = new Date();
    let seconds = (asleepUntil-now)/1000;

    let d = Math.floor(seconds / (3600*24));
    let h = Math.floor(seconds % (3600*24) / 3600);
    let m = Math.floor(seconds % 3600 / 60);
    let s = Math.floor(seconds % 60);

    let dDisplay = d > 0 ? d + (d === 1 ? " day, " : " days, ") : "";
    let hDisplay = h > 0 ? h + (h === 1 ? " hour, " : " hours, ") : "";
    let mDisplay = m > 0 ? m + (m === 1 ? " minute, " : " minutes, ") : "";
    let sDisplay = s > 0 ? s + (s === 1 ? " second" : " seconds") : "";
    return dDisplay + hDisplay + mDisplay + sDisplay;
}

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
    waypoints = req;
}


function getCanvasCoords(x, y){
    return [(x* mapScale - mapXOffset), (y* mapScale - mapYOffset)];
}

async function updateHeader(){
    let agents = await fetch("/agent")
        .then(res => res.json())

    let contracts = await fetch("/contracts")
        .then(res => res.json())

    let agent = agents[Object.keys(agents)[0]];

    document.getElementById("credits").innerText = agent.credits.toLocaleString()+" credits"

     if(contracts){
         let contract = contracts[Object.keys(contracts)[0]];
         const deliverable = contract.terms.deliver[0];
         document.getElementById("contract").innerText = `${contract.type}: ${deliverable.unitsFulfilled}/${deliverable.unitsRequired} ${deliverable.tradeSymbol} for ${(contract.terms.payment.onAccepted+contract.terms.payment.onFulfilled).toLocaleString()}`
     }
}

async function initMap(){
    await populateMap()

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

function getWaypoint(symbol){
    return waypoints.find((w)=>w.waypoint === symbol);
}

let mapSystem = "X1-PS43";

function drawMap(){
    if(!viewingMap)return;
    let canvas = document.getElementById("mapCanvas");
    let ctx = canvas.getContext("2d");

    ctx.fillStyle = "black";
    ctx.fillRect(0, 0, canvas.width, canvas.height);
    ctx.fillStyle = "white";
    ctx.font = `${mapScale*10}px serif`

    waypoints.forEach(waypoint => {
        if(waypoint.system !== mapSystem)return;
        let icon = mapIcons[waypoint.waypointData.type] || "?";
        let [x, y] = getCanvasCoords(waypoint.waypointData.x, waypoint.waypointData.y);
        ctx.fillText(icon, x, y)//, 10 * mapScale, 10 * mapScale)
        ctx.font = `${mapScale*5}px serif`
        ctx.fillText(waypoint.waypoint, x-5, y+10)
        ctx.font = `${mapScale*10}px serif`
    })

    shipStates.forEach((ship)=>{
        if(ship.nav.systemSymbol !== mapSystem)return;
        if(ship.nav.status !== "IN_TRANSIT") {
            const waypoint = getWaypoint(ship.nav.waypointSymbol);
            if (!waypoint) {
                console.log(`Unable to find waypoint belonging to ship`, ship);
                return
            }
            let [x, y] = getCanvasCoords(waypoint.waypointData.x, waypoint.waypointData.y);
            drawShip(ctx, x, y, ship);
            return
        }

        const originWaypoint = getWaypoint(ship.nav.route.origin.symbol);
        let [oX, oY] = getCanvasCoords(originWaypoint.waypointData.x, originWaypoint.waypointData.y);
        const destWaypoint = getWaypoint(ship.nav.route.destination.symbol);
        let [dX, dY] = getCanvasCoords(destWaypoint.waypointData.x, destWaypoint.waypointData.y);
        const departedAt = new Date(ship.nav.route.departureTime);
        const arrivalAt = new Date(ship.nav.route.arrival);
        const now = new Date();
        const percentageComplete = (now-departedAt)/(arrivalAt-departedAt);

        ctx.strokeStyle = "red";
        ctx.beginPath();
        ctx.setLineDash([5, 15]);
        ctx.moveTo(oX, oY);
        ctx.lineTo(dX, dY);
        ctx.stroke();
        let [sX, sY] = interpolatePoint(originWaypoint.waypointData.x, originWaypoint.waypointData.y, destWaypoint.waypointData.x, destWaypoint.waypointData.y, percentageComplete);
        let [ssX, ssY] = getCanvasCoords(sX, sY)
        let heading = Math.atan2(dY - oY, dX - oX) + 0.8;
        drawShip(ctx, ssX, ssY, ship, heading);

    })
}

function drawShip(ctx, x, y, ship, heading = 0) {
    let fontSize = mapScale*5;
    if(heading !== 0){
        ctx.save();
        ctx.translate(x-(fontSize/2), y);
        ctx.rotate(heading);
        ctx.translate(-(x-fontSize/2), -y);
    }
    ctx.fillText("🚀", x, y-fontSize)
    if(heading !== 0){
        ctx.restore();
    }
    let lastFont = ctx.font;
    ctx.font = `${fontSize}px serif`
    ctx.fillText(ship.name, x, y+(fontSize*2))
    ctx.font = lastFont;

}

function interpolatePoint(x1, y1, x2, y2, t) {
    const x = x1 + (x2 - x1) * t;
    const y = y1 + (y2 - y1) * t;
    return [ x, y ];
}

let viewingMap = true;
function toggleViewMap() {
    viewingMap = !viewingMap;
    document.getElementById("shipState").style.display = viewingMap ? "none" : null;
    document.getElementById("map").style.display = !viewingMap ? "none" : null;
}



initMap().then(toggleViewMap)
