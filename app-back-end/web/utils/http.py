import json
import requests


def post_file(url, file_path, party, file_id):
    with open(file_path, 'rb') as file:
        files = {'file': file}
        response = requests.post(url, data={'id': file_id, 'party': party}, files=files)
    return response.text


def get_request(url, params):
    response = requests.get(url, params=params)
    return response


def check_exception(response: str):
    resp_json = json.loads(response)
    if 'error' in resp_json:
        raise ValueError(f"Error: {resp_json['error']}")
    