// Create a WebSocket connection to the server
const socket = new WebSocket("ws://localhost:3001/ws");

// Event listener for when connection is open
socket.addEventListener("open", (event) => {
  console.log("WebSocket connection opened");
});

// Event listener for receiving messages from the server
socket.addEventListener("message", (event) => {
  const data = JSON.parse(event.data);

  switch (data.action) {
    case "room-message":
      const messagesDiv = document.getElementById("messages");
      const newMessage = document.createElement("div");
      newMessage.textContent = `${data.data.member_id} to ${data.data.room_id}:  ${data.data.message}`;
      messagesDiv.appendChild(newMessage);
  }
});

// Function to send message to the WebSocket server
function sendMessage() {
  const input = document.getElementById("messageInput");
  const message = input.value;

  if (message !== "") {
    // Send the message to the WebSocket server
    // socket.send(message);
    socket.send(
      JSON.stringify({
        action: "message-room",
        data: {
          room_id: "public",
          member_id: "JohnCena",
          message: message,
        },
      }),
    );

    // Display the message in the chat
    const messagesDiv = document.getElementById("messages");
    const newMessage = document.createElement("div");
    newMessage.textContent = "You: " + message;
    messagesDiv.appendChild(newMessage);

    // Clear the input field
    input.value = "";
  }
}

// Event listener for WebSocket close event
socket.addEventListener("close", () => {
  console.log("WebSocket connection closed");
});

// Event listener for WebSocket error event
socket.addEventListener("error", (error) => {
  console.error("WebSocket error:", error);
});

function joinRoom(room_id) {
  if (room_id !== "") {
    // Send the message to the WebSocket server
    // socket.send(message);
    socket.send(
      JSON.stringify({
        action: "join-room",
        data: {
          room_id,
          joiner_id: "JohnCena",
        },
      }),
    );
  }
}
