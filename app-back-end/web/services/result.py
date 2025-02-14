from pathlib import Path
from fastapi import HTTPException
from fastapi.responses import FileResponse
# from runner import DEFAULT_DIR_OUT


async def result_serv(
    id: str, 
    DEFAULT_DIR_OUT: Path
) -> FileResponse:
    file_path = DEFAULT_DIR_OUT / id / "Verified.csv"
        
    if not file_path.exists():
        raise HTTPException(
            status_code=404, 
            detail=f"Verified data is not found for ID '{id}'."
        )

    return FileResponse(file_path, media_type="text/csv", filename=f"{id}_verified.csv")
