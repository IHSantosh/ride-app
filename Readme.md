
raw
Readme · MD
# 🚗 Ride-App — Nepal Ride Sharing Backend
 
> A production-grade ride-sharing backend built in Go, designed for low-bandwidth networks (2G/3G), Nepal payment gateways, and a phased build plan.
 
---
 
## 📌 Project Status
 
**Current Phase: Phase 1 — Core Backend (In Progress)**
**Started:** May 2026
**Backend Prototype Target:** ~2 months from now
 
### Progress Today (June 7, 2026)
 
| Module | Status |
|---|---|
| Environment + Docker + CI/CD | ✅ Complete |
| Database Schema — all 5 tables | ✅ Complete |
| Auth — OTP + JWT + Refresh | ✅ Complete |
| Rider Profile CRUD | ✅ Complete |
| Driver Profile CRUD | ✅ Complete |
| Wallet + Double-entry Ledger | ✅ Complete |
| Fare Calculation Engine | 🔄 Next |
| Ride Request + State Machine | ⬜ Pending |
| Driver Matching (Geohash) | ⬜ Pending |
| MQTT Real-time Location | ⬜ Pending |
| Safety + SOS | ⬜ Pending |
| React Native Mobile App | ⬜ Pending |
 
---
 
## ⏱️ Realistic Time Estimate
 
At current pace (2–3 files/day, consistent sessions):
 
| What | Estimate |
|---|---|
| Fare engine | 1–2 sessions |
| Ride request + state machine | 3–4 sessions |
| Driver matching (geohash) | 2–3 sessions |
| Payment flow (cash + eSewa mock) | 2–3 sessions |
| MQTT location layer | 3–4 sessions |
| Safety + SOS | 1–2 sessions |
| API hardening + tests | 2–3 sessions |
| **Working backend prototype** | **~6–8 weeks** |
| React Native mobile app | +3–4 months |
| Full pilot-ready product | +2–3 months after app |
 
**Backend alone is 6–8 weeks away at current velocity.**
That means a working prototype by August 2026 is realistic — well within college final project timeline.
 
---
 
## 🏗️ Architecture
 
### Tech Stack
| Layer | Technology | Why |
|---|---|---|
| Language | Go | Fast, low memory, great for concurrent APIs |
| Database | PostgreSQL + PostGIS | Geospatial queries for ride matching |
| Cache + Sessions | Redis | OTP, rate limiting, refresh tokens, driver locations |
| Real-time | MQTT (Mosquitto) | Low-bandwidth protocol, ideal for 2G/3G Nepal |
| Containerization | Docker Compose | Simulates 3-server setup on single PC |
| CI/CD | GitHub Actions | Auto-deploy to AWS EC2 on push |
| Hosting | AWS EC2 | Dev/staging server |
 
### Server Architecture (Production Target)
```
PC1 — Go API Server        (port 8080)
PC2 — PostgreSQL + Redis
PC3 — MQTT Broker (Mosquitto)
```
During development, all three run via Docker Compose on a single machine.
 
### Folder Structure
```
ride-app/
├── cmd/server/          # Entry point — HTTP server, route wiring
├── internal/
│   ├── auth/            # OTP, JWT, middleware, refresh
│   ├── users/           # Rider profile CRUD
│   ├── drivers/         # Driver profile CRUD
│   ├── rides/           # Ride lifecycle + fare engine (next)
│   ├── wallet/          # Ledger + balance + topup
│   ├── payments/        # eSewa, Khalti, cash (planned)
│   └── safety/          # SOS, incidents (planned)
├── pkg/
│   ├── db/              # PostgreSQL connection pool (pgx)
│   ├── cache/           # Redis connection
│   ├── mqtt/            # MQTT client (planned)
│   └── response/        # Standard response helpers
├── migrations/          # SQL migrations (up + down)
├── docker/              # Docker Compose + Mosquitto config
└── docs/                # Wireframes + system design
```
 
---
 
## 🗄️ Database Schema
 
### Tables (All 5 Complete)
| Table | Purpose |
|---|---|
| `users` | All users (riders + drivers), phone-based auth, soft deletes |
| `wallets` | One wallet per user, balance in paisa |
| `wallet_ledger` | Immutable double-entry transaction log |
| `rider_profiles` | Emergency contacts, preferences, ratings |
| `driver_profiles` | License, vehicle, online status, location |
 
### Key Design Decisions
- **Balance in paisa (integer)** — no floating point money bugs
- **Double-entry ledger** — balance = SUM of all ledger entries, never updated directly
- **Idempotency keys** on wallet_ledger — prevents double charges
- **Soft deletes** on users — `deleted_at` column, data never lost
- **PostGIS** — geospatial ride matching queries
---
 
