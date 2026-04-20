import { useState } from "react";
import { SideBar } from "./SideBar";
import { ProfileCard } from "./ProfileCard";
import { Setting } from "./Setting";
import { Bot } from "./Bot";
import { Logout } from "./Logout";

interface OptionPopUpProps {
  onClose: () => void;
}

export function OptionPopUp({ onClose }: OptionPopUpProps) {
  const [currentOption, setCurrentOption] = useState<String | null>("profile");

  return (
    <div className="popup-overlay">
      <div className="profile-container">
        <SideBar
          currentOption={currentOption}
          setCurrentOption={setCurrentOption}
        />
        <div className="profile-card">
          <button className="profile-back-btn" onClick={onClose}>
            ← Back
          </button>
          {currentOption === "profile" && <ProfileCard />}
          {currentOption === "bot" && <Bot />}
          {currentOption === "settings" && <Setting />}
          {currentOption === "logout" && <Logout />}
        </div>
      </div>
    </div>
  );
}
