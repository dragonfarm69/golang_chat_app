# Chat App

A real-time chat application that allows users to participate in **multiple chat rooms simultaneously**. Built with a Go WebSocket backend, a React + Tauri desktop frontend, and PostgreSQL for persistence, with Keycloak handling authentication.

---

## Features

- **Multi-room chatting** — Join and chat in multiple rooms at the same time
- **Real-time messaging** — Powered by WebSocket connections via Gorilla WebSocket
- **Room management** — Create, join, and leave rooms on the fly
- **User authentication** — Secure login and registration through Keycloak (OAuth2)
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
                                                  │   │    ├── Hub (Room)│
                                                  │   │    └── ...      │
                                                  └───┼─────────────────┘
                                                      │
                                          ┌───────────┴───────────┐
                                          │                       │
                                    ┌─────▼─────┐         ┌──────▼──────┐
                                    │ PostgreSQL │         │  Keycloak   │
                                    │  Database  │         │  Auth (8081)│
                                    └────────────┘         └─────────────┘
```

---

## Tech Stack

| Layer          | Technology                                                    |
| -------------- | ------------------------------------------------------------- |
| **Frontend**   | React 18, TypeScript, Vite, React Router, @dnd-kit            |
| **Desktop**    | Tauri 2                                                       |
| **Backend**    | Go (net/http), Gorilla WebSocket, rs/cors                     |
| **Database**   | PostgreSQL 18, Prisma ORM (schema management), pgx (Go driver)|
| **Auth**       | Keycloak 26 (OAuth2 / OpenID Connect)                         |
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
│   │   ├── Components/      # Reusable UI components
│   │   ├── Context/         # Auth & User context providers
│   │   ├── Hook/            # Custom hooks (WebSocket)
│   │   ├── Modules/         # Client & message modules
│   │   └── Payload/         # Message payload types
│   └── src-tauri/           # Tauri native shell (Rust)
├── prisma/
│   ├── schema.prisma        # Database schema
│   ├── init_db.sql          # Initial DB migration script
│   └── seed.ts              # Database seeder
├── docker/
│   └── docker-keycloak.yaml # Docker Compose for PostgreSQL + Keycloak
└── package.json             # Root dependencies (Prisma, pg)
```

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

Create a `.env` file in the `docker/` directory with the required variables:

```env
DATABASE_PORT=5432
DATAB_BASE_NAME=app_main
DATABASE_USERNAME=your_db_user
DATABASE_PASSWORD=your_db_password
KEYCLOAK_USERNAME=admin
KEYCLOAK_PASSWORD=admin
```

### 2. Initialize the Database

```bash
psql -U <your_db_user> -d app_main -f prisma/init_db.sql
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

### 4. Start the Frontend (Dev Mode)

```bash
cd front/chat-app-frontend
npm install
npm run tauri dev
```

---

## API Endpoints

| Method | Endpoint         | Description                          |
| ------ | ---------------- | ------------------------------------ |
| GET    | `/ws`            | WebSocket connection (`?hub=&username=`) |
| GET    | `/newhub`        | Create a new chat room               |
| GET    | `/hublist`       | List all available room IDs          |
| GET    | `/join`          | Join a room (`?hub=&client=`)        |
| GET    | `/disconnect/`   | Disconnect from a room (`?hub=&client=`) |
| POST   | `/register`      | Register a new user                  |

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