# FacturMe

A modern time tracking and invoicing platform for freelancers and small businesses.

https://github.com/user-attachments/assets/5186f3d5-1489-46f9-b0ca-275f040b79aa

## Features

- **Time Tracking** - Track billable hours with heatmaps and analytics
- **Client Management** - Organize work and set custom rates per client
- **Invoice Generation** - Create professional PDF invoices from time entries
- **Multi-Currency** - 12+ currencies with daily exchange rate updates
- **Dark UI** - Modern interface with interactive tour and demo data

## Tech Stack

- **Backend**: Go, Echo v4, PostgreSQL, JWT Auth, Swagger
- **Frontend**: SvelteKit, Tailwind CSS, Flowbite
- **DevOps**: Docker, Docker Compose

## Quick Start

### Prerequisites

- Docker and Docker Compose

### Installation

```bash
# Clone repository
git clone https://github.com/raihanstark/facturme.git
cd facturme

# Setup environment
cp .env.example .env
# Edit .env and change JWT_SECRET!

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f
```

**Access the application:**
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- API Docs: http://localhost:8080/swagger

**Useful commands:**
```bash
docker-compose logs -f        # View logs
docker-compose restart        # Restart services
docker-compose down           # Stop services
docker-compose up -d --build  # Rebuild and restart
```

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `DATABASE_URL` | PostgreSQL connection string | Yes |
| `JWT_SECRET` | Secret for JWT signing (change in production!) | Yes |
| `SMTP_HOST` | SMTP server for emails | No |
| `SMTP_USERNAME` | SMTP username | No |
| `SMTP_PASSWORD` | SMTP password | No |

## License

GPL-3.0 License - see LICENSE file for details.

---

Built with ❤️ for freelancers and small businesses
