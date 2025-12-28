import { useState } from "react";
import "./App.css";
import ChatPage from "./Pages/ChatRoom/chatMain";
import JoinRoomPage from "./Pages/JoinRoom/joinMain";
import LoginPage from "./Pages/Authentication/LoginPage";
import RegisterPage from "./Pages/Authentication/RegisterPage";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import ErrorPage from "./Pages/Error/errorPage";
import CallBackPage from "./Pages/Authentication/CallBack";
import HomePage from "./Pages/MainPage/Home";

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

  // return (
  //   // <ChatMain />
  //   <JoinRoomPage onJoin={handleJoin}/>
  // );

  return (
    // <RegisterPage></RegisterPage>
    // <LoginPage />
    <BrowserRouter>    
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/join" element={<JoinRoomPage onJoin={handleJoin}/>} />
        <Route path="/callback" element={<CallBackPage />} />
        <Route path="*" element={<ErrorPage />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App;