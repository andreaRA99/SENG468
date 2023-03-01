## Transaction-Server API

## Getting account balance for a user when logging in (creates user if not exists)  
`GET /users/:id`  
**Response**
- `200 OK` on succes
```json
{
    "id": "mike123",
    "balance": 100
}
```

## Adding money to an account
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

## Request for Stock Quote  
`GET /users/:id/quote/:stock`  
**Response**
```json
{
    "stock_symbol": "APPL",
    "price": 250.01
}
```

## Buy Quote  
`POST /users/:id/buy/:stock`  
**Arguments**
- `"id":string` User ID 
- `"stock":string` Stock Symbol
- `"buy":float64` Dollar amount to buy  

**Response**
```json
{
    "stock_symbol": "APPL",
    "price": 250.01, // last quoted to specific user or in general?? probably need to check how old quote is...
    "num_stocks": 1, // number of  stocks buy amount is worth, DOES THIS NEED BE INT??
    "buy_id": 1 // some way to identify orders and time them out, int for simplicity but can be diff
}
```

## Confirm Buy   
`POST /users/:id/buy/:stock/:buyid`  
**Arguments**
- `"id":string` User ID 
- `"stock":string` Stock Symbol
- `"buy":float64` Dollar amount to buy
- `"buyid":int` Order identifier  

**Response**
```json
{
    "stock_symbol": "APPL",
    "price": 250.01, // last quoted to specific user or in general?? probably need to check how old quote is...
    "num_stocks": 1, // number of  stocks buy amount is worth, DOES THIS NEED BE INT??
    "balance": 20 // balance after transaction
}
```

## Cancel Buy  
`DELETE /users/:id/buy/cancel/:buyid`  
**Response**
-`404 Not Found`

## Sell Quote  
**Definitions**
`POST /users/:id/sell/:stock`  
**Arguments**
- `"id":string` User ID 
- `"stock":string` Stock Symbol
- `"sell":float64` Dollar amount to sell  

**Response**
```json
{
    "stock_symbol": "APPL",
    "price": 250.01, // current price of stock
    "num_stocks": 1 // number of  stocks buy amount is worth
}
```

## Confirm Sell  
`POST /users/:id/sell/:stock/:sellid`  
**Arguments**
- `"id":string` User ID 
- `"stock":string` Stock Symbol
- `"sell":float64` Dollar amount to buy
- `"sellid":int` Order identifier  

**Response**
```json
{
    "stock_symbol": "APPL",
    "price": 250.01, // last quoted to specific user or in general?? probably need to check how old quote is...
    "num_stocks": 1, // number of  stocks buy amount is worth, DOES THIS NEED BE INT??
    "balance": 20 // balance after transaction
}
```

## Cancel Sell  
`DELETE /users/:id/sell/cancel/:sellid`  
**Response**
-`404 Not Found`


