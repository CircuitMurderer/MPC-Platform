from fastapi import HTTPException
from runner import tasks


async def status_serv(id: str):
    task_id = id
    global tasks

    if task_id not in tasks:
        raise HTTPException(status_code=404, detail=f"Task (ID = '{id}') not found, or created failed.")

    task = tasks[task_id]
    if task["status"] == "running":
        return {
            "status": "success",
            "task_id": task_id,
            "task_stat": "running",
            "task_info": task["info"],
            "task_stage": task["stage"]
        }
    elif task["status"] == "completed":
        return {
            "status": "success",
            "task_id": task_id,
            "task_stat": "completed",
            "task_info": task["info"],
            "task_result": task["result"]
        }
    elif task["status"] == "failed":
        return {
            "status": "failed",
            "task_id": task_id,
            "error": task["error"]
        }
    else:
        return {
            "status": "unknown",
            "task_id": task_id,
        }