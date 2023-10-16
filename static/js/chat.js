
const msgRequest = document.getElementById('msgRequest');
const sendBtn = document.getElementById('send');
const chatContainer = document.getElementById('chat-container');
const actionForm = document.getElementById('action-form');

function addMsg(text, msgClass) {
	let newMsg = document.createElement("div");
	newMsg.innerText = text;
	newMsg.classList.add(msgClass);
	chatContainer.appendChild(newMsg);
	chatContainer.scrollTop = chatContainer.scrollHeight;
}

function handleResponse(data) {
	console.log(data)
	if (data.success) {
		addMsg(data.response, "gpt")
	}
}


function makeRequest() {
	const msgText = msgRequest.value.trim();
	msgRequest.value = '';
	if (msgText == '') return;

	addMsg(msgText, "my")

    var http = new XMLHttpRequest();
    http.open("POST", "/chat", true);
    http.setRequestHeader("Content-type","application/x-www-form-urlencoded");
    var params = "message=" + msgText;
    http.send(params);
    http.onload = function() {
    	const response = JSON.parse(http.responseText);
        handleResponse(response);
    }


}

sendBtn.addEventListener('click', makeRequest);

