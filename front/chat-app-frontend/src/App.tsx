import { useState } from "react";
import "./App.css";
// import ChatPage from "./Pages/ChatRoom/chatMain";
import LoginPage from "./Pages/Authentication/LoginPage";
import RegisterPage from "./Pages/Authentication/RegisterPage";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import ErrorPage from "./Pages/Error/errorPage";
import CallBackPage from "./Pages/Authentication/CallBack";
import HomePage from "./Pages/MainPage/Home";
import { AuthProvider } from "./Context/AuthContext";
import { ProtectedRoute } from "./Components/protectedRoute";

function App() {
  return (
    // <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route
            path="/"
            element={
              // <ProtectedRoute>
                <HomePage />
              // </ProtectedRoute>
            }
          />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/callback" element={<CallBackPage />} />
          <Route path="*" element={<ErrorPage />} />
        </Routes>
      </BrowserRouter>
    // </AuthProvider>
  );
}

export default App;