import { useTheme, Theme } from "../../Context/ThemeContext";

export function Setting() {
  const { theme, setTheme } = useTheme();

  return (
    <>
      <div className="profile-banner">
        <h2 style={{ padding: "15px 28px", margin: 0, color: "white" }}>
          Settings
        </h2>
      </div>
      <div
        className="profile-info-grid"
        style={{ display: "flex", flexDirection: "column", gap: "20px" }}
      >
        <div className="profile-info-item">
          <label className="profile-info-label">Theme</label>
          <select
            value={theme}
            onChange={(e) => setTheme(e.target.value as Theme)}
            className="theme-selector"
          >
            <option value="dark">🌙 Dark</option>
            <option value="light">☀️ Light</option>
            <option value="system">💻 System</option>
          </select>
        </div>
        <div className="profile-info-item">
          <label className="profile-info-label">Notifications</label>
          <div style={{ marginTop: "8px", color: "var(--color-text-primary)" }}>
            <input
              type="checkbox"
              id="notif-sound"
              defaultChecked
              style={{ marginRight: "10px" }}
            />
            <label htmlFor="notif-sound">Play Sound</label>
          </div>
        </div>
      </div>
    </>
  );
}
