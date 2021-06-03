let messages = [];
let webSocket = null;
let open = null;
let authToken = null;
let guid = null;
// var token = '1';

async function login(e) {
    e.preventDefault();

    if (open) {
        console.error('WebSocket connection is already running')
        return
    }
    let username = e.target.elements.username.value;
    let password = e.target.elements.password.value;
    guid = e.target.elements.guid.value;

    if (!username)
        username = 'tester';
    if (!password)
        password = 'test';
    if (!guid)
        guid = '7e01f1c3-50bd-4db2-8031-323f466d7dc7'

    // Login a user
    await axios.post('http://stage.hsprod.tech:30033/user/login',
        {'username': username, 'password': password})
        .then((response) => {
            if (response.data.ok)
                authToken = response.data.data.token;
            else {
                authToken = null;
                document.querySelector('.login-info').innerHTML = 'Failed to authenticate - ' + response.data.error.message;
                document.querySelector('.login-info').style.backgroundColor = 'red';
            }
        })
        .catch((error) => {
            authToken = null;
            console.error('Failed to authenticate - ' + error)
        });

    let userId = null;
    if (authToken) {
        const id = await getId(authToken);
        if (id)
            userId = id;
        else
            return;
    } else
        return;

    document.querySelector('.login-info').innerHTML = 'Logged in as: <b>' + username + '</b><br/>User ID: <b>' + userId + '</b>';
    document.querySelector('.login-info').style.backgroundColor = 'green';
    webSocket = openNewSocketConnection();
}

async function getId(authToken) {
    return axios({
        method: 'POST',
        url: 'http://stage.hsprod.tech:30033/user',
        headers: {'Authorization': authToken},
        validateStatus: (status) => status === 200
    }).then(response => {
        if (response.data.ok)
            return response.data.data.user.id
        else
            return null;
    }).catch(error => {
        console.error('Failed get user ID - ' + error)
        return null;
    });
}

function openNewSocketConnection() {
    const wsPath = 'ws://127.0.0.1:8080/chat?&token=' + authToken + '&guid=' + guid;
    let socket = new WebSocket(wsPath);
    socket.onopen = onopen;
    socket.onclose = onclose;
    socket.onmessage = onmessage;
    socket.onerror = onerror;
    return socket;
}

String.prototype.paddingLeft = function (paddingLength) {
    const paddingValue = " ".repeat(paddingLength);
    return String(paddingValue + this).slice(-paddingLength);
};

String.prototype.paddingRight = function (paddingLength) {
    const paddingValue = " ".repeat(paddingLength);
    return String(paddingValue + this).slice(-paddingLength);
};

let onopen = event => {
    open = true;
    console.log('Opening connection');
};

let onmessage = event => {
    console.log('Got websocket message');
    console.log(JSON.parse(event.data));

    let newMessages = JSON.parse(event.data)
    newMessages = Array.isArray(newMessages) ? newMessages : [newMessages]
    messages = newMessages.concat(messages);

    let displayMessage = "<pre>";
    let i = 0;
    for (let msg of messages) {
        let meta = [msg.username.paddingLeft(10), msg.timestamp.substring(0, 19).paddingRight(20)].join('|')
        const stringMsg = `${meta}|${msg.text}\n`
        displayMessage += stringMsg
        i++;
        if (i >= 10)
            break;
    }
    ;
    displayMessage += "</pre>";

    document.getElementById('message-board').innerHTML = displayMessage
}

let onclose = event => {
    open = false;
    console.log('Closing connection');
};

let onerror = error => {
    open = false;
    console.log('Error with websocket connection');
};

function switchConnection() {
    console.log(open ? "Closing websocket connection" : "Opening websocket connection");
    if (open) {
        console.log(webSocket);
        messages = [];
        webSocket.close();
        webSocket = null;
    } else {
        webSocket = openNewSocketConnection();
    }
}

window.onload = () => {
    document.querySelector("#message-form").addEventListener("submit", (e) => {
        e.preventDefault();
        if (webSocket) {
            const inputs = document.getElementById("message-form").elements;
            const message = {
                "text": inputs["message"].value
            };
            inputs["message"].value = "";
            webSocket.send(JSON.stringify(message));
        } else {
            console.error("Socket connection is closed...");
        }
    });

    document.querySelector('#login-form').addEventListener('submit', (e) => login(e))
}