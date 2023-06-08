import yaml
from typing import List, Union
from fastapi import FastAPI, Query
from starlette.responses import JSONResponse
import utils

# load_config 
with open("config/ping_server.yaml", "r") as f:
    config = yaml.load(f, Loader=yaml.SafeLoader)["ping"]

app = FastAPI(docs_url="/")

@app.get("/ping/")
async def ping(
        token:str = Query(
            default=None, 
            description="api token"
        ),
        method:str = Query(
            default="icmp",
            description="ping method"
        ),
        addr: Union[List[str], None] = Query(
            default=None,  example="101.6.6.6",
            description="public ip address"
        )
    ):

    global config

    if token not in config["token"]:
        return JSONResponse(
            status_code=403,
            content={"message": "authenticate failed"},
        )
    
    if method not in ["icmp", "tcpping"]:
        return JSONResponse(
            status_code=400,
            content={"message": "method not allowed or not exist"}
        )
    
    
    pinger = utils.load_module("Pinger",method)

    try:
        results = await pinger.ping(addr)
        results = [x.is_alive for x in results]
    except Exception as e:
        error_msg = str(e)
        return JSONResponse(
            status_code=400,
            content={"message:", f"An error occured: {error_msg}"}
        )

    return {addr[idx]:results[idx] for idx in range(len(addr))}

