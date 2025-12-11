# 🟣 Chat Service — Real-Time WebSocket Messaging (Go + MongoDB)

![Go Version](https://img.shields.io/badge/go-1.23-blue)

A high-performance real-time chat microservice built with **Go**, **WebSocket**, and **MongoDB**, designed for mentor–student communication inside an educational platform.  
It integrates with a Django backend for authentication and lesson validation.

---

## 🚀 Features

### 🔐 Authentication & Access Control
- JWT authentication via Django API  
- Role-aware access: **mentor** or **student**  
- Each chat is strictly bound to a specific **lesson**  
- Mentors can chat with multiple students; students only access their own chat  

### 💬 Real-Time WebSocket Communication
- Instant message delivery  
- Typing indicator (`typing`)  
- Read receipts (`viewed`)  
- Each chat uses its own room, identified by `conversation_id`  

### 🗂 Conversation Management
- Unique conversation per **(lesson_id + mentor_id + student_id)**  
- Duplicate creation prevented via MongoDB unique index  

### 📦 REST API Support
- List conversations  
- Get conversation details  
- Get messages  
- Check presence (online/offline)  

### 🏗 Tech Stack
- **Go (net/http)**
- **Gorilla WebSocket**
- **MongoDB**
- **Django Integration**
- Clean modular architecture  

---

## 📁 Project Structure

```bash

chat-service/
│── cmd/server/main.go
│── go.mod
│── pkg/
│ ├── auth/ # Django API integration (users, lessons)
│ ├── db/ # MongoDB repositories
│ ├── http/ # REST API handlers
│ ├── ws/ # WebSocket hub, client, events

```


---

## ⚙️ Environment Variables

| Variable           | Description |
|-------------------|-------------|
| `MONGO_URI`       | MongoDB connection string |
| `MONGO_DB`        | Database name |
| `PORT`            | HTTP server port |
| `DJANGO_BASE_URL` | Django backend base URL |

---

## ▶️ Running Locally

```bash
git clone https://github.com/YOUR_USERNAME/chat-service
cd chat-service

go mod tidy
go run main.go

```


## 🔌 WebSocket Connection URL

```bash
ws://<host>/ws?token=<JWT>&user_id=<UID>&lesson_id=<LID>&student_id=<SID>
```

### Notes:

 - student_id is required only for mentors
 - Students automatically become the student_id of the conversation



## 📡 WebSocket Events

### 📤 Inbound (Client → Server)

```rust

| Type      | Description            |
| --------- | ---------------------- |
| `message` | Send a chat message    |
| `typing`  | Typing indicator       |
| `viewed`  | Mark message as viewed |

```


### 📥 Outbound (Server → Clients)

1️⃣ New Message

```json
{
  "type": "message",
  "conversation_id": "abc123",
  "sender_id": 10,
  "text": "Hello!",
  "created_at": "2025-01-01T10:00:00Z"
}

```


2️⃣ Typing Indicator

```json
{
  "type": "typing",
  "conversation_id": "abc123",
  "sender_id": 10,
  "is_typing": true
}
```


3️⃣ Read Receipt

```json
{
  "type": "viewed",
  "message_id": "6512391abc...",
  "viewer_id": 15
}
```


## 📚 REST API Endpoints

- GET /conversations

```bash
/conversations?mentor_id=10
/conversations?student_id=25
```


- GET /conversations/{id}

```bash
/conversations/663af8129dde...
```

- GET /messages

```bash
/messages?conversation_id=663af8129dde...
```


- GET /presence

```bash
/presence?user_id=15
```


## 🧠 Architecture Overview

```rust
 ┌───────────────┐       ┌─────────────┐       ┌──────────────┐
 │   Web/Mobile   │ <---> │ Chat Service │ <---> │   MongoDB     │
 │     Client     │       │ (WebSocket) │       │ (Messages)   │
 └───────────────┘       └─────────────┘       └──────────────┘
            │
            ▼
     ┌────────────────┐
     │ Django Backend │  (Auth + Lesson validation)
     └────────────────┘
```



## 🧩 Key Advantages

- ✔ Highly scalable WebSocket hub
- ✔ Clean separation of concerns
- ✔ Persistent chat history
- ✔ Typing + read receipts
- ✔ Strong mentor–student permissions model
- ✔ Production-ready architecture


