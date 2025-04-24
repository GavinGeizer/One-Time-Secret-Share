# One-Time Secret API

A simple Go API that lets you store a secret and retrieve it exactly once. After it's viewed, it's deleted.

## Features
- POST `/secret` – Store a secret, get a UUID
- GET `/secret/:id` – Retrieve the secret once, then delete it
- Secrets auto-expire after 10 minutes
- In-memory storage, no database

## Run It
```bash
go run main.go
```


## Example
### POST /secret

```json
{ "secret": "my message" }
```
### Response

```json
{ "id": "your-uuid-here" }
```
### GET /secret/:id

```json
{ "secret": "my message" }
```
#### After that, the secret is gone.
