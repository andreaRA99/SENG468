## Usage

### Getting account balance for a user when logging in (creates user if not exists)
**Definition**
`GET /users/:id`
**Response**
- `200 OK` on succes
```json
{
    "id": "mike123",
    "balance": 100
}
```
### Adding money to an account
**Definitions**
`PUT /users/:id/addBal`
**Arguments**
- `"id":string` User ID 
- `"amount":float64` money to add to account
**Response**
```json
{
    "id": "mike123",
    "balance": 100
}
```

### Request for Stock Quote
**Definitions**
`GET /users/:id/quote/:stock`
**Response**
```json
{
    "stock_symbol": "APPL",
    "quote": 250.01
}
```


