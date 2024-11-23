import os
import asyncio

from pathlib import Path
from typing import Optional, Dict

from fastapi import FastAPI, HTTPException, UploadFile, Form, Query
from fastapi.middleware.cors import CORSMiddleware


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

DEFAULT_DIR_OUT = Path("./base/")
DEFAULT_CAL_DIR = Path("./data/")
DEFAULT_COMBINED_FILE = "combinedResult.csv"
DEFAULT_URI = "http://localhost:9000"

os.makedirs(DEFAULT_DIR_OUT, exist_ok=True)
os.makedirs(DEFAULT_CAL_DIR, exist_ok=True)

tasks: Dict[str, dict] = {}
app = FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"], 
    allow_credentials=True,
    allow_methods=["*"], 
    allow_headers=["*"], 
)


from web.services.update import update_serv
from web.services.verify import verify_serv
from web.services.result import result_serv
from web.services.delete import delete_serv
from web.services.status import status_serv


@app.post("/update")
async def file_update(
    file: UploadFile,
    id: str = Form("test"),
    party: str = Form(...),
):
    try:
        return await update_serv(file, id, party)
    
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
    

@app.get("/verify")
async def run_process(
    id: str = "test",
    operator: Optional[str] = None,
    operate: Optional[int] = None,
    split_n: int = 0,
    workers: int = 8,
    scale: int = 1,
):
    try:
        asyncio.create_task(verify_serv(id, operator, operate, split_n, workers, scale, True))
        return {
            "status": "success",
            "message": f"Verify task (ID = '{id}') has been started."
        }
    
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
    

@app.get("/result")
async def get_result(id: str = "test"):
    try:
        return await result_serv(id)
    
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
    

@app.get("/delete")
async def delete_files(id: Optional[str] = Query(default=None)):
    try:
        return await delete_serv(id)
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
    

@app.get("/stat")
async def get_task_stat(id: str = Query(...)):
    try:
        return await status_serv(id)
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
    