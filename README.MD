# Chat App

A real-time chat application that allows users to participate in **multiple chat rooms simultaneously**. Built with a Go WebSocket backend, a React + Tauri desktop frontend, and PostgreSQL for persistence, with Keycloak handling authentication.

---

## Features

- **Multi-room chatting** — Join and chat in multiple rooms at the same time
- **Real-time messaging** — Powered by WebSocket connections via Gorilla WebSocket
- **Room management** — Create, join, and leave rooms on the fly
- **User authentication** — Secure login via Keycloak (OAuth2 PKCE flow)
- **Secure token storage** — Access & refresh tokens stored in the OS keyring (via `keyring` crate)
- **Direct messages** — Send private messages between users
- **Threaded replies** — Reply to specific messages with parent message support
- **Room invitations** — Invite users to rooms with shareable invite codes
- **Role-based access** — Room members can be owners, admins, moderators, or members
- **Drag & drop UI** — Rearrangeable room tabs with `@dnd-kit`
- **Desktop app** — Packaged as a native desktop application with Tauri

---

## Architecture

```
┌─────────────────────┐        WebSocket         ┌─────────────────────┐
│                     │ ◄──────────────────────►  │                     │
│   React + Tauri     │        HTTP/REST          │   Go Server (:8080) │
│   Frontend          │ ◄──────────────────────►  │                     │
│                     │                           │   ┌─── Hub Manager  │
└─────────────────────┘                           │   │    ├── Hub (Room)│
         │                                        │   │    ├── Hub (Room)│
         │ Tauri IPC (commands)                   │   │    └── ...      │
         ▼                                        └───┼─────────────────┘
┌─────────────────────┐                              │
│   Rust / Tauri      │              ┌───────────────┴───────────┐
│   Shell (lib.rs)    │              │                           │
│   ├── OS Keyring    │        ┌─────▼─────┐             ┌──────▼──────┐
│   ├── JWKS Validate │        │ PostgreSQL │             │  Keycloak   │
│   └── WS Client     │        │  Database  │             │  Auth (8081)│
└─────────────────────┘        └────────────┘             └─────────────┘
```

---

## Tech Stack

| Layer          | Technology                                                    |
| -------------- | ------------------------------------------------------------- |
| **Frontend**   | React 18, TypeScript, Vite, React Router, @dnd-kit            |
| **Desktop**    | Tauri 2                                                       |
| **Backend**    | Go (net/http), Gorilla WebSocket, rs/cors                     |
| **Database**   | PostgreSQL 18, Prisma ORM (schema management), pgx (Go driver)|
| **Auth**       | Keycloak 26 (OAuth2 PKCE), `jsonwebtoken` (RS256 / JWKS)      |
| **Keyring**    | OS keyring via `keyring` crate (secure credential storage)    |
| **Infra**      | Docker Compose                                                |

---

## Project Structure

```
.
├── server/                  # Go backend
│   ├── main.go              # HTTP routes & server entry point
│   ├── server.go            # Hub & HubManager (room management)
│   ├── client.go            # WebSocket client & message handling
│   ├── init.go              # DB pool & Keycloak config initialization
│   ├── user_management.go   # User registration (Keycloak + DB)
│   └── message_management.go# Message persistence
├── front/chat-app-frontend/ # React + Tauri frontend
│   ├── src/
│   │   ├── Pages/           # Route pages (Home, ChatRoom, Auth, Error)
│   │   │   └── Authentication/
│   │   │       └── CallBack.tsx  # PKCE OAuth2 callback & login finalization
│   │   ├── Components/      # Reusable UI components
│   │   ├── Context/         # Auth & User context providers
│   │   ├── Hook/            # Custom hooks (WebSocket)
│   │   ├── Modules/         # Client & message modules
│   │   └── Payload/         # Message payload types
│   └── src-tauri/           # Tauri native shell (Rust)
│       └── src/lib.rs       # Tauri commands: keyring, JWKS validation, WS
├── prisma/
│   ├── schema.prisma        # Database schema
│   ├── init_db.sql          # Initial DB migration script
│   └── seed.ts              # Database seeder
├── docker/
│   └── docker-keycloak.yaml # Docker Compose for PostgreSQL + Keycloak
└── package.json             # Root dependencies (Prisma, pg)
```

---

## Authentication Flow

The app uses **OAuth2 Authorization Code flow with PKCE** (Proof Key for Code Exchange), handled natively inside the Tauri/Rust shell:

1. **Login initiated** — The frontend generates a PKCE `code_verifier` and `code_challenge`, stores the verifier in `sessionStorage`, then redirects the user to Keycloak's authorization endpoint.
2. **Keycloak callback** — After successful login, Keycloak redirects back to the app's `/callback` route with an authorization `code` in the URL.
3. **Token exchange** (`CallBack.tsx` → `fetch_token`) — The Rust `fetch_token` command sends the `code` + PKCE `verifier` to Keycloak's token endpoint and receives `access_token` and `refresh_token`.
4. **Secure storage** (`finalize_login`) — Both tokens are saved to the **OS keyring** under the `chat-app-service` namespace via the `keyring` crate, keeping them out of `localStorage` or `sessionStorage`.
5. **Token validation** (`checkAuth`) — The access token is read from the keyring, its header is decoded to extract the `kid`, and it is validated against Keycloak's **JWKS** public keys (RS256), checking issuer and audience.
6. **User info fetch** (`fetch_account_info`) — Account details are fetched from Keycloak's userinfo endpoint, enriched with profile data from the Go backend (`/fetch_user_info`), and cached in Tauri's persistent store (`cache_data.json`).
7. **Logout** (`logout`) — Both `access_token` and `refresh_token` credentials are deleted from the OS keyring.

