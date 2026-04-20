import { useNavigate } from "react-router-dom";
import { commands } from "../../bindings";
import { useChatData } from "../../Context/DataContext";

export function Logout() {
  const nagivate = useNavigate();
  const { clearData } = useChatData();

  //clear all data in indexDB
  async function handleLogout() {
    await commands.logout().then(() => {
      clearData();
      nagivate("/authentication");
    });
  }

  return (
    <>
      <div className="profile-banner"></div>
      <h2 style={{ padding: "15px 28px", margin: 0, color: "white" }}>
        Are you sure you want to logout?
      </h2>
      <div className="logout-container">
        <button
          onClick={handleLogout}
          style={{ backgroundColor: "red", color: "white" }}
        >
          Yes
        </button>
        <button>No</button>
      </div>
    </>
  );
}
