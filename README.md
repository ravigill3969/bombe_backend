# 🚗 Bombe

Bombe is a full-stack ride-hailing platform inspired by services such as Uber. It allows riders to request rides, drivers to accept and manage trips, and both sides to receive real-time trip and location updates.

The project was built to explore building a production-style distributed application using Go, PostgreSQL, Redis, AWS SQS, Stripe, WebSockets, React, and Docker.

## ✨ Features

### Rider

* Create an account and authenticate
* Request a ride
* Select ride type
* View estimated fare
* Track driver location in real time
* View active trip
* Cancel a trip
* Receive trip status updates
* Complete payments through Stripe

### Driver

* Create an account and authenticate
* Go online/offline
* Receive available ride requests
* Accept rides
* View active trips
* Send real-time location updates
* Pick up and complete rides
* Cancel trips
* View earnings

### Platform

* Real-time communication using WebSockets
* Asynchronous service communication using AWS SQS
* Stripe payment processing
* Payment/refund handling
* Redis for real-time/location-related data
* PostgreSQL for persistent application data
* Docker-based deployment
* Nginx reverse proxy
* HTTPS with Cloudflare

## 🏗️ Architecture

Bombe uses a service-oriented backend architecture.

```text
                    ┌──────────────┐
                    │   React App  │
                    │ TypeScript   │
                    └──────┬───────┘
                           │ HTTPS / WebSocket
                           ▼
                    ┌──────────────┐
                    │    Nginx     │
                    │ Reverse Proxy│
                    └──────┬───────┘
                           │
                           ▼
                    ┌──────────────┐
                    │ Vital Gateway│
                    │    :8080     │
                    └──────┬───────┘
                           │
          ┌────────────────┼────────────────┐
          │                │                │
          ▼                ▼                ▼
    ┌──────────┐     ┌──────────┐     ┌──────────┐
    │   Auth   │     │   Trip   │     │ Payment  │
    │  :50051  │     │  :50053  │     │  :50052  │
    └──────────┘     └────┬─────┘     └────┬─────┘
                          │                │
                          └───────┬────────┘
                                  ▼
                           ┌────────────┐
                           │ AWS SQS    │
                           └─────┬──────┘
                                 │
                    ┌────────────┴────────────┐
                    ▼                         ▼
              ┌───────────┐             ┌──────────┐
              │ PostgreSQL│             │  Redis   │
              └───────────┘             └──────────┘
```

## 🛠️ Tech Stack

### Backend

* Go
* gRPC
* PostgreSQL
* Redis
* AWS SQS
* Stripe
* WebSockets

### Frontend

* React
* TypeScript
* Mapbox
* shadcn/ui

### Infrastructure

* Docker
* Docker Compose
* Nginx
* Cloudflare
* Ubuntu VPS

## 🔄 Trip Flow

A typical ride follows this flow:

```text
Rider requests ride
       ↓
Trip service creates trip
       ↓
Driver receives available ride
       ↓
Driver accepts ride
       ↓
Trip status → assigned
       ↓
Driver sends location updates
       ↓
Rider receives location through WebSocket
       ↓
Driver picks up rider
       ↓
Trip status → picked
       ↓
Driver completes trip
       ↓
Trip status → completed
       ↓
Payment is finalized
```

## 💳 Payment Flow

Bombe uses Stripe for payment processing.

```text
Rider requests ride
       ↓
Payment session created
       ↓
Stripe Checkout
       ↓
Stripe Webhook
       ↓
Payment service
       ↓
AWS SQS
       ↓
Trip service
       ↓
Trip confirmed
```

Cancellation and refund operations are handled through the payment service.

## 📡 Real-Time Communication

WebSockets are used for real-time events such as:

* Driver location updates
* Trip status changes
* Driver assignment
* Ride cancellation
* Rider/driver notifications

Drivers periodically send their current location while online and during active rides.

## 🗄️ Data Storage

### PostgreSQL

PostgreSQL stores persistent application data including:

* Users
* Trips
* Payments
* Driver/rider information
* Trip status
* Fare information
* Timestamps

### Redis

Redis is used for temporary and real-time data such as driver location and online state.

### AWS SQS

SQS provides asynchronous communication between services and helps decouple operations such as:

* Trip events
* Payment events
* Refund processing
* Service-to-service messaging

## 🔐 Security

* HTTPS
* Secure authentication cookies
* Environment-based secrets
* Database credentials stored outside source code
* Stripe webhook processing
* Nginx reverse proxy
* Firewall rules on the VPS

Secrets and API keys are intentionally excluded from the repository.

## 🚀 Running Locally

### Requirements

* Go
* Node.js
* PostgreSQL
* Redis
* Docker
* Docker Compose
* Stripe account for payment functionality

### Clone

```bash
git clone <repository-url>
cd bombe
```

### Environment Variables

Create the required environment files and configure:

```env
DATABASE_URL=
REDIS_URL=
STRIPE_SECRET_KEY=
STRIPE_WEBHOOK_SECRET=
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_REGION=
SQS_QUEUE_URL=
```

Never commit real credentials.

### Start Backend

```bash
docker compose up --build
```

### Start Frontend

```bash
cd frontend
npm install
npm run dev
```

## 🌐 Production Deployment

Bombe is deployed using Docker containers on an Ubuntu VPS.

The production stack consists of:

```text
Cloudflare
    ↓
Nginx
    ↓
Docker Network
    ↓
Vital Gateway
    ↓
Backend Services
```

Nginx handles HTTPS termination and routes traffic to the gateway.

## 📁 Project Structure

```text
backend/
├── auth/
├── payment/
├── trip/
├── vital-gateway/
├── proto/
└── docker-compose.yml

frontend/
├── src/
│   ├── components/
│   ├── contexts/
│   ├── hooks/
│   ├── pages/
│   └── services/
└── package.json

docs/
├── architecture.md
├── api.md
├── database.md
├── deployment.md
└── development.md
```

## 🎯 Project Goals

Bombe was built to gain practical experience with:

* Distributed backend services
* Go concurrency
* gRPC communication
* Event-driven architecture
* Asynchronous messaging
* Real-time WebSocket communication
* Payment processing
* Database design
* Docker-based deployment
* Production infrastructure

## 🔮 Future Improvements

Possible future improvements include:

* Driver/rider ratings
* Push notifications
* Ride history improvements
* Advanced driver matching
* Surge pricing
* Better observability
* Automated testing
* CI/CD pipelines
* Metrics and tracing

## 👨‍💻 Author

Built as a full-stack engineering project to explore production-style distributed systems and real-time applications.
