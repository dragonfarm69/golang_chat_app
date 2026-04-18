export function Setting() {
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
            style={{
              maxWidth: "200px",
              marginTop: "8px",
              padding: "8px",
              borderRadius: "4px",
              background: "var(--color-bg-dark)",
              color: "white",
              border: "1px solid var(--color-dark-blue)",
            }}
          >
            <option>Dark</option>
            <option>Light</option>
            <option>System</option>
          </select>
        </div>
        <div className="profile-info-item">
          <label className="profile-info-label">Notifications</label>
          <div style={{ marginTop: "8px", color: "white" }}>
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
