import os
import time
import asyncio
import pandas as pd

from pathlib import Path
from typing import Optional, Dict, Any
from fastapi import HTTPException

from ..utils.file import split_csv, check_equal_row_count, process_files, combine_results
# from runner import tasks, OPERA_DICT, DEFAULT_DIR_OUT, DEFAULT_URI


async def verify_serv(
    id: str,
    operator: Optional[str],
    operate: Optional[int],
    split_n: int,
    workers: int,
    scale: int,
    tasks: Dict[str, Any],
    OPERA_DICT: Dict[str, int],
    DEFAULT_DIR_OUT: Path,
    DEFAULT_URI: str,
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
                    "desc": "Checking data files.",
                    "sub_stage": ""
                },
                **tasks[id]
            }

        df_a = pd.read_csv(file_name_a)
        df_b = pd.read_csv(file_name_b)
        df_r = pd.read_csv(file_name_r)

        row_count = check_equal_row_count([df_a, df_b, df_r])
        if split_n < 0 or split_n > row_count:
            raise HTTPException(
                status_code=400, 
                detail=f"split_n must be between 0 and {row_count}."
            )
        
        if split_n == 0:    # Auto detect
            split_n = row_count // 1000000 if row_count > 1000000 else 1

        os.makedirs(base_path / "split", exist_ok=True)
        os.makedirs(base_path / "temp", exist_ok=True)

        split_files = {}

        if is_async:
            tasks[id]["stage"] = "2/4"
            tasks[id]["info"] = {
                "desc": "Splitting original data:",
                "sub_stage": "1/3 - data of Alice."
            }        
        split_files['A'] = await split_csv(df_a, "Alice.csv", split_n, base_path / "split")

        if is_async:
            tasks[id]["info"]["sub_stage"] = "2/3 - data of Bob."
        split_files['B'] = await split_csv(df_b, "Bob.csv", split_n, base_path / "split")

        if is_async:
            tasks[id]["info"]["sub_stage"] = "3/3 - data of Result."
        split_files['R'] = await split_csv(df_r, "Result.csv", split_n, base_path / "split")

        result_file_names = []
        difference, comm_cost, time_cost = 0, 0, 0.

        if is_async:
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
                await asyncio.sleep(0.5)
            else:
                time.sleep(0.5)

        if is_async:
            tasks[id]["stage"] = "4/4"
            tasks[id]["info"] = {
                "desc": "Combining verified results.",
                "sub_stage": ""
            } 
        combine_results(result_file_names, base_path / "Verified.csv")

        verify_result = {
            "status": "success",
            "data_length": tasks[id]["length"],
            "checked_errors": difference,
            "error_rate": f"{round(float(difference) / tasks[id]['length'], 4) * 100}%",
            "comm_cost": comm_cost,
            "time_cost": time_cost,
            # "combined_file": base_path / "Verified.csv",
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
