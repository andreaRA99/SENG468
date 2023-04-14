import requests



r = '{"username": "test_user", "sym":"ccc"}'
n = requests.get("http://localhost:8080/users/test_user/quote/ccc").content.decode()
print(n)

#r = '{"ID": "test_user", "Stock":"ccc", "Amount": 9, "Price":100, "Type":"sell"}'
#n = requests.post("http://localhost:8081/new_limit", data=r).content.decode()
#print(n)

