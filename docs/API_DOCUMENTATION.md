# API Documentation

This document describes how to generate and view the API documentation for DropsAndGrinds.

## Generating Swagger Documentation

The API documentation is automatically generated from Swagger annotations in the handler files. To regenerate the documentation:

```bash
# Install swag if not already installed
go install github.com/swaggo/swag/cmd/swag@latest

# Generate swagger docs
swag init -g cmd/server/main.go -o docs
```

## New API Endpoints Added

### Store Health Check Endpoints

#### All Stores Health
- **Endpoint**: `GET /health/stores`
- **Description**: Returns health status for all store APIs
- **Response**: Map of store names to health status objects
- **Tags**: system

#### Individual Store Health
- **Endpoint**: `GET /health/stores/{store}`
- **Description**: Returns health status for a specific store API
- **Parameters**: 
  - `store` (path): Store name (steam, epic, xbox, playstation, nintendo, greenmangaming, fanatical, humble, indian)
- **Response**: Store health status object
- **Tags**: system

### Deal Alerts Endpoints

#### Create/List Deal Alerts
- **Endpoint**: `POST /api/deal-alerts` - Create a new deal alert
- **Endpoint**: `GET /api/deal-alerts` - List authenticated user's deal alerts
- **Authentication**: Required
- **Tags**: deal-alerts

#### Update/Delete Deal Alert
- **Endpoint**: `PATCH /api/deal-alerts/{id}` - Update target price for deal alert
- **Endpoint**: `DELETE /api/deal-alerts/{id}` - Remove deal alert
- **Authentication**: Required
- **Tags**: deal-alerts

### Indian Payment Offers Endpoint

#### Get Indian Payment Offers
- **Endpoint**: `GET /api/indian-offers`
- **Description**: Fetch Indian payment offers with optional filtering
- **Query Parameters**:
  - `store_id` (optional): Filter by store ID
  - `provider` (optional): Filter by payment provider (phonepe, gpay, paytm, etc.)
- **Response**: List of active Indian payment offers
- **Tags**: indian-offers

## Viewing the Documentation

### Swagger UI
Once the server is running, access the Swagger UI at:
```
http://localhost:8080/swagger/index.html
```

### Swagger JSON
The raw Swagger JSON is available at:
```
http://localhost:8080/swagger/doc.json
```

## API Response Formats

### Store Health Status Object
```json
{
  "store": "steam",
  "status": "up",
  "latency": 45,
  "last_check": "2025-05-06T14:30:00Z",
  "error": ""
}
```

### Indian Payment Offer Object
```json
{
  "id": 1,
  "store_id": 1,
  "offer_type": "upi_discount",
  "provider": "phonepe",
  "description": "10% instant discount on Steam purchases using PhonePe UPI",
  "discount_percent": 10,
  "max_discount_amount": 200.00,
  "min_order_amount": 500.00,
  "valid_from": "2025-05-06T00:00:00Z",
  "valid_until": "2025-08-06T00:00:00Z",
  "is_active": true
}
```

## Authentication

Most endpoints require JWT authentication. To authenticate:

1. Register or login to get a JWT token
2. Include the token in the Authorization header:
   ```
   Authorization: Bearer <your-jwt-token>
   ```

## Rate Limiting

All API endpoints are rate-limited to 60 requests per minute per IP address. If you exceed this limit, you'll receive a `429 Too Many Requests` response.

## Error Responses

All endpoints return errors in the following format:

```json
{
  "error": "Error message describing what went wrong"
}
```

Common HTTP status codes:
- `200 OK`: Successful request
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Authentication required or invalid
- `404 Not Found`: Resource not found
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error
- `503 Service Unavailable`: Service temporarily unavailable
