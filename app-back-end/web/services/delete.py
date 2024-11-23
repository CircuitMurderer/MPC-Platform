import os
import shutil

from typing import Optional
from runner import DEFAULT_DIR_OUT, DEFAULT_CAL_DIR


async def delete_serv(id: Optional[str]):
    if id is not None:
        origin_path = os.path.join(DEFAULT_DIR_OUT, id)
        if os.path.exists(origin_path):
            shutil.rmtree(origin_path)
            origin_message = f"Deleted folder: {origin_path}"
        else:
            origin_message = f"Folder not found: {origin_path}"

        deleted_data_folders = []
        for folder in os.listdir(DEFAULT_CAL_DIR):
            if folder.startswith(f"{id}_"):
                folder_path = os.path.join(DEFAULT_CAL_DIR, folder)
                shutil.rmtree(folder_path)
                deleted_data_folders.append(folder_path)

        data_message = (
            f"Deleted folders: {', '.join(deleted_data_folders)}"
            if deleted_data_folders
            else f"No folders found in {DEFAULT_CAL_DIR} starting with '{id}_'."
        )

        return {
            "status": "success",
            "origin": origin_message,
            "data": data_message,
        }
    
    else:
        if os.path.exists(DEFAULT_DIR_OUT):
            shutil.rmtree(DEFAULT_DIR_OUT)
            os.makedirs(DEFAULT_DIR_OUT)

        if os.path.exists(DEFAULT_CAL_DIR):
            shutil.rmtree(DEFAULT_CAL_DIR)
            os.makedirs(DEFAULT_CAL_DIR)

        return {
            "status": "success",
            "message": f"All files and folders in {DEFAULT_DIR_OUT} and {DEFAULT_CAL_DIR} have been deleted.",
        }
