import csv
import random

def generate_csv(filename, N):
    with open(filename, mode='w', newline='') as file:
        writer = csv.writer(file)
        writer.writerow(['number', 'data'])
        
        for i in range(1, N + 1):
            data = random.uniform(0, 1)
            writer.writerow([i, data])

N = 100_0000  
generate_csv('data1M.csv', N)
print(f"生成了包含{N}行数据的CSV文件 'data1M.csv'")
