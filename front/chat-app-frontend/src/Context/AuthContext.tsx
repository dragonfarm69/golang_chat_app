import { createContext, useContext, useEffect, useState } from "react";
import { ReactNode } from "react";
import { Store } from "@tauri-apps/plugin-store";
import { invoke } from "@tauri-apps/api/core";
import { commands } from "../bindings";
import { useNavigate } from "react-router-dom";
interface AuthContextType {
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (code: string, verifier: string) => Promise<void>;
  logout: () => Promise<void>;
  getAccessToken: () => Promise<string | null>;
}

const authContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [store, setStore] = useState<Store | null>(null);
  const navigate = useNavigate()

  useEffect(() => {
    const checkTokenExpired = async () => {
      try {
        console.log("checking...");
        const result = await commands.checkLoginStatus();
        if (result.status === "ok") {
          if(result.data === true) {
            setIsAuthenticated(result.data)
            navigate("/callback")
          }
        }
      } catch (e) {
        console.log("Error when trying to check token expired", e);
      } finally {
        setIsLoading(false)
      }
    };

    checkTokenExpired()
  }, []);

  const checkAuth = async (storeInstance: Store) => {
    try {
      const authStatus = await commands.checkAuth();
      if (authStatus.status === "ok") {
        setIsAuthenticated(authStatus.data);
      } else {
        throw new Error("Authentication failed");
      }
    } catch (e) {
      console.log("error when trying to check authentication: ", e);
      await clearToken(storeInstance);
      // await clearToke
    } finally {
      setIsLoading(false);
    }
  };

  const clearToken = async (storeInstance: Store) => {
    await storeInstance.delete("access_token");
    await storeInstance.delete("refresh_token");
    // await storeInstance.delete("") //more stuff
  };

  const login = async (access_token: string, refresh_token: string) => {
    // console.log("Login success: ", loginSuccess);
    try {
      const loginSuccess = await commands.finalizeLogin(
        access_token,
        refresh_token,
      );
      console.log("Login success: ", loginSuccess);

      if (loginSuccess) {
        setIsAuthenticated(true);
      } else {
        setIsAuthenticated(false);
      }

      // setUser(userInfo)
    } catch (e) {
      console.log("Error when trying to login: ", e);
      return;
    }
  };

  const logout = async () => {
    if (!store) {
      console.error("store not initialized");
    }
    try {
      await invoke("logout");
      setIsAuthenticated(false);
    } catch (e) {
      console.log("Error when trying to logout: ", e);
      return;
    }
  };

  const getAccessToken = async (): Promise<string | null> => {
    try {
      const token = await invoke<string>(`get_data_from_keyring`, {
        name: "access_token",
      });
      return token;
    } catch (e) {
      console.log("Error when trying to get access token: ", e);
      return null;
    }
  };

  return (
    <authContext.Provider
      value={{
        isAuthenticated,
        isLoading,
        login,
        logout,
        getAccessToken,
      }}
    >
      {children}
    </authContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(authContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return context;
};
