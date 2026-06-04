# 🚗 Ride-App — Nepal Ride Sharing Backend

> A production-grade ride-sharing backend built in Go, designed for low-bandwidth networks (2G/3G), Nepal payment gateways, and a phased 2.5-year build plan.

---

## 📌 Project Status

**Current Phase: Phase 0 → Phase 1 (Foundation)**
**Started:** May 2026
**Target Completion:** ~2.5 years (4–5 hrs/week)

| Module | Status |
|---|---|
| Environment + Docker | ✅ Done |
| Database Schema (all tables) | ✅ Done |
| Auth — OTP + JWT + Refresh | ✅ Done |
| Rider Profile CRUD | ✅ Done |
| Driver Profile CRUD | 🔄 In Progress |
| Wallet + Ledger | 🔄 Next |
| Fare Calculation Engine | ⬜ Pending |
| Ride Request + Matching | ⬜ Pending |
| MQTT Real-time Layer | ⬜ Pending |
| React Native Mobile App | ⬜ Pending |

---

## 🏗️ Architecture

### Tech Stack
| Layer | Technology | Why |
|---|---|---|
| Language | Go | Fast, low memory, great for APIs |
| Database | PostgreSQL + PostGIS | Geospatial queries for location |
| Cache + Sessions | Redis | OTP storage, rate limiting, refresh tokens |
| Real-time | MQTT (Mosquitto) | Low-bandwidth protocol, ideal for 2G/3G |
| Containerization | Docker Compose | Simulates 3-server setup on single PC |
| CI/CD | GitHub Actions | Auto-deploy to AWS EC2 |
| Hosting | AWS EC2 | Current dev/staging server |

### Server Architecture (Production Target)
```
PC1 — Go API Server (port 8080)
PC2 — PostgreSQL + Redis
PC3 — MQTT Broker (Mosquitto)
```
During development, all three are simulated via Docker Compose on a single machine.

### Folder Structure
```
ride-app/
├── cmd/server/          # Entry point
├── internal/
│   ├── auth/            # OTP, JWT, middleware
│   ├── users/           # Rider profile
│   ├── drivers/         # Driver profile + vehicle
│   ├── rides/           # Ride lifecycle + fare engine
│   ├── wallet/          # Ledger + balance
│   ├── payments/        # eSewa, Khalti, cash
│   └── safety/          # SOS, incidents
├── pkg/
│   ├── db/              # PostgreSQL connection pool
│   ├── cache/           # Redis connection
│   ├── mqtt/            # MQTT client
│   └── response/        # Standard JSON response helpers
├── migrations/          # SQL migrations (up + down)
├── docker/              # Docker Compose + Mosquitto config
└── docs/                # Wireframes + system design
```

---

## 🗄️ Database Schema

### Tables
| Table | Purpose |
|---|---|
| `users` | All users (riders + drivers), phone-based auth |
| `wallets` | One wallet per user, balance tracking |
| `wallet_ledger` | Double-entry ledger, immutable transaction log |
| `rider_profiles` | Rider-specific data, emergency contacts, preferences |
| `driver_profiles` | Driver data, license, location, online status |

### Key Design Decisions
- **Balance in paisa (integer)** — no floating point money bugs
- **Double-entry ledger** — balance = SUM of all ledger entries, never updated directly
- **Soft deletes** on users (`deleted_at` column)
- **PostGIS extension** enabled for geospatial ride matching
- **Idempotency keys** on wallet_ledger — prevents double charges

---

## 🔐 Auth System

Phone-based OTP authentication. No passwords.

### Flow
```
POST /v1/auth/otp/send     → Generate OTP, store in Redis (5min TTL)
POST /v1/auth/otp/verify   → Verify OTP → create user if new → return JWT pair
POST /v1/auth/refresh      → Refresh token → new access token
```

### Token Design
- **Access token** — JWT, 15 min expiry, contains `user_id` + `role`
- **Refresh token** — UUID, 30 days, stored in Redis
- **Rate limiting** — max 3 OTP requests per phone per hour

### Roles
| Value | Role |
|---|---|
| 1 | Rider |
| 2 | Driver |
| 3 | Admin |

### Redis Keys
```
otp:{phone}           → "123456"      TTL: 300s
otp_rate:{phone}      → "3"           TTL: 3600s
refresh:{uuid}        → user_id       TTL: 2592000s
```

---

## 💳 Wallet System

### Design
- Every financial event = one new ledger row
- Balance never stored directly — always calculated as `SUM(amount_paisa)`
- Negative amounts = debits, positive = credits
- All operations require idempotency key

### Transaction Types
| Type | Description |
|---|---|
| `topup` | Rider adds money (eSewa, Khalti, cash agent) |
| `ride_payment` | Deducted after ride completion |
| `refund` | Reversed after dispute |
| `cancellation_fee` | Charged on late cancellation |

