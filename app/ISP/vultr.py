from ISP.interface import ISPInterface
from typing import List
import vultr as _vultr
from time import sleep

class Vultr(ISPInterface):

    def __init__(
            self,
            api_key : str,
            safe_mode : bool = False
        ):


        self.vultr_client = _vultr.Vultr(api_key=api_key)
        self._update_ip_server_map()

    """
    return a list of public ip
    """
    def get_node_address_list(self):
        return self.ip_server_map.keys()

    """
    change specified ip and return new one
    """
    
    def update_node_address(self, oldip):
        if oldip not in self.ip_server_map:
            raise Exception("cannot get oldip information")
        sid = self.ip_server_map[oldip]

        # TODO: the following wont work, delete an instance 
        # and create a new one instead


        # create then destroy, or newip maybe same as oldip
        self.vultr_client.server.ipv4.create(sid)
        self.vultr_client.server.ipv4.destroy(sid,oldip)
        
        # wait and get new ip
        old_ip_server_map = self.ip_server_map
        new_ip_server_map = old_ip_server_map
        while new_ip_server_map == old_ip_server_map:
            sleep(1)
            new_ip_server_map = self._update_ip_server_map()
        newip = next(ip for ip in new_ip_server_map if ip not in old_ip_server_map)
        
        # reboot the server 
        # reboot does work, halt and start instead
        self.vultr_client.server.halt(sid)
        print(self.vultr_client.server.list(sid))\
    
        return newip

    def _update_ip_server_map(self):
        self.ip_server_map = {}
        server_list = self.vultr_client.server.list()
        for sid in server_list:
            ips = self.vultr_client.server.ipv4.list(sid)[sid]
            if len(ips) > 2 and self.safe_mode:
                raise Exception("more than one secondary ip detected on" +\
                  f"server {server_list[sid]['label']}, exiting now")
            secondary_ips = [ip["ip"] for ip in ips if ip["type"] == "secondary_ip"]
            self.ip_server_map.update({ip:sid for ip in secondary_ips})
        return self.ip_server_map

    