### Tauri Auth & Keyring Commands

| Command                 | Description                                              |
| ----------------------- | -------------------------------------------------------- |
| `fetch_token`           | Exchanges auth code + PKCE verifier for tokens           |
| `finalize_login`        | Saves access & refresh tokens to the OS keyring          |
| `checkAuth`             | Reads token from keyring, validates JWT via JWKS (RS256) |
| `fetch_account_info`    | Fetches user profile from Keycloak + backend DB          |
| `get_data_from_keyring` | Reads a named credential from the OS keyring             |
| `logout`                | Deletes access & refresh tokens from the OS keyring      |

---

## Getting Started

### Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/) & npm
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- [Rust](https://www.rust-lang.org/tools/install) (for Tauri)

### 1. Start Infrastructure (PostgreSQL + Keycloak)

```bash
cd docker
docker compose -f docker-keycloak.yaml up -d
```

Create a `.env` file in the `docker/` directory:

```env
DATABASE_PORT=5432
DATAB_BASE_NAME=app_main
DATABASE_USERNAME=your_db_user
DATABASE_PASSWORD=your_db_password
KEYCLOAK_USERNAME=admin
KEYCLOAK_PASSWORD=admin
```

### 2. Initialize the Database

> **Note:** All Prisma commands must be run from the **root of the project**, where `package.json` and `prisma.config.ts` are located.

```bash
npm install
```

Create a `.env` file in the root directory:

```env
DATABASE_URL=postgres://<user>:<password>@localhost:5432/app_main
```

Run the initial SQL migration, then generate the Prisma client and seed:

```bash
psql -U <your_db_user> -d app_main -f prisma/init_db.sql
npx prisma generate
npx prisma db seed
```

### 3. Start the Go Backend

Create a `.env` file in the `server/` directory:

```env
DB_URL=postgres://<user>:<password>@localhost:5432/app_main
SCHEMA_NAME=chat
ADMIN_ID=<keycloak_client_id>
ADMIN_SECRET=<keycloak_client_secret>
KEYCLOAK_TOKEN_URL=http://localhost:8081/realms/<realm>/protocol/openid-connect/token
KEYCLOAK_API_URL=http://localhost:8081/admin/realms/<realm>/users
KEYCLOAK_ACCOUNT_LOGIN_URL=http://localhost:8081/realms/<realm>/protocol/openid-connect/auth
```

```bash
cd server
go run .
```

The server will start on `http://localhost:8080`.

### 4. Configure Tauri Environment Variables

These are compiled into the Rust binary via `env!()` macros. Create a `.env` file in `front/chat-app-frontend/src-tauri/`:

```env
BACKEND_REGISTER_URL=http://localhost:8080/register
REDIRECT_URL=<your-tauri-deep-link-callback-url>
TOKEN_URL=http://localhost:8081/realms/<realm>/protocol/openid-connect/token
CLIENT_ID=<your-keycloak-client-id>
BACKEND_URL=http://localhost:8080
```

### 5. Start the Frontend (Dev Mode)

```bash
cd front/chat-app-frontend
npm install
npm run tauri dev
```

---

## API Endpoints

| Method | Endpoint              | Description                                        |
| ------ | --------------------- | -------------------------------------------------- |
| GET    | `/ws`                 | WebSocket connection (`?user_id=`)                 |
| GET    | `/newhub`             | Create a new chat room                             |
| GET    | `/hublist`            | List all available room IDs                        |
| GET    | `/join`               | Join a room (`?hub=&client=`)                      |
| GET    | `/disconnect/`        | Disconnect from a room (`?hub=&client=`)           |
| POST   | `/register`           | Register a new user via Keycloak                   |
| GET    | `/fetch_rooms`        | Fetch rooms for a user (`?user_id=`)               |
| GET    | `/fetch_room_message` | Fetch messages for a room (`?room_id=`)            |
| GET    | `/fetch_user_info`    | Fetch user profile by username (`?username=`)      |

### WebSocket Message Format

```json
{
  "hubId": "<room-uuid>",
  "content": "Hello, world!"
}
```

---

## Database Schema

The app uses a `chat` schema in PostgreSQL with the following tables:

- **users** — User profiles (username, email, avatar, status)
- **rooms** — Chat rooms with owner, privacy settings, and descriptions
- **room_members** — Many-to-many relationship with role-based access
- **messages** — Room messages with thread/reply support
- **direct_messages** — Private messages between two users
- **room_invitations** — Shareable invite codes with optional expiry and usage limits

---

## License

This project is open source. Feel free to use and modify it.