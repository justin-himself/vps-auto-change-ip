from cloudflare_ddns import CloudFlare


class Cloudflare:

    def __init__(self, config):

        self.config = config["panel"]["cloudflare"]
        self.cf = CloudFlare(self.config["email"],self.config["api_key"],self.config["domain"])

    def get_records(self) -> dict[str, str]:
        self.cf.refresh()
        subdomain_ip_map = {}
        for subdomain in self.config["subdomains"]:
            record = self.cf.get_record("A", subdomain)
            subdomain_ip_map[subdomain] = record["content"]
        return subdomain_ip_map

    def update_ip(self, oldip, newip):
        subdomain_ip_map = self.get_records()
        for subdomain, ip in subdomain_ip_map.items():
            if ip == oldip:
                self.cf.update_record('A', subdomain, newip)