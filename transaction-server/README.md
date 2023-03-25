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
`PUT /users/addBal`  
**Arguments**
- `"id":string` user id 
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
`POST /users/buy`  
**Arguments**
- `"id":string` user id
- `"stock":string` Stock Symbol
- `"amount":float64` Dollar amount to buy  

**Response**
```json
{
    "stock_symbol": "APPL",
    "price": 250.01, // last quoted to specific user or in general?? probably need to check how old quote is...
    "num_stocks": 1, // number of  stocks buy amount is worth, DOES THIS NEED BE INT??
    //"buy_id": 1 // some way to identify orders and time them out, int for simplicity but can be diff
}
```

## Commit Buy   
`POST /users/buy/commit`  
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
`DELETE /users/buy/cancel`  
**Arguments**
- `"id":string` User ID 
**Response**
<!-- -`404 Not Found` -->

## Sell Quote  
**Definitions**
`POST /users/sell`  
**Arguments**
- `"id":string` User ID 
- `"stock":string` Stock Symbol
- `"amount":float64` Dollar amount to sell  

**Response**
```json
{
    "stock_symbol": "APPL",
    "price": 250.01, // current price of stock
    "num_stocks": 1 // number of  stocks buy amount is worth
}
```

## Commit Sell  
`POST /users/sell`  
**Arguments**
- `"id":string` User ID 

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
`DELETE /users/:id/sell/cancel` 
**Arguments**
- `"id":string` User ID  
**Response**
<!-- -`404 Not Found` -->

## Set Buy Amount  
`POST /users/setbuy`  
**Arguments**
- `"id":string` User ID 
- `"stock":string` Stock Symbol
- `"amount":float64` Dollar amount to sell  

**Response**
```json
{
    "stock_symbol": "APPL",
    "price": 250.01, // current price of stock
    "num_stocks": 1 // number of  stocks buy amount is worth
}

## Cancel Set Buy Amount
`POST /users/setbuy/cancel`
**Arguments**
- `"id":string` User ID 
- `"stock":string` Stock Symbol

**Response**
```json
{
    "stock_symbol": "APPL",
    "price": 250.01, // current price of stock
    "num_stocks": 1 // number of  stocks buy amount is worth
}

## Set Buy Trigger  
`POST /users/setbuy/trigger`  
**Arguments**
- `"id":string` User ID 
- `"stock":string` Stock Symbol
- `"amount":float64` Dollar amount 
**Response**
```json
{
    "stock_symbol": "APPL",
    "price": 250.01, // current price of stock
    "num_stocks": 1 // number of  stocks buy amount is worth
}

## Set Sell Amount  
`POST /users/setsell`  
**Arguments**
- `"id":string` User ID 
- `"stock":string` Stock Symbol
- `"amount":float64` Dollar amount to sell  

**Response**
```json
{
    "stock_symbol": "APPL",
    "price": 250.01, // current price of stock
    "num_stocks": 1 // number of  stocks buy amount is worth
}

## Cancel Set Sell Amount
`POST /users/setsell/cancel`
**Arguments**
- `"id":string` User ID 
- `"stock":string` Stock Symbol

**Response**
```json
{
    "stock_symbol": "APPL",
    "price": 250.01, // current price of stock
    "num_stocks": 1 // number of  stocks buy amount is worth
}

## Set Sell Trigger  
`POST /users/setsell/trigger`  
**Arguments**
- `"id":string` User ID 
- `"stock":string` Stock Symbol
- `"amount":float64` Dollar amount 
**Response**
```json
{
`
}

## Dumplog  
`POST /dumplog`  
**Arguments**
- `"id":string` User ID 
- `"filename":string` File to write log
**Response**
```json
{

}

## Display Summary  
`POST /displaysummary`  
**Arguments**
- `"id":string` User ID 
**Response**
```json
{
 
}
