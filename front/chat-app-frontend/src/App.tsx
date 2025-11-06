import { useState } from "react";
import "./App.css";
import ChatPage from "./Pages/ChatRoom/chatMain";
import JoinRoomPage from "./Pages/JoinRoom/joinMain";

function App() {
  const [roomId, setRoom] = useState<string | null>(null);

  if (roomId) {
    return <ChatPage roomId={roomId} />;
  }

  return (
    // <ChatMain />
    <JoinRoomPage onJoin={setRoom}/>
  );
}

export default App;