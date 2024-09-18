import os
import sys
import argparse
import pandas as pd
import requests
import time

from typing import List
from tqdm import tqdm


def wait(n):
    time.sleep(n)


def check_equal_row_count(dfs: List[pd.DataFrame]):
    row_counts = [len(df) for df in dfs]
    if len(set(row_counts)) > 1:
        print("Error: CSV files have different number of rows.")
        sys.exit(1)
    return row_counts[0]


def split_csv(df: pd.DataFrame, file: str, num_parts: int, output_dir: str):
    rows_per_part = len(df) // num_parts
    remainder = len(df) % num_parts
    base_name = os.path.basename(file)
    name, ext = os.path.splitext(base_name)

    part_filenames = []
    print(f'Splitting files: {file}')
    for i in tqdm(range(num_parts)):
        start_idx = i * rows_per_part + min(i, remainder)
        end_idx = (i + 1) * rows_per_part + min(i + 1, remainder)
        part_df = df[start_idx:end_idx]
        
        part_file_name = os.path.join(output_dir, f"{name}-{i + 1}{ext}")
        part_df.to_csv(part_file_name, index=False)
        part_filenames.append(part_file_name)
        # print(f"Saved: {part_file_name}")
    
    return part_filenames


def post_file(url, file_path, party, file_id):
    files = { 'file': open(file_path, 'rb') }
    response = requests.post(url, data={'id': str(file_id), 'party': party}, files=files)
    print(f"POST {party}: {response.text}")


def get_request(url, params):
    response = requests.get(url, params=params)
    return response


def process_files(part_files, file_id, result_dir, operate=2, base_url="http://localhost:9000", workers=8, scale=1):
    post_file(base_url + "/update", part_files['A'], 'Alice', file_id)
    post_file(base_url + "/update", part_files['B'], 'Bob', file_id)
    post_file(base_url + "/update", part_files['R'], 'Result', file_id)
    wait(1)

    response = get_request(
        base_url + "/verify", 
        params={
            'id': str(file_id), 
            'operate': str(operate), 
            'workers': workers,
            'scale': scale
        }
    )
    print(f"Verify Response: {response.text}")
    wait(1)

    response = get_request(base_url + "/result", params={'id': str(file_id)})
    result_file = f"result_{file_id}.csv"
    with open(result_dir + result_file, 'wb') as f:
        f.write(response.content)
    print(f"Result saved as {result_file}")

    # response = get_request(base_url + "/delete", params={'id': str(file_id)})
    # print(f"Delete Response: {response.text}")


def main():
    parser = argparse.ArgumentParser(description="Verifier Controller All-in-One")
    parser.add_argument('-a', '--file-a', type=str, required=True, help="Alice's file path")
    parser.add_argument('-b', '--file-b', type=str, required=True, help="Bob's file path")
    parser.add_argument('-r', '--file-r', type=str, required=True, help="Result's file path")
    parser.add_argument('-n', '--split-n', type=int, required=True, help="Split number")
    parser.add_argument('-o', '--operator', type=int, required=True, default=2, help="Operator: 0~6 -> +,-,*,/,+',/',^")
    parser.add_argument('-w', '--workers', type=int, default=8, help="Verify workers")
    parser.add_argument('-s', '--scale', type=int, default=1, help="Precision control")
    parser.add_argument('-d', '--dir-out', type=str, default='./temp/', help="Output dir")
    args = parser.parse_args()

    files = {
        'A': args.file_a,
        'B': args.file_b,
        'R': args.file_r
    }

    dfs = {
        'A': pd.read_csv(files['A']),
        'B': pd.read_csv(files['B']),
        'R': pd.read_csv(files['R'])
    }

    row_count = check_equal_row_count(dfs.values())
    if args.split_n <= 0 or args.split_n > row_count:
        print(f"Error: split N must be between 1 and {row_count}.")
        sys.exit(1)

    split_files = {'A': [], 'B': [], 'R': []}

    for label, file in files.items():
        part_files = split_csv(dfs[label], file, args.split_n, args.dir_out)
        split_files[label] = part_files

    for x in range(args.split_n):
        part_files = {
            'A': split_files['A'][x],
            'B': split_files['B'][x],
            'R': split_files['R'][x]
        }
        process_files(part_files, file_id=x+1, result_dir='./results/', 
                      operate=args.operator, workers=args.workers, scale=args.scale)
        wait(1)



if __name__ == "__main__":
    main()

