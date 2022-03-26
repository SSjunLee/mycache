import requests

for i in range(5):
    r = requests.get("http://localhost:9999/api?key=Tom")
    print(r.text[:200])
    