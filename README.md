# Balkan Storage — Secure File Vault System
[Software Documentation](https://docs.google.com/document/d/1KfMVkXaGF-abqTyrB7nHL_WsPTMSaWiWWT23fCG1v34/edit?usp=sharing)

## Overview

Balkan Storage is a **secure, production-grade file vault system** designed as part of the **BalkanID Full Stack Engineering Intern — Capstone Hiring Task**.  
It enables users to **upload, organize, search, and share files** while ensuring **deduplication, controlled access, and storage efficiency**.

The project demonstrates backend engineering with **Go**, frontend design with **React + TypeScript**, robust **PostgreSQL** modeling, and containerized deployment with **Docker Compose**.

---

## Tech Stack

- **Backend:** Go (Golang), Gin framework
    
- **Frontend:** React.js + Next.js + TypeScript + TailwindCSS
    
- **Database:** PostgreSQL
    
- **API Layer:** REST (extendable to GraphQL)
    
- **Authentication:** JWT-based secure sessions
    
- **Containerization:** Docker + Docker Compose
    
- **UI/UX:** Glassmorphism, dark/light theme support, modern gradient palettes
    

---

## Core Features

### 1. File Deduplication

- Detects duplicate uploads using **SHA-256 hashing**.
    
- Prevents redundant storage by saving only references.
    
- Displays **per-user storage savings** from deduplication.
    

### 2. File Uploads

- **Single and multi-file uploads** supported.
    
- **Drag-and-drop uploads** from the frontend.
    
- **MIME type validation** to prevent renamed/malicious uploads.
    
- Progress indicators and animated UI feedback.
    

### 3. File Management & Sharing

- List and view files with full metadata: uploader, size, upload date, deduplication info.
    
- Organize files into **folders** (create, rename, move, delete).
    
- Sharing options:
    
    - **Public sharing** with a link.
        
    - **Private access** (visible only to owner).
        
    - **Role-based access control (RBAC)** for users and admins.
        
- Public files display **download counters**.
    
- Strict deletion rules:
    
    - Only uploader can delete.
        
    - Deduplication-aware deletion.
        

### 4. Search & Filtering

- Search by **filename**.
    
- Filters by **MIME type, size, date range, tags, uploader**.
    
- Supports **combined filters** for precision search.
    
- Optimized queries for **scalability**.
    

### 5. Rate Limiting & Quotas

- Per-user API rate limits (default: **2 calls/sec**).
    
- Per-user storage quotas (default: **10 MB**, configurable).
    
- Returns structured error codes for quota/limit violations.
    

### 6. Storage Statistics

- Displays **total usage, deduplicated savings, percentages**.
    
- Visualized with **animated charts and circular progress bars**.
    

### 7. Admin Panel

- Admins can upload/share files for users.
    
- Global file listing with uploader details.
    
- Access to **system statistics** and **download analytics**.
    

---

## Additional Features

- **Audit Logging**: Tracks uploads, downloads, deletions, and sharing activity.
    
- **Security**:
    
    - JWT authentication with expiry.
        
    - Passwords hashed with **bcrypt**.
        
    - CORS-protected API.
        
- **Scalability**: Modular backend with clear separation of concerns (auth, db, handlers, middleware).
    
- **Frontend UX Enhancements**:
    
    - **Glassmorphism design** (frosted glass effects, blur, gradients).
        
    - Dynamic purple-to-blue gradients for branding.
        
    - Dark/Light mode toggle with smooth transitions.
---
## System Architecture


- **Frontend**: Serves the UI, communicates with backend via REST API.
    
- **Backend**: Handles authentication, file operations, deduplication, and quotas.
    
- **Database**: Stores users, files, folders, audit logs, metadata.
    
- **Docker Compose**: Orchestrates `db`, `backend`, and `frontend` containers.

- **Frontend**: Serves the UI, communicates with backend via REST API.  
- **Backend**: Handles authentication, file operations, deduplication, and quotas.  
- **Database**: Stores users, files, folders, audit logs, metadata.  
- **Docker Compose**: Orchestrates `db`, `backend`, and `frontend` containers.  

---

## Setup & Installation  

### Prerequisites  
- [Docker Desktop](https://www.docker.com/products/docker-desktop) installed and running.  
- `.env` file in project root with environment variables:  

```env
# Database
DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=balkan_storage

# Backend
PORT=8080
JWT_SECRET=supersecret
STORAGE_PATH=/data/storage

# Frontend
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Steps

1. **Clone the repository**
    

`git clone <repo-url> cd vit-2026-capstone-internship-hiring-task-mahidharreddyg`

2. **Run with Docker Compose**
    

`docker compose up --build`

3. Access the services:
    

- Frontend: [http://localhost:3000](http://localhost:3000)
    
- Backend: [http://localhost:8080](http://localhost:8080)
    
- Database: localhost:5432
---
## API Documentation

- REST endpoints are structured under `/login`, `/signup`, `/files`, `/folders`, `/stats`, `/admin`, `/share`.
    
- Token-based authentication via **Authorization header**:
    
    `Authorization: Bearer <JWT>`
    
- Can be tested with **Postman** or **cURL**.
    
- (Future extension: OpenAPI Spec or GraphQL SDL).
    

---

## Design & Architecture Writeup

### Why Go + Gin?

- High performance, minimal footprint, great for APIs.
    
- Middleware support for authentication, logging, and rate limiting.
    

### Why PostgreSQL?

- Relational model fits well for users, files, folders, and audit logs.
    
- Strong support for indexing, full-text search, and JSON fields for metadata.
    

### Why JWT?

- Stateless, scalable authentication.
    
- Easy integration across frontend and backend.
    

### Why React + Next.js + TypeScript?

- **Type safety** (catch errors at build time).
    
- **Next.js** supports fast builds and SSR if needed.
    
- **Modern UI** with TailwindCSS and **Glassmorphism** for a polished UX.
    

### Architecture Principles

1. **Separation of Concerns**: Handlers, middleware, and DB layers are decoupled.
    
2. **Scalability**: Deduplication and quotas keep storage efficient.
    
3. **Security by Design**: JWT, bcrypt, CORS, rate limiting.
    
4. **Modern UX**: Responsive, visually appealing, intuitive interactions.
    

---

## Deliverables

- **Backend**: Go REST API (fully containerized).
    
- **Database**: PostgreSQL schema and migrations.
    
- **Frontend**: Modern UI with file management and sharing.
    
- **Deployment**: Docker Compose setup.
    
- **Documentation**: README, API docs, architecture overview.
---

## License

This project was developed as part of the BalkanID Full Stack Engineering Intern — Capstone Hiring Task.