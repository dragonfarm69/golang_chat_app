export const SideBarOptions = [
  {
    id: "profile",
    name: "Profile",
    icon: "👤",
  },
  {
    id: "bot",
    name: "My bot",
    icon: "👤",
  },
  {
    id: "settings",
    name: "Settings",
    icon: "⚙️",
  },
  {
    id: "logout",
    name: "Logout",
    icon: "🚪",
  },
];

interface SideBarProps {
  currentOption?: string | null | String;
  setCurrentOption?: (option: string) => void;
}

export function SideBar({ currentOption, setCurrentOption }: SideBarProps) {
  return (
    <div className="profile-side-bar">
      {SideBarOptions.map((option) => (
        <button 
          key={option.id} 
          className="side-bar-option"
          style={{ background: currentOption === option.id ? "rgba(255, 255, 255, 0.1)" : undefined }}
          onClick={() => {
            if (setCurrentOption) setCurrentOption(option.id);
          }}
        >
          {option.icon} {option.name}
        </button>
      ))}
    </div>
  );
}
