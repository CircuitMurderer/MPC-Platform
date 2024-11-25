import os
import aiofiles

from pathlib import Path
from fastapi import HTTPException, UploadFile
from ..utils.file import file_summary
# from runner import DEFAULT_DIR_OUT


async def update_serv(
    file: UploadFile,
    id: str,
    party: str,
    DEFAULT_DIR_OUT: Path
):
    if not party in {'Alice', 'Bob', 'Result'}:
        raise HTTPException(
            status_code=400, 
            detail=f"'{party}' is not in ['Alice', 'Bob', 'Result']."
        )
    
    os.makedirs(DEFAULT_DIR_OUT / id, exist_ok=True)
    save_path = DEFAULT_DIR_OUT / id / f"{party}.csv"

    # buffer = bytes('', encoding='utf-8')
    # chunk_size = 1024 * 1024 * 1024 # 1GB

    buffer = await file.read()
    async with aiofiles.open(save_path, "wb") as f:
        await f.write(buffer)

    summary = await file_summary(file_cont=buffer)

    return {
        "status": "success",
        "message": f"File for {party} uploaded successfully.",
        "file_path": save_path,
        **summary
    }
