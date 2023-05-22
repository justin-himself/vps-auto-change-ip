from CDN.interface import CDNInterface
from cloudflare_ddns import CloudFlare

class Cloudflare(CDNInterface):

    def __init__(
            self, 
            email:str,
            api_key:str,
            domain:str
        ):

        self.cf = CloudFlare(
            email=email,
            api_key=api_key,
            domain=domain
        )

    def get_record_list(self):
        self.cf.refresh()
        subdomain_ip_map = {}
        for subdomain in self.config["subdomains"]:
            record = self.cf.get_record("A", subdomain)
            subdomain_ip_map[subdomain] = record["content"]
        return subdomain_ip_map
        
    def update_record(self, record, new_ip_addr):
        self.cf.update_record('A', record, new_ip_addr)