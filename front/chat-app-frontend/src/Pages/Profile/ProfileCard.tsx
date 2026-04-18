import { useUser } from "../../Context/userContext";

export function ProfileCard() {
  const { userData } = useUser();

  if (!userData) {
    return (
      <div className="profile-loading">
        <div className="profile-loading-spinner" />
        <span>Loading profile…</span>
      </div>
    );
  }

  // Format dates nicely
  const formatDate = (iso: string | null | undefined) => {
    if (!iso) return "—";
    try {
      return new Date(iso).toLocaleDateString(undefined, {
        year: "numeric",
        month: "long",
        day: "numeric",
      });
    } catch {
      return iso;
    }
  };

  // Derive a two-letter avatar fallback from the username
  const initials = userData.username
    .split(/\s+/)
    .slice(0, 2)
    .map((w) => w[0]?.toUpperCase() ?? "")
    .join("");

  return (
    <>
      {/* Avatar */}
      <div className="profile-avatar-wrapper">
        {userData.avatar_url ? (
          <img
            className="profile-avatar-img"
            src={userData.avatar_url}
            alt={userData.username}
          />
        ) : (
          <div className="profile-avatar-fallback">{initials}</div>
        )}
        <span className="profile-online-dot" title="Online" />
      </div>

      {/* Name + edit button */}
      <div className="profile-header-row">
        <div>
          <h2 className="profile-username">{userData.username}</h2>
          <span className="profile-badge">Member</span>
        </div>
        <button className="profile-edit-btn">Edit Profile</button>
      </div>

      <div className="profile-divider" />

      {/* Info grid */}
      <div className="profile-info-grid">
        <div className="profile-info-item">
          <span className="profile-info-label">Email</span>
          <span className="profile-info-value">{userData.email}</span>
        </div>
        <div className="profile-info-item">
          <span className="profile-info-label">Member Since</span>
          <span className="profile-info-value">
            {formatDate(userData.created_at)}
          </span>
        </div>
        <div className="profile-info-item">
          <span className="profile-info-label">Last Updated</span>
          <span className="profile-info-value">
            {formatDate(userData.updated_at)}
          </span>
        </div>
      </div>
    </>
  );
}
