## Ethereum Engine API Client
A Go implementation of a client for the Ethereum Engine API, which facilitates communication between consensus and execution clients in post-merge Ethereum. This client provides a robust interface for interacting with the Engine API endpoints, including fork choice updates, payload management, and configuration exchange.

### Features

- 🔒 JWT authentication for secure API communication
- 🔄 Automatic retry mechanism with configurable policies
- ⚡ Support for Engine API methods (forkchoiceUpdated, newPayload)
- 🛠️ Configurable client options (timeout, retry policy)
- 📝 Type-safe request and response handling
- 🎯 Context-aware operations
