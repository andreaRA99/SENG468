import requests

""" Limit order 
	Stock  string
	Price  float64
	Type 	 string 
	Amount float64
	User   string 
"""


""" Tserver order 
	ID     string  `json:"id"`
	Stock  string  `json: "stock"`
	Amount float64 `json:"amount"`
	Price  float64
	Qty    int
"""


r = '{"ID": "test_user", "Stock":"ccc", "Amount": 9, "Price":40, "Type":"buy"}'
n = requests.post("http://localhost:8081/new_limit", data=r).content.decode()
print(n)

#r = '{"ID": "test_user", "Stock":"ccc", "Amount": 9, "Price":100, "Type":"sell"}'
#n = requests.post("http://localhost:8081/new_limit", data=r).content.decode()
#print(n)

