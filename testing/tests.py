import requests

print("Begin tests:")
print()
tests_passed = 0 
tests_failed = 0
print("1. All users:    ", end="")
try:
   n = requests.get("http://host.docker.internal:8080/users").content.decode()
   if "Bad request" in n or "404 page not found" in n:
      raise Exception     
   tests_passed = tests_passed + 1
   print("passed")
except Exception:
   tests_failed = tests_failed + 1
   print("failed")

print("2. Some user:    ", end="")
try:
   n = requests.get("http://host.docker.internal:8080/users/test_user").content.decode()
   if "Bad request" in n or "404 page not found" in n:
      raise Exception     
   if "test_user" not in n:
      raise Exception
   tests_passed = tests_passed + 1
   print("passed")
except Exception:
   tests_failed = tests_failed + 1
   print("failed")

print("3. Add value:    ", end="")
try:
   r = '{"ID": "test_user", "Amount": 990}'
   n = requests.put("http://host.docker.internal:8080/users/addBal", data=r).content.decode()
   if "Bad request" in n or "404 page not found" in n:
      raise Exception     
   tests_passed = tests_passed + 1
   print("passed")
except Exception:
   tests_failed = tests_failed + 1
   print("failed")

print("4. Get Quote:    ", end="")
try:
   n = requests.get("http://host.docker.internal:8080/users/test_user/quote/ccc").content.decode()
   if "Bad request" in n or "404 page not found" in n:
      raise Exception  
   tests_passed = tests_passed + 1
 
   print("passed")
except Exception:
   tests_failed = tests_failed + 1
   print("failed")

print("5. Bad request:    ", end="")
try:
   n = requests.get("http://host.docker.internal:8080/no_url").content.decode()
   if "Bad request" in n or "404 page not found" not in n:
      raise Exception
   tests_passed = tests_passed + 1
   print("passed")
except Exception:
   tests_failed = tests_failed + 1
   print("failed")


print("6. Buy stock:    ", end="")
try:
   r = '{"ID": "test_user", "Stock":"ccc", "Amount": 9}'
   n = requests.post("http://host.docker.internal:8080/users/buy", data=r).content.decode()
   if "Bad request" in n or "404 page not found" in n:      
      raise Exception
   tests_passed = tests_passed + 1
   print("passed")
except Exception:
   tests_failed = tests_failed + 1
   print("failed")   

print("7. Commit buy:    ", end="")
try:
   r = '{"ID": "test_user", "Stock":"ccc"}'
   n = requests.post("http://host.docker.internal:8080/users/buy/commit", data=r).content.decode()
   if "Bad request" in n or "404 page not found" in n:
      raise Exception
   tests_passed = tests_passed + 1
   print("passed")
except Exception:
   tests_failed = tests_failed + 1
   print("failed")   

print("8. Cancel buy:    ", end="")
try:
   r = '{"ID": "test_user", "Stock":"ccc", "Amount": 9}'
   n = requests.post("http://host.docker.internal:8080/users/buy", data=r).content.decode()
   if "Bad request" in n or "404 page not found" in n:
      raise Exception
   n = requests.post("http://host.docker.internal:8080/users/test_user/buy/cancel", data=r).content.decode()
   if "Bad request" in n or "404 page not found" in n:
      raise Exception
   tests_passed = tests_passed + 1
   print("passed")
except Exception:
   tests_failed = tests_failed + 1
   print("failed")   

print("9. Sell stock:    ", end="")
try:
   r = '{"ID": "test_user", "Stock":"ccc"}'
   n = requests.post("http://host.docker.internal:8080/users/sell", data=r).content.decode()
   if "Bad request" in n or "404 page not found" in n:
      raise Exception
   tests_passed = tests_passed + 1
   print("passed")
except Exception:
   tests_failed = tests_failed + 1
   print("failed")   

print("10. Commit sell:    ", end="")
try:
   r = '{"ID": "test_user", "Stock":"ccc"}'
   n = requests.post("http://host.docker.internal:8080/users/test_user/sell/commit", data=r).content.decode()
   if "Bad request" in n or "404 page not found" in n:
      raise Exception
   tests_passed = tests_passed + 1
   print("passed")
except Exception:
   tests_failed = tests_failed + 1
   print("failed")  

