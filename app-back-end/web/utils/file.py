import os
import json
import hashlib
import asyncio
import pandas as pd

from io import BytesIO
from typing import List
from concurrent.futures import ThreadPoolExecutor

from .http import check_exception, post_file, get_request


async def file_summary(file_cont: bytes):
    md5_hash = hashlib.md5()
    md5_hash.update(file_cont)
    file_md5 = md5_hash.hexdigest()

    df = pd.read_csv(BytesIO(file_cont))
    if "data" not in df.columns:
        raise ValueError("The column 'data' is missing in the file.")

    return {
        "md5": file_md5,
        "items": len(df),
        "mean": float(df["data"].mean()),
        "std": float(df["data"].std()),
        "max": float(df["data"].max()),
        "min": float(df["data"].min())
    }


async def process_files(
    part_files, 
    task_id,
    file_id, 
    operate, 
    workers, 
    scale, 
    result_dir, 
    base_url
):
    real_file_id = f'{task_id}_{file_id}'

    check_exception(post_file(base_url + "/update", part_files['A'], 'Alice', real_file_id))
    check_exception(post_file(base_url + "/update", part_files['B'], 'Bob', real_file_id))
    check_exception(post_file(base_url + "/update", part_files['R'], 'Result', real_file_id))

    response = get_request(
        base_url + "/verify", 
        params={
            'id': real_file_id, 
            'operate': str(operate), 
            'workers': workers,
            'scale': scale
        }
    )
    verify_response = json.loads(response.text)

    share_info = verify_response['share_info']
    verify_info = verify_response['verify_info']

    real_comm = max(share_info['output_alice']['comm_cost'], share_info['output_bob']['comm_cost'])
    real_comm += max(verify_info['output_alice']['comm_cost'], verify_info['output_bob']['comm_cost'])

    real_time = max(share_info['output_alice']['total_time'], share_info['output_bob']['total_time'])
    real_time += max(verify_info['output_alice']['total_time'], verify_info['output_bob']['total_time'])

    response = get_request(base_url + "/result", params={'id': real_file_id})
    result_file = os.path.join(result_dir, f"result_{file_id}.csv")
    with open(result_file, 'wb') as f:
        f.write(response.content)

    return result_file, verify_response['checked_errors'], real_comm, real_time


def combine_results(result_files: str, combined_filename: str):
    combined_df = pd.concat([pd.read_csv(f, dtype={'number': int, 'data': str}) for f in result_files])
    combined_df.to_csv(combined_filename, index=False)


def save_part(pair):
    part_df, part_file_name = pair
    part_df.to_csv(part_file_name, index=False)
    return part_file_name


def save_part_csv(part_df: pd.DataFrame, part_file_name):
    part_df.to_csv(part_file_name, index=True, index_label='number')


async def split_csv(df: pd.DataFrame, file_name: str, num_parts: int, output_dir: str):
    rows_per_part = len(df) // num_parts
    remainder = len(df) % num_parts
    base_name = os.path.splitext(file_name)[0]

    part_dfs: List[pd.DataFrame] = []
    part_filenames: List[str] = []
    for i in range(num_parts):
        start_idx = i * rows_per_part + min(i, remainder)
        end_idx = (i + 1) * rows_per_part + min(i + 1, remainder)
        part_df = df[start_idx:end_idx]
        part_file_name = os.path.join(output_dir, f"{base_name}-{i + 1}.csv")
        # part_df.to_csv(part_file_name, index=False)
        part_dfs.append(part_df)
        part_filenames.append(part_file_name)
        await asyncio.sleep(0.01)

    for part_df, part_file_name in zip(part_dfs, part_filenames):
        part_df.to_csv(part_file_name, index=False)
        await asyncio.sleep(0.01)

    # with Pool(processes=min(num_parts, os.cpu_count())) as pool:
    #     _ = pool.map(save_part, list(zip(part_dfs, part_filenames)))

    return part_filenames


async def boost_split_csv(df: pd.DataFrame, file: str, data_len: int, num_parts: int, output_dir: str):
    rows_per_part = data_len // num_parts
    remainder = data_len % num_parts
    base_name = os.path.basename(file)
    name, ext = os.path.splitext(base_name)

    os.makedirs(output_dir, exist_ok=True)
    part_filenames = []

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
            await asyncio.sleep(0.01)
        
        for future in futures:
            future.result() 
            await asyncio.sleep(0.01)
    
    return part_filenames


def check_equal_row_count(dfs: List[pd.DataFrame]):
    row_counts = [len(df) for df in dfs]
    if len(set(row_counts)) > 1:
        raise ValueError("Error: CSV files have different number of rows.")
    return row_counts[0]