## 🔐 Auth System
 
Phone-based OTP. No passwords.
 
### Endpoints
```
POST /v1/auth/otp/send     → Generate OTP, store in Redis (5min TTL)
POST /v1/auth/otp/verify   → Verify OTP → create user + wallet → return JWT pair
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
 
---
 
## 💳 Wallet System
 
### Design
- Every financial event = one new immutable ledger row
- Balance calculated as `SUM(amount_paisa)` — never stored directly
- Negative = debit, positive = credit
- All operations require idempotency key (auto-generated if not provided)
### Transaction Types
| Type | Description |
|---|---|
| `topup` | Rider adds money (eSewa, Khalti, cash agent) |
| `ride_payment` | Deducted after ride completion |
| `refund` | Reversed after dispute |
| `cancellation_fee` | Charged on late cancellation |
 
### Limits
- Min top-up: NPR 50 (5,000 paisa)
- Max wallet balance: NPR 10,000 (1,000,000 paisa)
---
 
## 🚕 Ride System (Next — Being Built)
 
### State Machine
```
REQUESTED → SEARCHING → MATCHED → DRIVER_ARRIVED → IN_PROGRESS → COMPLETED → PAID
                ↓
            CANCELLED
```
 
### Driver Matching Algorithm
- **Geohash bucketing** precision 7 (~150m cells)
- Driver locations stored in Redis as geohash sets
- Check rider's cell + 8 neighboring cells
- Filter: online + no active ride
- Sort: actual distance
- 15s timeout per driver, then offer to next
- Fallback: expand to precision 6 (~1.2km) if no drivers found
### Fare Engine
- Server-side only — client cannot influence fare
- Base fare + per_km + per_min + zone multiplier
- Returns fare range (min/max), never fixed price
- All values in paisa
---
 
## 📡 Real-time Layer (Planned)
 
### MQTT Topics
```
ride/{ride_id}/state       → ride state changes     (QoS 1)
driver/{driver_id}/loc     → delta location updates (QoS 0)
rider/{rider_id}/notify    → push notifications     (QoS 1)
```
 
### Location Updates
- Delta compression — only send change from last known position
- ~20 bytes per update (vs ~50 bytes full coordinates)
- Cached in Redis for instant access
- Sent every 3–5 seconds on movement
---
 
## 🚀 Running Locally
 
### Prerequisites
- Docker + Docker Compose
- Go 1.22+
- `migrate` CLI
### Setup
```bash
git clone https://github.com/IHSantosh/ride-app
cd ride-app
cp .env.example .env
# Edit .env — set DB_HOST=127.0.0.1, REDIS_HOST=127.0.0.1
 
cd docker && docker compose up -d
migrate -path ./migrations -database "postgres://rideapp:localdev123@127.0.0.1:5432/rideapp?sslmode=disable" up
cd .. && air
```
 
### Verify
```bash
curl localhost:8080/health
# {"status":"ok","service":"ride-app","version":"0.0.9"}
```
 
---
 
## 📋 Current API Endpoints
 
```
GET    /health                        Service health check
 
POST   /v1/auth/otp/send             Send OTP to phone
POST   /v1/auth/otp/verify           Verify OTP → JWT tokens
POST   /v1/auth/refresh              Refresh access token
 
GET    /v1/rider/profile             Get rider profile
PATCH  /v1/rider/profile/update      Update rider profile
 
POST   /v1/driver/register           Register as driver
GET    /v1/driver/profile            Get driver profile
 
GET    /v1/wallet                    Get wallet + last 10 transactions
POST   /v1/wallet/topup              Add funds to wallet
```
 
---
 
## ⚠️ Known Edge Cases Handled
 
- Duplicate ride request → idempotency key deduplicates
- Wallet frozen → topup rejected with reason
- Max balance exceeded → topup rejected
- Driver profile already exists → 400 with clear error
- JWT expired → 401, use refresh token
- OTP rate limit → 429 after 3 attempts/hour
- Null GPS coordinates → pointer scan, returns 0
---
 
## 📝 Notes
 
- `.env` never committed — see `.env.example`
- All money in **paisa** (1 NPR = 100 paisa)
- Dev mode returns OTP in response (no real SMS)
- Default country code `+977` (Nepal)
- Encoding: JSON now, Protobuf planned for Phase 3 mobile
- Designed for 2G/3G network conditions
---
 
*Built by santoshih — HTB CPTS Top 10% · CS Student · Nepal*
*Repo: github.com/IHSantosh/ride-app*