print("11. Cancel sell:    ", end="")
try:
   r = '{"ID": "test_user", "Stock":"ccc"}'
   n = requests.post("http://host.docker.internal:8080/users/test_user/sell/ccc/amount/9", data=r).content.decode()
   if "Bad request" in n or "404 page not found" in n:
      raise Exception
   n = requests.post("http://host.docker.internal:8080/users/test_user/sell/cancel", data=r).content.decode()
   if "Bad request" in n or "404 page not found" in n:
      raise Exception
   tests_passed = tests_passed + 1
   print("passed")
except Exception:
   tests_failed = tests_failed + 1
   print("failed")   

print("12. Set buy amount:    ", end="")
try:
   r = '{"ID": "test_user", "Stock":"ccc"}'
   n = requests.post("http://host.docker.internal:8080/users/test_user/set_buy/ccc/amount/9", data=r).content.decode()
   if "Bad request" in n or "404 page not found" in n:
      raise Exception   
   tests_passed = tests_passed + 1
   print("passed")
except Exception:
   tests_failed = tests_failed + 1
   print("failed")   

print("13. cancel set buy amount:    ", end="")
try:
   r = '{"ID": "test_user", "Stock":"ccc"}'
   n = requests.post("http://host.docker.internal:8080/users/test_user/set_buy/cancel/ccc/", data=r).content.decode()
   if "Bad request" in n or "404 page not found" in n:
      raise Exception   
   tests_passed = tests_passed + 1
   print("passed")
except Exception:
   tests_failed = tests_failed + 1
   print("failed")   

print("14. set buy trigger:    ", end="")
try:
   r = '{"ID": "test_user", "Stock":"ccc"}'
   n = requests.post("http://host.docker.internal:8080/users/test_user/set_buy/trigger/ccc/amount/9", data=r).content.decode()
   if "Bad request" in n or "404 page not found" in n:
      raise Exception   
   tests_passed = tests_passed + 1
   print("passed")
except Exception:
   tests_failed = tests_failed + 1
   print("failed")   

print("15. set sell trigger:    ", end="")
try:
   r = '{"ID": "test_user", "Stock":"ccc"}'
   n = requests.post("http://host.docker.internal:8080/users/test_user/set_sell/trigger/ccc/amount/9", data=r).content.decode()
   if "Bad request" in n or "404 page not found" in n:
      raise Exception   
   tests_passed = tests_passed + 1
   print("passed")
except Exception:
   tests_failed = tests_failed + 1
   print("failed")  


print("16. Set sell amount:    ", end="")
try:
   r = '{"ID": "test_user", "Stock":"ccc"}'
   n = requests.post("http://host.docker.internal:8080/users/test_user/set_sell/ccc/amount/9", data=r).content.decode()
   if "Bad request" in n or "404 page not found" in n:
      raise Exception   
   tests_passed = tests_passed + 1
   print("passed")
except Exception:
   tests_failed = tests_failed + 1
   print("failed")   

print("17. cancel set sell amount:    ", end="")
try:
   r = '{"ID": "test_user", "Stock":"ccc"}'
   n = requests.post("http://host.docker.internal:8080/users/test_user/set_sell/cancel/ccc/", data=r).content.decode()
   if "Bad request" in n or "404 page not found" in n:
      raise Exception   
   tests_passed = tests_passed + 1
   print("passed")
except Exception:
   tests_failed = tests_failed + 1
   print("failed")       

print("18. dumplog:    ", end="")
try:
   r = '{"ID": "test_user", "Stock":"ccc"}'
   n = requests.post("http://host.docker.internal:8080/dumplog/test_file", data=r).content.decode()
   if "Bad request" in n or "404 page not found" in n:
      raise Exception   
   tests_passed = tests_passed + 1
   print("passed")
except Exception:
   tests_failed = tests_failed + 1
   print("failed")  

print("19. display summary:    ", end="")
try:
   r = '{"ID": "test_user", "Stock":"ccc"}'
   n = requests.post("http://host.docker.internal:8080/users/test_user/display_summary", data=r).content.decode()
   if "Bad request" in n or "404 page not found" in n:
      raise Exception   
   tests_passed = tests_passed + 1
   print("passed")
except Exception:
   tests_failed = tests_failed + 1
   print("failed")  

print("20. delete user:    ", end="")
try:
   r = '{"ID": "test_user"}'
   n = requests.post("http://host.docker.internal:8080/users/test_user/delete_user", data=r).content.decode()
   if "Bad request" in n or "404 page not found" in n:
      raise Exception   
   tests_passed = tests_passed + 1
   print("passed")
except Exception:
   tests_failed = tests_failed + 1
   print("failed")  

print("Tests passed: %d" % tests_passed)
print("Tests failed: %d" % tests_failed)