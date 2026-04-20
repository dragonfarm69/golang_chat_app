import "./App.css";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import ErrorPage from "./Pages/Error/errorPage";
import CallBackPage from "./Pages/Authentication/CallBack";
import HomePage from "./Pages/MainPage/Home";
import { AuthProvider } from "./Context/AuthContext";
import { ProtectedRoute } from "./Components/protectedRoute";
import MainAuthenticationPage from "./Pages/Authentication/Main";
import { ChatDataProvider } from "./Context/DataContext";
import { UserProvider } from "./Context/userContext";
import { ThemeProvider } from "./Context/ThemeContext";

function App() {
  return (
    <ThemeProvider>
      <BrowserRouter>
        <AuthProvider>
          <UserProvider>
            <Routes>
              <Route
                path="/"
                element={
                  <ProtectedRoute>
                    <ChatDataProvider>
                      <HomePage />
                    </ChatDataProvider>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/authentication"
                element={<MainAuthenticationPage />}
              />
              <Route path="/callback" element={<CallBackPage />} />
              <Route path="*" element={<ErrorPage />} />
            </Routes>
          </UserProvider>
        </AuthProvider>
      </BrowserRouter>
    </ThemeProvider>
  );
}

export default App;
