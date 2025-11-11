import { createContext, useContext, useState, Dispatch, SetStateAction } from "react";
import UserInfo from "../Modules/clientModule";

interface UserContextType {
    user: UserInfo | null;
    setUser: Dispatch<SetStateAction<any>>;
}
const UserContext = createContext<UserContextType | null>(null)

export const UserProvider = ({children}: {children: React.ReactNode}) => {
    const [user, setUser] = useState(null);

    return (
        <UserContext.Provider value={{ user, setUser }}>
            {children}
        </UserContext.Provider>
    );
};

export const useUser = () => {
    const context = useContext(UserContext);
    if (context === null) {
        throw new Error("Something is null");
    }
    return context
}