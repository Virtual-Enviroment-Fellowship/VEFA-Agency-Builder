<div align="center">

# 🏛️ VEFA Agency Builder (`v0.1.2`)
### High-Density, AI-Orchestrated Multi-Tenant VPS Architecture for Ubuntu 24.04 LTS

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![OS Support](https://img.shields.io/badge/Ubuntu-24.04%20LTS-E95420?style=for-the-badge&logo=ubuntu)](https://ubuntu.com/)
[![Web Server](https://img.shields.io/badge/Nginx-FastCGI%20Cache-009639?style=for-the-badge&logo=nginx)](https://nginx.org/)
[![Database](https://img.shields.io/badge/MariaDB-10.11%20LTS-003545?style=for-the-badge&logo=mariadb)](https://mariadb.org/)
[![PHP](https://img.shields.io/badge/PHP-8.3%20OPcache-777BB4?style=for-the-badge&logo=php)](https://www.php.net/)
[![Official Portal](https://img.shields.io/badge/VEFA-Official%20Portal-38BDF8?style=for-the-badge)](https://www.VEFA.club)

<p align="center">
  <b>VEFA™: Virtual Environment Fellowship of America</b><br>
  <i>Official Community & Standards: <a href="https://www.VEFA.club">www.VEFA.club</a></i><br>
  <sub>(VEFA Trademarked as of 2026 - All Rights Reserved)</sub>
</p>

---

<p align="center">
  <b>Host 200+ secure, high-performance client micro-sites on a single low-cost Hostinger VPS.</b><br>
  Consolidates PHP-FPM pools, eliminates monolith memory bloat, automates database isolation, and connects asynchronous AI tool queues without blocking web workers.
</p>

</div>

---

## 📑 Table of Contents
- [🏛️ System Architecture](#️-system-architecture)
- [🛡️ Enterprise VPS Security & Infrastructure](#️-enterprise-vps-security--infrastructure)
- [⚡ Extreme Performance & Resource Optimization](#-extreme-performance--resource-optimization)
- [🎛️ The ConsultDevin Client Dashboard](#️-the-consultdevin-client-dashboard)
- [🤖 Asynchronous AI Orchestration (MCP)](#-asynchronous-ai-orchestration-mcp)
- [💾 Database Architecture & Data Isolation](#-database-architecture--data-isolation)
- [🚀 Quick Start & Setup Wizard (.exe)](#-quick-start--setup-wizard-exe)
- [📁 Repository Structure](#-repository-structure)
- [🔒 Security & GitHub Best Practices](#-security--github-best-practices)
- [🌐 Community & Legal Notice](#-community--legal-notice)

---

## 🏛️ System Architecture

Traditional multi-site installations (such as running 200 separate WordPress instances or heavy Laravel monoliths) cripple low-tier VPS servers due to PHP-FPM worker exhaustion and duplicate memory overhead. 

**VEFA Agency Builder** solves this by routing all incoming traffic through a **single, highly optimized Nginx catch-all vHost** and a lightweight **Slim PHP Front Controller**.

```mermaid
flowchart TD
    subgraph Edge ["Edge & Browser Layer"]
        A[Client Domain 1: plumber-site.com] -->|HTTPS Port 443| N[Nginx Catch-All vHost]
        B[Client Domain 2: tutor-site.com] -->|HTTPS Port 443| N
        C[Client Domain 200: auto-care.com] -->|HTTPS Port 443| N
        A -.->|Client-Side Polling w/ Jitter| N
    end

    subgraph Security ["Hardened OS & Security Layer"]
        N --- F2B[Fail2Ban Intrusion Prevention]
        N --- UFW[UFW Firewall: 22, 80, 443, 8443]
        N --- RESTIC[Restic Automated Whole-Server Snapshots]
    end

    subgraph WebApp ["Single FPM Pool /var/www/master-app"]
        N -->|FastCGI unix socket| FC[Slim PHP Front Controller index.php]
        FC -->|1. Resolve Hostname| CDB[(db_agency_core)]
        FC -->|2. Dynamically Inject DB| TDB[(Isolated Tenant DB)]
        FC -->|3. Route to Industry Logic| TM[Trade Modules: /app/Trades/*]
        TM -->|4. Push Async AI Job| CDB
    end

    subgraph BackgroundAI ["Asynchronous Daemon Worker"]
        DW[worker.php (Systemd: agency-worker.service)] -->|Polls Pending Jobs| CDB
        DW -->|Async cURL POST w/ Idempotency Key| MCP[External MCP AI Orchestrator / Local LLM]
        MCP -->|Webhook Callback /api/webhooks/mcp-relay| FC
    end
```

---

## 🛡️ Enterprise VPS Security & Infrastructure

Built upon **CloudPanel** on a fresh **Ubuntu 24.04 LTS** base, providing enterprise-grade server controls with near-zero background RAM usage:

| Security Feature | Implementation Details |
| :--- | :--- |
| **Fail2Ban Intrusion Defense** | Actively bans malicious IP addresses targeting SSH (`Port 22`) and CloudPanel Admin (`Port 8443`) brute-force attacks. |
| **UFW Firewall Hardening** | Strict kernel-level packet filtering restricted exclusively to essential ports: `22` (SSH), `80` (HTTP), `443` (HTTPS), and `8443` (CloudPanel). |
| **Automated SSL Management** | Automated generation and automatic 90-day renewal of free **Let's Encrypt Wildcard & Multi-SAN SSL** certificates across all client domains. |
| **Restic Automated Backups** | Built-in Restic backup engine taking deduplicated, encrypted whole-server snapshot images to remote off-site storage. |
| **Isolated Linux Users** | CloudPanel automatically provisions isolated system users (`clp`) and permissions, ensuring cross-tenant file access is strictly prevented. |

---

## ⚡ Extreme Performance & Resource Optimization

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                        MEMORY & SPEED OPTIMIZATION MATRIX                       │
├────────────────────────────────┬────────────────────────────────────────────────┤
│ 🚀 Single PHP-FPM Pool         │ 1 consolidated worker pool instead of 200      │
│                                │ separate pools fighting for RAM.               │
├────────────────────────────────┼────────────────────────────────────────────────┤
│ 🧠 PHP 8.3 OPcache Bytecode    │ Pre-compiles and memorizes shared agency logic │
│                                │ in shared memory for sub-millisecond execution.│
├────────────────────────────────┼────────────────────────────────────────────────┤
│ ⚡ FastCGI Cache-Control       │ Nginx aggressive browser cache headers offload │
│                                │ bandwidth, avoiding CDN costs until ready.     │
├────────────────────────────────┼────────────────────────────────────────────────┤
│ 🌐 Central Master Asset Domain │ All 200 sites hotlink core JS/CSS/Fonts from   │
│                                │ `assets.youragency.com` for maximum caching.   │
└────────────────────────────────┴────────────────────────────────────────────────┘
```

---

## 🎛️ The ConsultDevin Client Dashboard

Upgrades and enhancements made to the shared core immediately roll out across all 200+ tenant micro-sites without touching individual site deployments.

* **Client Control Plane**: Every client receives access to their customized **ConsultDevin Dashboard** to manage appointments, volunteer rosters, and business data.
* **Universal Instant Bug Fixes**: Patching a security vulnerability or trade module in `/app/Trades/` or `/app/Core/` fixes it simultaneously for all active client websites.
* **Flexible AI Provider Integration**: Clients can choose between:
  - **Locally Hosted AI**: Zero-cost, privacy-first inference via LM Studio / Ollama connected over secure VPN.
  - **Paid API Key Integration**: Direct token billing via OpenAI, Claude, or Gemini API keys entered into the client settings.

---

## 🤖 Asynchronous AI Orchestration (MCP)

Third-party AI API calls can take 2 to 30 seconds to generate responses. If executed synchronously inside PHP, web server threads freeze and drop connections. 

**VEFA Agency Builder** utilizes an **Asynchronous Database-Backed Job Queue**:

1. **202 Accepted Response**: The client clicks "Generate AI Match/Schedule"; PHP writes the job to `ai_job_queue` and returns an immediate `202 Accepted` status with a polling URL.
2. **CLI Background Daemon (`worker.php`)**: A dedicated Systemd service (`agency-worker.service`) claims pending queue items and dispatches them asynchronously to the external Model Context Protocol (MCP) orchestrator via cURL.
3. **Webhook Relay Receiver (`/api/webhooks/mcp-relay`)**: When generation completes, the AI orchestrator posts results back to the secure webhook endpoint with Bearer authentication.
4. **Tenant DB Materialization**: The webhook updates the queue state and securely writes structured results directly into the specific tenant's isolated database.
5. **Jitter-Protected Polling**: The frontend JavaScript (`polling.js`) checks job progress using **Exponential Backoff with Random Jitter**, preventing server thundering-herd overload.

---

## 💾 Database Architecture & Data Isolation

```
db_agency_core (Global Core Database)
├── tenants               # UUID, domain, trade_module, db_name, api_keys
├── users                 # Centralized SSO identity provider credentials
└── ai_job_queue          # UUID, tenant_id, status, payload, response

db_tenant_[trade]_[id] (200+ Isolated Tenant Databases)
├── tutor_sessions        # Student sessions, AI match prep notes, timestamps
├── volunteers            # Availability blocks, contact details, status
└── [custom_trade_tables] # Unique schemas per industry (Plumbing, Auto, etc.)
```

---

## 🚀 Quick Start & Setup Wizard (`setup-wizard.exe`)

We provide a self-contained, standalone Windows executable (**`setup-wizard.exe`**) built in Go. It connects directly over SSH to your Ubuntu VPS and automates the entire 7-phase installation.

### Running the Setup Wizard
```powershell
# Interactive mode (Guides you with line-by-line tooltips):
.\setup-wizard.exe

# Dry-run mode (Preview all commands without making remote changes):
.\setup-wizard.exe -dry-run
```

### Compiling from Source (Requires Go 1.22+)
```powershell
# Download dependencies
go mod tidy

# Build standalone zero-dependency Windows executable
go build -o setup-wizard.exe ./cmd/wizard
```

### Interactive Post-Setup Menu
Once installation completes, the wizard presents an action suite:
* **`[1] 🌐 Open CloudPanel in your Browser`**: Automatically launches `https://<VPS_IP>:8443`.
* **`[2] 🧪 Run End-to-End System Tests`**: Verifies Nginx, PHP-FPM, MariaDB tables, and background AI queue loop over live SSH.
* **`[3] 📋 View DNS Configuration & SSL Guide`**: Pre-formatted A-Record copy-paste tables.
* **`[4] 📄 Export Setup Report`**: Writes `setup-summary.txt` with credentials and paths.
* **`[5] 🌐 Visit www.VEFA.club`**: Direct link to the official VEFA community.
* **`[6] 🧹 Clean Up & Exit`**: Removes `/tmp/*.sh` installer scripts from the VPS.

---

## 📁 Repository Structure

```text
├── cmd/
│   └── wizard/
│       └── main.go               # Interactive CLI wizard, banners, recovery & post-setup actions
├── internal/
│   ├── config/                   # Configuration model & persistence (agency-config.json)
│   ├── provisioner/              # 7-phase remote VPS provisioning engine
│   └── sshclient/                # Secure SSH client with streaming & diagnostic recovery
├── templates/                    # Embedded application templates (compiled into binary)
│   ├── nginx/
│   │   └── agency-master.conf    # Nginx Catch-All vHost template with FastCGI params
│   ├── php/
│   │   ├── app/
│   │   │   ├── Core/routes.php   # MCP Webhook relay receiver & IdP portal routes
│   │   │   └── Trades/           # Modular industry logic (Tutor, Plumber, etc.)
│   │   ├── public/index.php      # Slim PHP 4 front controller with dynamic DB injection
│   │   ├── composer.json         # Slim 4, PHP-DI, PSR-7 dependencies
│   │   └── worker.php            # CLI background daemon for AI job queue dispatch
│   ├── public/
│   │   └── polling.js            # Frontend exponential backoff polling utility
│   ├── sql/
│   │   ├── schema.sql            # MariaDB core schema (tenants, users, ai_job_queue)
│   │   └── seed_tenant.sql       # Demo tenant schema and initial seed data
│   └── systemd/
│       └── agency-worker.service # Systemd service unit for continuous daemon execution
├── agency-config.example.json    # Sanitized configuration template
├── go.mod                        # Go module definition
├── go.sum                        # Go module checksums
├── .gitignore                    # Strict Git ignore (protects runtime secrets & keys)
└── README.md                     # Master documentation
```

---

## 🔒 Security & GitHub Best Practices

* **Zero-Secret Guarantee**: Source code and templates contain zero hardcoded credentials or IP addresses.
* **Local Runtime Config**: When you run `setup-wizard.exe`, runtime credentials are saved locally to `agency-config.json`. This file is **strictly excluded** by [`.gitignore`](.gitignore) and will never be pushed to GitHub.
* **Sanitized Template**: Use [`agency-config.example.json`](agency-config.example.json) as a reference when sharing with teammates.

---

## 🌐 Community & Legal Notice

* **Official Portal**: [https://www.VEFA.club](https://www.VEFA.club)
* **Trademark**: **VEFA™: Virtual Environment Fellowship of America** *(Registered / Trademarked as of 2026 - All Rights Reserved)*.
* **Support**: For architecture guides, trade modules, and agentic orchestration updates, visit the official VEFA portal.
