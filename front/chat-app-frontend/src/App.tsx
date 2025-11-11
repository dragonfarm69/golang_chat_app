import { useState } from "react";
import "./App.css";
import ChatPage from "./Pages/ChatRoom/chatMain";
import JoinRoomPage from "./Pages/JoinRoom/joinMain";

function App() {
  const [roomId, setRoom] = useState<string | null>(null);
  const [clientName, setClientName] = useState<string | null>("");

  const handleJoin = (newRoomId: string, currentClientName: string) => {
    setRoom(newRoomId);
    setClientName(currentClientName)
  };

  if (roomId && clientName) {
    return <ChatPage roomId={roomId} clientName={clientName} />;
  }

  return (
    // <ChatMain />
    <JoinRoomPage onJoin={handleJoin}/>
  );
}

export default App;