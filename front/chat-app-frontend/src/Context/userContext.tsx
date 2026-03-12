import {
  createContext,
  useContext,
  useState,
  Dispatch,
  SetStateAction,
} from "react";
import { UserInfo } from "../bindings";

interface UserContextType {
  userData: UserInfo | null;
  setUserData: Dispatch<SetStateAction<UserInfo | null>>;
  initializeUserContext: (user_info: UserInfo) => Promise<void>;
}

const UserContext = createContext<UserContextType | null>(null);

export const UserProvider = ({ children }: { children: React.ReactNode }) => {
  const [userData, setUserData] = useState<UserInfo | null>(null);

  const initializeUserContext = async (user_info: UserInfo) => {
    try {
      setUserData(user_info);
    } catch (e) {
      console.log(
        "An error has occured when trying to initialize user context: ",
        e,
      );
    }
  };

  return (
    <UserContext.Provider
      value={{ userData, setUserData, initializeUserContext }}
    >
      {children}
    </UserContext.Provider>
  );
};

export const useUser = () => {
  const context = useContext(UserContext);
  if (context === null) {
    throw new Error("Something is null");
  }
  return context;
};
