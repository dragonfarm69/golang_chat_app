import { useState } from "react";
import { SideBar } from "./SideBar";
import { ProfileCard } from "./ProfileCard";
import { Setting } from "./Setting";
import { Bot } from "./Bot";

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
        </div>
      </div>
    </div>
  );
}
