import { createContext, useContext, useEffect, useState } from "react";
import { ReactNode } from "react";
import { Store } from "@tauri-apps/plugin-store";
import { invoke } from "@tauri-apps/api/core";
import UserInfo from "../Modules/clientModule";

interface AuthContextType {
    isAuthenticated: boolean,
    isLoading: boolean,
    user: UserInfo | null;
    login: (code: string, verifier: string) => Promise<void>;
    logout: () => Promise<void>
    getAccessToken: ()=>Promise<string|null>
}

const authContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({children}: {children: ReactNode}) => {
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const [isLoading, setIsLoading] = useState(false)
    const [store, setStore] = useState<Store | null>(null);
    const [user, setUser] = useState<UserInfo | null>(null);

    useEffect(()=> {
        initStore()
    }, [])

    const initStore = async () => {
        try {
            const storeInst = await Store.load("auth.json")
            setStore(storeInst)
            await checkAuth(storeInst)
        }
        catch (e) {
            console.log("error when trying to init store: ", e)
            setIsLoading(false)
        }
    }

    const checkAuth = async (storeInstance: Store) => {
        try {
            const authStatus = await invoke<boolean>("checkAuth")
            console.log("current auth status: ", authStatus)
            setIsAuthenticated(authStatus)
        }
        catch (e) {
            console.log("error when trying to check authentication: ", e)
            await clearToken(storeInstance)
            // await clearToke
        }
        finally {
            setIsLoading(false)
        }
    }

    const clearToken = async (storeInstance: Store) => {
        await storeInstance.delete("access_token")
        await storeInstance.delete("refresh_token")
        // await storeInstance.delete("") //more stuff
    }

    const login = async (access_token: string, refresh_token: string) => {
        if(!store) {
            console.log("store not initialized")
        }
        try {
            const userInfo = await invoke<UserInfo>('login', {
                access_token: access_token,
                refresh_token: refresh_token
            })

            setUser(userInfo)
            setIsAuthenticated(true)
        }
        catch(e) {
            console.log("Error when trying to login: ", e)
            return
        }
    }

    const logout = async () => {
        if(!store) {
            console.log("store not initialized")
        }
        try {
            await invoke("logout")
            setIsAuthenticated(false)
            setUser(null)
        }
        catch(e) {
            console.log("Error when trying to logout: ", e)
            return
        }
    }

    const getAccessToken = async (): Promise<string| null> => {
        try {
            const token = await invoke<string>(`get_data_from_keyring`, {name: "access_token"})
            return token
        }
        catch(e) {
            console.log("Error when trying to get access token: ", e)
            return null
        }
    }

    return (
        <authContext.Provider value={{
            isAuthenticated,
            isLoading,
            user,
            login,
            logout,
            getAccessToken,
        }}>
            {children}
        </authContext.Provider>
    )
}

export const useAuth = () => {
    const context = useContext(authContext)
    if (!context) {
        throw new Error('useAuth must be used within AuthProvider');
    }
    return context;
}