import os
import io
import csv
import random

import aiofiles
import numpy as np
import dask.array as da
import dask.dataframe as dd

from pathlib import Path
from typing import Dict, Any
from fastapi import HTTPException

from ..utils.data import data_summary


def generate_df(N):
    df = dd.from_array(da.arange(1, N + 1), columns=['number'])
    df['data'] = da.random.uniform(0, 1, size=N)
    df = df.set_index('number')
    return df


def do_calculate(df_a, df_b, operator):
    data_a = da.from_array(df_a['data'].compute())
    data_b = da.from_array(df_b['data'].compute())

    df_r = dd.from_array(da.arange(1, data_a.shape[0] + 1), columns=['number'])
    df_r = df_r.set_index('number')

    if operator.lower() == 'add':
        data_r = data_a + data_b
    elif operator.lower() == 'sub':
        data_r = data_a - data_b
    elif operator.lower() == 'mul':
        data_r = data_a * data_b
    elif operator.lower() == 'div':
        data_r = data_a / data_b
    elif operator.lower() == 'exp':
        data_r = data_a ** data_b
    else:
        raise ValueError(f"Unsupported operator: {operator}")

    df_r['data'] = data_r.compute()
    return df_r


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
    save_paths = {
        'A': DEFAULT_DIR_OUT / id / "Alice.csv",
        'B': DEFAULT_DIR_OUT / id / "Bob.csv",
        'R': DEFAULT_DIR_OUT / id / "Result.csv",
    }

    df_a = generate_df(data_length)
    df_b = generate_df(data_length)
    df_r = do_calculate(df_a, df_b, operator)

    summary_a = data_summary(df_a)
    summary_b = data_summary(df_b)
    summary_r = data_summary(df_r)

    df_a.to_hdf(save_paths["A"], key='data', mode='w')
    df_b.to_hdf(save_paths["B"], key='data', mode='w')
    df_r.to_hdf(save_paths["R"], key='data', mode='w')

    if not id in tasks:
        tasks[id] = {}

    tasks[id]["length"] = data_length
    tasks[id]["Alice"] = {"summary": summary_a}
    tasks[id]["Bob"] = {"summary": summary_b}
    tasks[id]["Result"] = {"summary": summary_r}

    return {
        "status": "success",
        "task_id": id,
        "message": f"File for id='{id}' generated successfully.",
        "data_length": data_length,
        "summary": {
            "a": summary_a,
            "b": summary_b,
            "r": summary_r
        }
    }