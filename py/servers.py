import os
from multiprocessing import Pool

func = lambda:os.system("go run main/main.go -port 8002")


def f(cmd):
    os.system(cmd)

if __name__=='__main__':
    os.chdir("../")
    pool = Pool(processes=3)
    args = ["go run main/main.go -port 8001",
    "go run main/main.go -port 8002",
    "go run main/main.go -port 8003"]
    for i in args:
        pool.apply_async(f,(i,))

    pool.close()
    pool.join()    
    
    

