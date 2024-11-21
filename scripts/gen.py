import csv
import random
import argparse


def generate_csv(filename, N):
    with open(filename, mode='w', newline='') as file:
        writer = csv.writer(file)
        writer.writerow(['number', 'data'])
        
        for i in range(1, N + 1):
            data = random.uniform(0, 1)
            writer.writerow([i, data])


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Generate CSV with random data from 0 to 1.")
    parser.add_argument('-n', '--number', type=int, required=True, help="Number of rows to generate.")
    parser.add_argument('-f', '--filename', type=str, required=True, help="Output filename.")
    args = parser.parse_args()
    
    generate_csv(args.filename, args.number)
    print(f"Generated {args.number} rows of data in {args.filename}.")

