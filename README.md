# Chat with AI – Scalable Backend System

## Project Overview
This project delivers a high-performance, production-ready backend for an AI-powered chat application. It is designed for scalability, real-time communication, and robust conversation management, with a strong focus on reliability and developer experience.

## Features
- **High-Speed Redis Caching**: Utilizes Redis to optimize database communication and reduce latency.
- **Graceful Shutdown**: Ensures no messages are lost during shutdown.
- **User Authentication with OAuth 2.0**: Secure user authentication using OAuth 2.0.
- **Message Transport with WebSocket**: Real-time message transport using WebSocket.
- **Conversation CRUD**: Create, Read, Update, and Delete operations for managing conversations.

## Installation
1. Clone the repository:
   ```bash
   git clone git@github.com:Ray-red-byte/chat-ai-backend.git
   ```
2. Navigate to the project directory:
   ```bash
   cd chat-ai-backend
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Start the server:
    - With other hosting Redis and Mongo DB 
   ```bash
   go run cmd/server/main.go
   ```

   - Run with full system (App + Mongo + Redis)
   ```bash
   docker-compose up --build

## Usage
- Access the API at `http://localhost:8080`.
- Use the Swagger documentation for API details.

## API Documentation
The API is documented using Swagger. You can view the documentation by navigating to `/swagger` endpoint on the running server.

## Contributing
Contributions are welcome! Please fork the repository and submit a pull request.

## License
This project is licensed under the MIT License.
