from Pinger.interface import PingerInterface
import requests

class RemotePing(PingerInterface):

    def __init__(
            self,
            server_addr:str,
            server_token:str,
            ping_method:str = "icmpping"
        ):

        self.addr = server_addr
        self.token = server_token
        self.method = ping_method

    def ping(self, address_list):
        
        args = "&addr=" + "&addr=".join(address_list)
        response = requests.get(f"{self.addr}/ping/?token={self.token}?method={self.method}" + args)
        if response.status_code != 200:
            raise Exception(response.text)
        return {address_list[idx]:response.json().values()[idx] for idx in range(len(address_list))}