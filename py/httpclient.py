import requests,_thread

def run(name):
    r = requests.get("http://localhost:9999/api?key=Tom")
    print(name+r.text[:200])



    
if __name__=='__main__':

    for i in range(10):
        _thread.start_new_thread(run, ("t"+str(i)+" ",))
    while 1:
        pass