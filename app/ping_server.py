import yaml
import os, sys
from typing import List, Union
from fastapi import FastAPI, Query
from starlette.responses import JSONResponse
from Pinger.icmpping import Icmpping
from Pinger.tcpping import Tcpping

# load_config 
with open("config/config.yaml", "r") as f:
    config = yaml.load(f, Loader=yaml.SafeLoader)["ping"]

app = FastAPI(docs_url="/")

@app.get("/ping/")
async def ping(
        token:str = Query(
            default=None, 
            description="api token"
        ),
        addr: Union[List[str], None] = Query(
            default=None,  example="101.6.6.6",
            description="public ip address"
        )
    ):

    global config

    if token not in config["server_config"]["token"]:
        return JSONResponse(
            status_code=403,
            content={"message": "authenticate failed"},
        )
    
    test_list = [
        config["test_address"]["china_ip"],
        config["test_address"]["china_ipv6"],
        config["test_address"]["foreign_ip_unblocked"],
        config["test_address"]["foreign_ip_blocked"]
    ]

    results = await async_multiping(addr + test_list)
    results = [x.is_alive for x in results]


    if not results[-4]:
        return JSONResponse(
            status_code=500,
            content={"message": "server cannot access internet"},
        )

    if not results[-2]:
        return JSONResponse(    
            status_code=500,
            content={"message": "server internet access is restricted"},
        )
    
    if results[-1]:
        return JSONResponse(
            status_code=500,
            content={"message": "server not in china or icmp traffic proxied"},
        )

    return {addr[idx]:results[idx] for idx in range(len(addr))}

