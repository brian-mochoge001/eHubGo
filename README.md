# 🚀 eHubGo: The High-Performance Engine

![eHub Banner](https://images.unsplash.com/photo-1558494949-ef010cbdcc31?auto=format&fit=crop&w=1200&q=80)

eHubGo is the core backend engine powering the entire **eHub SuperApp Ecosystem**. Built with Go for maximum concurrency and performance, it orchestrates real-time services, high-volume ecommerce, and geospatial operations.

## 🏗️ Architectural Overview

eHubGo acts as the central nervous system for three distinct frontends:
*   **eHub Mobile**: Native React Native app for mobility and daily services.
*   **eHubWeb**: SvelteKit-powered marketplace for web consumers.
*   **eHubPortal**: SvelteKit-powered management console for vendors and staff.

### Core Systems
*   **Geospatial Coordinator**: Integrated with OSRM for road-aware ETA and driver discovery with 15-minute guard rails.
*   **E-Commerce Engine**: Production-grade product filtering using PostgreSQL JSONB for Amazon-like dynamic attributes.
*   **Identity Provider**: Hybrid auth supporting Firebase (Mobile) and custom JWT (Web/Portal) with strict RBAC (Role Based Access Control).
*   **Asynchronous Worker Pool**: High-throughput processing for notifications, stock management, and status updates.

## 🛠️ Technology Stack

| Layer | Technology |
| :--- | :--- |
| **Language** | Go (Golang) 1.26+ |
| **Framework** | Gin Gonic |
| **Database** | PostgreSQL (Neon) |
| **Caching** | Redis (Render) |
| **Auth** | Firebase Admin SDK + custom JWT/Bcrypt |
| **ORM/SQL** | SQLC (Type-safe SQL) |
| **Logging** | Uber-Zap |

## 📡 API Endpoints

### 🔐 Authentication
*   `POST /api/v1/auth/login` - Standard JWT login.
*   `POST /api/v1/auth/register` - Account creation with role assignment.

### 🛍️ E-Commerce & Marketplace
*   `GET /api/v1/products/filter` - Dynamic attribute-based search.
*   `GET /api/v1/categories` - Hierarchical category trees.
*   `POST /api/v1/checkout` - ACID-compliant stock reservation and order creation.

### 🚕 Mobility (eTaxi)
*   `GET /api/v1/drivers/nearby` - Road-distance sorted driver discovery.
*   `POST /api/v1/taxi/request` - Real-time trip coordination.

## 🚀 Deployment

The project is dockerized and optimized for **Render**:

```bash
# Local Development
docker-compose up

# Manual Build
go build -o main .
./main
```

## 📜 Database Management

We use `sqlc` for type-safe database interactions.
*   **Schema**: `schema.sql` defines the source of truth.
*   **Queries**: `query.sql` defines the optimized access patterns.
*   **Extensions**: `db/extensions.go` handles complex JSONB and search logic.

---
© 2026 Infinnity Developers. Built for the future of urban commerce.
