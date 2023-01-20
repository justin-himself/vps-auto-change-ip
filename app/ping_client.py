import yaml
import requests  
from ipaddress import ip_address
from icmplib import multiping

def local_ping(ipaddress_list : list, config):

    contains_v6 = False

    for a in ipaddress_list:
        try:
            ip_validator = ip_address(a)
        except:
            raise Exception(f"invalid ip address {a}")

        if not ip_validator.is_global:
            raise Exception(f"{a} not a global ip address")
        
        if ":" in a:
            contains_v6 = True
    
    test_list = [
        config["ping"]["test_address"]["china_ip"],
        config["ping"]["test_address"]["china_ipv6"],
        config["ping"]["test_address"]["foreign_ip_unblocked"],
        config["ping"]["test_address"]["foreign_ip_blocked"]
    ]

    results = multiping(ipaddress_list + test_list)
    results = [x.is_alive for x in results]


    if not results[-4]:
        raise Exception("cannot access internet")

    if contains_v6 and not results[-3]:
        raise Exception("not capable of ping ipv6 addresses")

    if not results[-2]:
        raise Exception("internet access is restricted")
    
    if results[-1]:
        raise Exception("not in china or icmp traffic proxied")

    return results[:-4]

def remote_ping(ipaddress_list : list, config):
    addr = config["ping"]["client_config"]["server_addr"]
    token = config["ping"]["client_config"]["server_token"]
    args = "&addr=" + "&addr=".join(ipaddress_list)
    response = requests.get(f"http://{addr}/ping/?token={token}" + args)
    if response.status_code != 200:
        raise Exception(response.text)
    return [ip for ip in response.json().values()]
