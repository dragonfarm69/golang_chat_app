import { useState } from "react";
import "./App.css";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import ErrorPage from "./Pages/Error/errorPage";
import CallBackPage from "./Pages/Authentication/CallBack";
import HomePage from "./Pages/MainPage/Home";
import { AuthProvider } from "./Context/AuthContext";
import { ProtectedRoute } from "./Components/protectedRoute";
import MainAuthenticationPage from "./Pages/Authentication/Main";

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <HomePage />
              </ProtectedRoute>
            }
          />
          <Route path="/authentication" element={<MainAuthenticationPage />} />
          <Route path="/callback" element={<CallBackPage />} />
          <Route path="*" element={<ErrorPage />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  );
}

export default App;