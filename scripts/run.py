import os
import sys
import time
import json
import secrets
import argparse

import requests
import pandas as pd
# import dask.dataframe as dd
# import dask.array as da
import scipy.stats as stats
import numpy as np

from typing import List
from tqdm import tqdm
from concurrent.futures import ThreadPoolExecutor


OPERA_MAP = ['+', '-', '*', '/', "+'", "/'", '^']
OPERA_DICT = {
    'add': 0,
    'sub': 1,
    'mul': 2,
    'div': 3,
    'cadd': 4,
    'cdiv': 5,
    'exp': 6
}


def get_sample_size(confidence_level, error, total_sample_size, P=0.5):
    alpha = 1 - confidence_level
    z_alpha_2 = stats.norm.ppf(1 - alpha / 2)
    P_one_minus_P = P * (1 - P)

    if total_sample_size >= 1_0000_0000:
        n_infinite = (z_alpha_2 ** 2 * P_one_minus_P) / (error ** 2)
        return int(n_infinite)
    else:
        n_finite = (z_alpha_2 ** 2 * P_one_minus_P * total_sample_size) / ((total_sample_size - 1) * error ** 2 + z_alpha_2 ** 2 * P_one_minus_P)
        return int(n_finite)


def check_equal_row_count(dfs: List[pd.DataFrame]):
    row_counts = [len(df) for df in dfs]
    if len(set(row_counts)) > 1:
        print("Error: CSV files have different number of rows.")
        sys.exit(1)
    return row_counts[0]


def save_part_csv(part_df: pd.DataFrame, part_file_name):
    part_df.to_csv(part_file_name, index=True, index_label='number')


def split_csv(df: pd.DataFrame, file: str, num_parts: int, output_dir: str):
    rows_per_part = len(df) // num_parts
    remainder = len(df) % num_parts
    base_name = os.path.basename(file)
    name, ext = os.path.splitext(base_name)

    os.makedirs(output_dir, exist_ok=True)
    part_filenames = []
    # print(f'Splitting files: {file}')

    part_ranges = [
        (i * rows_per_part + min(i, remainder), (i + 1) * rows_per_part + min(i + 1, remainder))
        for i in range(num_parts)
    ]

    with ThreadPoolExecutor() as executor:
        futures = []
        for i, (start_idx, end_idx) in enumerate(part_ranges):
            part_df = df.iloc[start_idx:end_idx] 
            part_file_name = os.path.join(output_dir, f"{name}-{i + 1}{ext}")
            part_filenames.append(part_file_name)

            futures.append(executor.submit(save_part_csv, part_df, part_file_name))
        
        for future in tqdm(futures, desc=f"Splitting {file}", ncols=100, ascii=' #'):
            future.result() 
    
    return part_filenames


def post_file(url, file_path, party, file_id):
    files = { 'file': open(file_path, 'rb') }
    response = requests.post(url, data={'id': str(file_id), 'party': party}, files=files)
    return response.text


def get_request(url, params):
    response = requests.get(url, params=params)
    return response


def check_exception(response: str):
    resp_json = json.loads(response)
    if 'error' in resp_json:
        print(f"Error: {resp_json['error']}")
        sys.exit(1)


def process_files(part_files, file_id, result_dir, operate=2, base_url="http://localhost:9000", workers=8, scale=1):
    check_exception(post_file(base_url + "/update", part_files['A'], 'Alice', file_id))
    check_exception(post_file(base_url + "/update", part_files['B'], 'Bob', file_id))
    check_exception(post_file(base_url + "/update", part_files['R'], 'Result', file_id))

    response = get_request(
        base_url + "/verify", 
        params={
            'id': str(file_id), 
            'operate': str(operate), 
            'workers': workers,
            'scale': scale
        }
    )
    verify_response = json.loads(response.text)
    checked_errors = verify_response['checked_errors']

    share_info = verify_response['share_info']
    if share_info['error_alice'] or share_info['error_bob']:
        print('Something went wrong:', share_info['error_alice'], share_info['error_bob'])
        sys.exit(-1)

    verify_info = verify_response['verify_info']
    if verify_info['error_alice'] or verify_info['error_bob']:
        print('Something went wrong:', verify_info['error_alice'], verify_info['error_bob'])
        sys.exit(-1)

    real_comm = max(share_info['output_alice']['comm_cost'], share_info['output_bob']['comm_cost'])
    real_comm += max(verify_info['output_alice']['comm_cost'], verify_info['output_bob']['comm_cost'])

    real_time = max(share_info['output_alice']['total_time'], share_info['output_bob']['total_time'])
    real_time += max(verify_info['output_alice']['total_time'], verify_info['output_bob']['total_time'])

    os.makedirs(result_dir, exist_ok=True)
    response = get_request(base_url + "/result", params={'id': str(file_id)})
    result_file = os.path.join(result_dir, f"result_{file_id}.csv")
    with open(result_file, 'wb') as f:
        f.write(response.content)

    # response = get_request(base_url + "/delete", params={'id': str(file_id)})
    # print(f"Delete Response: {response.text}")\

    return result_file, checked_errors, real_comm, real_time


def combine_results(result_files, sample_indexes, combined_filename):
    combined_df = pd.concat(
        [pd.read_csv(f, dtype={'number': int, 'data': str}) for f in result_files])
    combined_df['number'] = sample_indexes
    combined_df = combined_df.sort_values(by='data', key=lambda x: x.map(lambda y: isinstance(y, float)), ascending=False)
    
    combined_df.to_csv(combined_filename, index=False)
    print(f"combined results saved as {combined_filename}")


