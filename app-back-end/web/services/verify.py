import os
import time
import secrets

import asyncio
import numpy as np
import pandas as pd

from pathlib import Path
from typing import Optional, Dict, Any
from fastapi import HTTPException

from ..utils.file import check_equal_row_count, process_files, combine_results, boost_split_csv
from ..utils.data import get_sample_size

async def verify_serv(
    id: str,
    operator: Optional[str],
    operate: Optional[int],
    split_n: int,
    workers: int,
    scale: int,
    conf_level: float,
    error_rate: float,
    tasks: Dict[str, Any],
    OPERA_DICT: Dict[str, int],
    DEFAULT_DIR_OUT: Path,
    DEFAULT_URI: str,
    is_csv: bool = True,
    check_all: bool = False,
    is_async: bool = True,
):
    try:
        if operator:
            if operator not in OPERA_DICT:
                raise HTTPException(
                    status_code=400, 
                    detail=f"Invalid operator: {operator}. Valid options are: {list(OPERA_DICT.keys())}."
                )
            operate = OPERA_DICT[operator.lower()] 
        elif operate is None:
            raise HTTPException(
                status_code=400, 
                detail="Either 'operator' or 'operate' must be specified."
            )

        base_path = DEFAULT_DIR_OUT / id
        if not base_path.exists():
            raise HTTPException(
                status_code=400, 
                detail=f"Origin data of ID '{id}' haven't been uploaded."
            )

        file_name_a = base_path / "Alice.csv"
        file_name_b = base_path / "Bob.csv"
        file_name_r = base_path / "Result.csv"

        if is_async:
            tasks[id] = {
                "status": "running",
                "stage": "1/4",
                "info": {
                    "desc": "Checking and applying data files.",
                    "sub_stage": ""
                },
                **tasks[id]
            }

        if is_csv:
            origin_df_a = pd.read_csv(file_name_a).set_index('number')
            origin_df_b = pd.read_csv(file_name_b).set_index('number')
            origin_df_r = pd.read_csv(file_name_r).set_index('number')

        else:
            origin_df_a = pd.read_hdf(file_name_a, key='data')
            origin_df_b = pd.read_hdf(file_name_b, key='data')
            origin_df_r = pd.read_hdf(file_name_r, key='data')

        row_count = check_equal_row_count([origin_df_a, origin_df_b, origin_df_r])
        if split_n < 0 or split_n > row_count:
            raise HTTPException(
                status_code=400, 
                detail=f"split_n must be between 0 and {row_count}."
            )
        
        sample_size = get_sample_size(conf_level, error_rate, row_count)
        if row_count <= 100_0000 or check_all:
            sample_size = row_count
            sample_indexes = np.arange(1, row_count + 1).tolist()
        else:
            np.random.seed(secrets.randbelow(2 ** 32 - 2))
            sample_indexes = np.random.choice(np.arange(1, row_count + 1), size=sample_size, replace=False).tolist()
        
        await asyncio.sleep(0.01)
        df_a: pd.DataFrame = origin_df_a.loc[sample_indexes]

        await asyncio.sleep(0.01)
        df_b: pd.DataFrame = origin_df_b.loc[sample_indexes]

        await asyncio.sleep(0.01)
        df_r: pd.DataFrame = origin_df_r.loc[sample_indexes]
            
        if split_n == 0:   # auto detect
            split_n = sample_size // 100_0000
            split_n += 1 if sample_size % 100_0000 != 0 else 0

        os.makedirs(base_path / "split", exist_ok=True)
        os.makedirs(base_path / "temp", exist_ok=True)

        split_files = {}

        if is_async:
            tasks[id]["status"] = "running"
            tasks[id]["stage"] = "2/4"
            tasks[id]["info"] = {
                "desc": "Splitting original data:",
                "sub_stage": "1/3 - data of Alice."
            }        
        split_files['A'] = await boost_split_csv(df_a, "Alice.csv", sample_size, split_n, base_path / "split")

        if is_async:
            tasks[id]["info"]["sub_stage"] = "2/3 - data of Bob."
        split_files['B'] = await boost_split_csv(df_b, "Bob.csv", sample_size, split_n, base_path / "split")

        if is_async:
            tasks[id]["info"]["sub_stage"] = "3/3 - data of Result."
        split_files['R'] = await boost_split_csv(df_r, "Result.csv", sample_size, split_n, base_path / "split")

        result_file_names = []
        difference, comm_cost, time_cost = 0, 0, 0.

        if is_async:
            tasks[id]["status"] = "running"
            tasks[id]["stage"] = "3/4"
            tasks[id]["info"] = {
                "desc": "Verifying calculated results:",
                "sub_stage": f"0/{split_n} - batch data."
            } 

        for x in range(split_n):
            if is_async:
                tasks[id]["info"]["sub_stage"] = f"{x + 1}/{split_n} - batch data."

            part_files = {
                'A': split_files['A'][x],
                'B': split_files['B'][x],
                'R': split_files['R'][x],
            }

            f_name, diff, c_cost, t_cost = await process_files(
                part_files, 
                task_id=id,
                file_id=x + 1, 
                operate=operate, 
                workers=workers, 
                scale=scale, 
                result_dir=base_path / "temp", 
                base_url=DEFAULT_URI
            )

            result_file_names.append(f_name)
            difference += diff
            comm_cost += c_cost
            time_cost += t_cost
            if is_async:
                await asyncio.sleep(0.1)
            else:
                time.sleep(0.1)

        if is_async:
            tasks[id]["status"] = "running"
            tasks[id]["stage"] = "4/4"
            tasks[id]["info"] = {
                "desc": "Combining verified results.",
                "sub_stage": ""
            } 
        combine_results(result_file_names, base_path / "Verified.csv")
        
        total_mistake_rate = float(difference) / sample_size
        mistake_rate = f'{round(total_mistake_rate * 100, 4)}% ± {round(error_rate * 100, 2)}%'
        mistakes = f'{max(round(row_count * total_mistake_rate) - round(row_count * error_rate), 0)} ~ ' \
                   f'{min(round(row_count * total_mistake_rate) + round(row_count * error_rate), row_count)}'

        if row_count <= 100_0000 or check_all:
            error_rate = 0.
            conf_level = 1.
            mistakes = f'{difference}'
            mistake_rate = f'{round(total_mistake_rate * 100, 4)}%'

        verify_result = {
            "status": "success",
            "data_length": row_count,
            "operate_between": operate,
            "checked_mistakes": mistakes,
            "mistake_rate": mistake_rate,
            "conf_level": f'{conf_level * 100}%',
            "error_rate": f'{error_rate * 100}%',
            "comm_cost": f'{comm_cost} bits',
            "time_cost": f'{round(time_cost, 4)} ms',
        }

        if is_async:
            tasks[id]["status"] = "completed"
            tasks[id]["stage"] = "done"
            tasks[id]["checked"] = verify_result
            tasks[id]["info"] = {
                "desc": "Verify all done.",
                "sub_stage": ""
            }

        return verify_result
    
    except Exception as e:
        tasks[id] = {
            "status": "failed",
            "error": str(e)
        }
        raise HTTPException(status_code=500, detail=str(e))
