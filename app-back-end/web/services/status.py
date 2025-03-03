from typing import Dict, Any
from fastapi import HTTPException
# from runner import tasks


async def status_serv(
    id: str, 
    tasks: Dict[str, Any]
) -> Dict[str, Any]:
    task_id = id

    if task_id not in tasks:
        raise HTTPException(status_code=404, detail=f"Task (ID = '{id}') not found, or created failed.")

    task = tasks[task_id]
    if task["status"] == "running":
        return {
            "status": "success",
            "task_id": task_id,
            "task_stat": task["status"],
            "task_info": task["info"],
            "task_stage": task["stage"]
        }
    elif task["status"] == "completed":
        return {
            "status": "success",
            "task_id": task_id,
            "task_stat": task["status"],
            "task_info": task["info"],
            "task_result": task["checked"]
        }
    elif task["status"] == "failed":
        return {
            "status": "success",
            "task_id": task_id,
            "task_stat": task["status"],
            "error": task["error"]
        }
    else:
        return {
            "status": "success",
            "task_id": task_id,
            "task_stat": "unknown",
        }