def main():
    parser = argparse.ArgumentParser(description="Verifier Controller All-in-One")
    parser.add_argument('-a', '--file-a', type=str, required=True, help="Alice's file path")
    parser.add_argument('-b', '--file-b', type=str, required=True, help="Bob's file path")
    parser.add_argument('-r', '--file-r', type=str, required=True, help="Result's file path")
    parser.add_argument('-o', '--operator', type=str, required=True, default="mul", help=f"Operator: {list(OPERA_DICT.keys())}")
    parser.add_argument('-n', '--split-n', type=int, default=0, help="Split number")
    parser.add_argument('-w', '--workers', type=int, default=8, help="Verify workers")
    parser.add_argument('-s', '--scale', type=int, default=1, help="Precision control")
    parser.add_argument('-d', '--dir-out', type=str, default='./temp/', help="Output dir")
    parser.add_argument('-f', '--result-file-dir', type=str, default='./results/', help="Results file dir")
    parser.add_argument('-c', '--combined-file', type=str, default='combinedResult.csv', help="Combined result file name")
    parser.add_argument('-u', '--uri', type=str, default='http://localhost:9000', help="Verifier URI")
    parser.add_argument('--confidence-level', type=float, default=0.9999, help="Confidence level setting")
    parser.add_argument('--error-rate', type=float, default=0.001, help="Error rate setting")
    parser.add_argument('--all', action='store_true', default=False, help="Completely check")
    args = parser.parse_args()

    files = {
        'A': args.file_a,
        'B': args.file_b,
        'R': args.file_r
    }

    print('Reading...')
    origin_dfs = {
        'A': pd.read_hdf(files['A'], key='data'),
        'B': pd.read_hdf(files['B'], key='data'),
        'R': pd.read_hdf(files['R'], key='data')
    }

    row_count = check_equal_row_count(origin_dfs.values())
    if args.split_n < 0 or args.split_n > row_count:
        print(f"Error: split N must be between 0 and {row_count}.")
        sys.exit(1)

    sample_size = get_sample_size(args.confidence_level, args.error_rate, row_count)

    if row_count <= 100_0000 or args.all:
        sample_size = row_count
        sample_indexes = np.arange(1, row_count + 1).tolist()
    else:
        np.random.seed(secrets.randbelow(2 ** 32 - 2))
        sample_indexes = np.random.choice(np.arange(1, row_count + 1), size=sample_size, replace=False).tolist()

    if args.split_n == 0:   # auto detect
        args.split_n = sample_size // 100_0000
        args.split_n += 1 if sample_size % 100_0000 != 0 else 0
    
    print('Processing...')
    samples_a: pd.DataFrame = origin_dfs['A'].loc[sample_indexes]
    samples_b: pd.DataFrame = origin_dfs['B'].loc[sample_indexes]
    samples_r: pd.DataFrame = origin_dfs['R'].loc[sample_indexes]

    samples_a.index = range(1, sample_size + 1)
    samples_b.index = range(1, sample_size + 1)
    samples_r.index = range(1, sample_size + 1)

    # print(samples_a.head(10))
    # print(samples_b.head(10))
    # print(samples_r.head(10))

    dfs = {
        'A': samples_a,
        'B': samples_b,
        'R': samples_r,
    }

    split_files = {'A': [], 'B': [], 'R': []}

    for label, file in files.items():
        part_files = split_csv(dfs[label], file, args.split_n, args.dir_out)
        split_files[label] = part_files

    result_file_names = []
    difference, comm_cost, time_cost = 0, 0, 0.

    for x in tqdm(range(args.split_n), desc='Verifying', ncols=100, ascii=' #'):
        part_files = {
            'A': split_files['A'][x],
            'B': split_files['B'][x],
            'R': split_files['R'][x]
        }

        f_name, diff, c_cost, t_cost = process_files(
            part_files, 
            file_id=x+1, 
            result_dir=args.result_file_dir, 
            operate=OPERA_DICT[args.operator.lower()], 
            workers=args.workers, 
            scale=args.scale
        )

        result_file_names.append(f_name)
        difference += diff
        comm_cost += c_cost
        time_cost += t_cost
        time.sleep(0.1)

    total_mistake_rate = float(difference) / sample_size
    if row_count <= 100_0000 or args.all:
        args.error_rate = 0.
        args.confidence_level = 1.
    
    print('Summary:')
    print(f'\tdata length - {row_count}')
    print(f'\toperate between - {args.operator}'
          f'({OPERA_MAP[OPERA_DICT[args.operator]]})')
    print(f'\tmistakes in the calculation result - '
          f'{max(round(row_count * total_mistake_rate) - round(row_count * args.error_rate), 0)} ~ '
          f'{min(round(row_count * total_mistake_rate) + round(row_count * args.error_rate), row_count)}')
    print(f'\tmistake rate of the calculation result - '
          f'{round(total_mistake_rate * 100, 4)}% ± {round(args.error_rate * 100, 2)}%')
    print(f'\tconfidence level of checking - {args.confidence_level * 100}%')
    print(f'\terror rate of checking - {args.error_rate * 100}%')
    print(f'\tcalculation comm cost - {comm_cost} bits')
    print(f'\tcalculation time cost - {time_cost} ms')
    
    print('Saving...', end=' ')
    combine_results(result_file_names, sample_indexes, args.combined_file)


if __name__ == "__main__":
    import time

    start = time.time()
    main()

    print(f"Total time cost: {time.time() - start}")
