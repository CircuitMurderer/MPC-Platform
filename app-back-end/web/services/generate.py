import os
import io
import csv
import random
import aiofiles
import numpy as np

from pathlib import Path
from typing import Dict, Any
from fastapi import HTTPException

from ..utils.file import file_summary


def generate_csv(N: int) -> bytes:
    output = io.BytesIO()
    writer = csv.writer(output, delimiter=',', quotechar='"', quoting=csv.QUOTE_MINIMAL)
    writer.writerow(['number', 'data'])
    
    for i in range(1, N + 1):
        data = random.uniform(0, 1)
        writer.writerow([i, data])

    output.seek(0)
    return output.read() 


def read_csv_data_from_bytes(data_bytes: bytes) -> np.ndarray:
    csv_str = data_bytes.decode('utf-8')
    f = io.StringIO(csv_str)
    reader = csv.reader(f)
    next(reader)

    data = []
    for row in reader:
        data.append(float(row[1]))

    return np.array(data)


def apply_operation(data1: np.ndarray, data2: np.ndarray, operator: str) -> np.ndarray:
    if operator == 'add':
        return data1 + data2
    elif operator == 'sub':
        return data1 - data2
    elif operator == 'mul':
        return data1 * data2
    elif operator == 'div':
        return np.divide(data1, data2, out=np.full_like(data1, np.inf), where=data2!=0)
    elif operator == 'exp':
        return np.power(data1, data2)
    else:
        raise ValueError(f"Unsupported operator: {operator}")


def calculate_result_data(alice_data: np.ndarray, bob_data: np.ndarray, operator: str) -> bytes:
    output = io.BytesIO()
    writer = csv.writer(output)
    writer.writerow(['number', 'operation', 'data'])

    result = apply_operation(alice_data, bob_data, operator)
    for i, res in enumerate(result, 1):
        writer.writerow([i, res])
    
    output.seek(0)
    return output.read()


async def gen_serv(
    id: str,
    operator: str,
    data_length: int,
    tasks: Dict[str, Any],
    OPERA_DICT: Dict[str, int],
    DEFAULT_DIR_OUT: Path
):
    if operator not in OPERA_DICT:
        raise HTTPException(
            status_code=400, 
            detail=f"Invalid operator: {operator}. Valid options are: {list(OPERA_DICT.keys())}."
        )
    
    os.makedirs(DEFAULT_DIR_OUT / id, exist_ok=True)
    party_datas = {}

    for party in ['Alice', 'Bob']:
        save_path = DEFAULT_DIR_OUT / id / f"{party}.csv"

        buffer = generate_csv(data_length)
        async with aiofiles.open(save_path, "wb") as f:
            await f.write(buffer)

        summary = await file_summary(file_cont=buffer)
        if not id in tasks:
            tasks[id] = {}
        tasks[id]["length"] = summary["items"]
        tasks[id][party] = {"summary": summary}

        party_datas[party] = read_csv_data_from_bytes(buffer)

    result_data = calculate_result_data(party_datas['Alice'], party_datas['Bob'])
    result_path = DEFAULT_DIR_OUT / id / 'Result.csv'
    async with aiofiles.open(result_path, 'wb') as file:
        await file.write(result_data)

    summary = await file_summary(file_cont=result_data)
    if not id in tasks:
        tasks[id] = {}
    tasks[id]["length"] = summary["items"]
    tasks[id]["Result"] = {"summary": summary}

    return {
        "status": "success",
        "message": f"File for id='{id}' generated successfully.",
        "data_length": summary["items"]
    }