### Limits
- Min top-up: NPR 50 (5000 paisa)
- Max wallet balance: NPR 10,000 (1,000,000 paisa)
- Max 3 top-ups per day

---

## 🚕 Ride Lifecycle (Planned — Phase 1)

### States
```
REQUESTED → SEARCHING → MATCHED → DRIVER_ARRIVED → IN_PROGRESS → COMPLETED → PAID
                ↓
            CANCELLED
```

### Driver Matching Algorithm
- **Geohash bucketing** (precision 7, ~150m cells)
- Driver locations stored in Redis as geohash sets
- On ride request: check rider's cell + 8 neighboring cells
- Filter: online + no active ride
- Sort: by actual distance
- Offer to nearest driver first (15s timeout, then next)
- Fallback: expand to precision 6 (~1.2km) if no drivers found

### Fare Engine
- Server-side only — client cannot influence fare
- Based on: base fare + per_km rate + per_min rate + zone multiplier
- Returns fare range (min/max), never a fixed price
- Surge pricing logic planned for Phase 1

---

## 📡 Real-time Layer (Planned — Phase 2)

### MQTT Topic Design
```
ride/{ride_id}/state       → ride state changes (QoS 1)
driver/{driver_id}/loc     → location delta updates (QoS 0)
rider/{rider_id}/notify    → push notifications (QoS 1)
```

### Location Updates
- Delta compression: only send change from last known position
- ~20 bytes per update vs ~50 bytes for full coordinates
- Sent every 3–5 seconds on movement
- Cached in Redis for instant access

### Reconnect Handling
- Full state sync via `GET /v1/rides/:id` on reconnect
- Vector clock versioning for conflict resolution
- Redis SETNX for ride lock (prevents duplicate driver assignment)

---

## 📱 Mobile App (Planned — Phase 3)

- React Native (Android first)
- Mapbox for maps (vector tiles, cacheable)
- Offline-first design
- Foreground service for background location (Android)
- Low data mode: driver dots only (no names/photos over 2G)

### Payment Gateways (Nepal)
- eSewa
- Khalti
- Cash (driver confirms receipt)

---

## 🗺️ 2.5 Year Roadmap

| Phase | Timeline | Goal |
|---|---|---|
| Phase 0 | Months 1–2 | Single PC setup, Docker, auth, schema |
| Phase 1 | Months 3–13 | Ride engine, fare, payments, security, tests |
| Phase 2 | Months 14–21 | MQTT real-time, location, state sync, notifications |
| Phase 3 | Months 22–27 | React Native rider + driver apps |
| Phase 4 | Month 28–30 | Monitoring, pilot launch, 20-user test |

**Pace:** 4–5 hours/week. Each phase has buffer built in.

---

## 🚀 Running Locally

### Prerequisites
- Docker + Docker Compose
- Go 1.22+
- `migrate` CLI tool

### Setup
```bash
# Clone repo
git clone https://github.com/IHSantosh/ride-app
cd ride-app

# Copy environment file
cp .env.example .env

# Start infrastructure (PostgreSQL, Redis, MQTT)
cd docker && docker compose up -d

# Run migrations
migrate -path ./migrations -database "postgres://rideapp:localdev123@127.0.0.1:5432/rideapp?sslmode=disable" up

# Start API with hot reload
cd .. && air
```

### Verify
```bash
curl localhost:8080/health
# {"status":"ok","service":"ride-app","version":"0.0.7"}
```

### API Endpoints (Current)
```
POST   /v1/auth/otp/send          Send OTP to phone number
POST   /v1/auth/otp/verify        Verify OTP, returns JWT
POST   /v1/auth/refresh           Refresh access token

GET    /v1/rider/profile          Get rider profile (protected)
PATCH  /v1/rider/profile/update   Update rider profile (protected)
```

---

## ⚠️ Edge Cases Handled

- Double-tap ride request → idempotency key deduplicates
- App crash during payment → webhook reconciliation
- Rider offline after match → state sync on reconnect
- GPS unavailable → manual pin placement fallback
- Wallet insufficient mid-booking → auto-switch to cash
- OTP SMS not delivered → resend with 60s cooldown
- Map tiles not loaded → cached tiles + text-mode fallback

---

## 📝 Notes

- `.env` is never committed — see `.env.example`
- All money values in **paisa** (1 NPR = 100 paisa)
- Development mode returns OTP in response body (no real SMS)
- Default country code: `+977` (Nepal)
- Designed and tested for 2G/3G network conditions

---

*Built by santoshih — HTB Academy Top 10% · Second semester CS student · Nepal*
