# RTX Trading Engine (Trading-Engine2) - Project Overview

## 1. Executive Summary
RTX Trading Engine is a high-performance, multi-asset trading platform designed for microsecond-latency execution. It features a robust backend written in **Go** and a suite of modern frontend applications built with **React** and **Next.js**.

**Current Status**: The system routes 100% real market data from YOFX (no mock data).

## 2. Core Architecture

The project is divided into four main functional areas:

### 🏛️ Backend (`/backend`)
The core processing unit handling OMS, Risk, Routing, and FIX connectivity.
- **Language**: Go (Golang)
- **Entry Point**: `backend/cmd/server/main.go`
- **Key Modules**:
    - `fix/`: FIX Gateway (LMAX/YOFX connectivity)
    - `oms/`: Order Management System
    - `risk/`: Risk Engine (Margin & Equity)
    - `ws/`: WebSocket Hub for real-time data broadcasting

### 🖥️ Desktop Terminal (`/clients/desktop`)
The primary trading interface for end-users.
- **Framework**: React + TypeScript + Vite
- **Features**: Real-time charts, order entry, account monitoring.
- **Location**: `clients/desktop`

### 👨‍💼 Broker Admin (`/admin/broker-admin`)
Dashboard for brokers to manage their clients and settings.
- **Framework**: Next.js
- **Location**: `admin/broker-admin`

### 👑 Super Admin (`/admin/super-admin`)
Platform-wide control center for system administrators.
- **Framework**: Next.js
- **Location**: `admin/super-admin`

---

## 3. How to Run the Project

You can run each component in a separate terminal window.

### Backend Scope
```bash
cd backend
go run cmd/server/main.go
# Starts on: http://localhost:8080 (REST) & ws://localhost:8080/ws (WebSocket)
```

### Desktop Terminal
```bash
cd clients/desktop
npm run dev
# Starts on: http://localhost:5173
```

### Broker Admin
```bash
cd admin/broker-admin
npm run dev
# Starts on: http://localhost:3000
```

### Super Admin
```bash
cd admin/super-admin
npm run dev
# Starts on: http://localhost:3001
```

## 4. Key Data Sources
- **Market Data**: Direct FIX connection to YOFX.
- **Database**: SQL schema located in `backend/database`.

## 5. Directory Map
```text
Trading-Engine2/
├── backend/                # Go Backend
│   ├── cmd/server/         # Main Entry Point
│   ├── internal/           # Core Logic
│   └── fix/                # Connectivity
├── clients/
│   ├── desktop/            # React Trading Terminal
│   └── mobile/             # Mobile App Source
├── admin/
│   ├── broker-admin/       # Broker Dashboard (Next.js)
│   └── super-admin/        # Super Admin Dashboard (Next.js)
└── README.md               # Original Documentation
```
