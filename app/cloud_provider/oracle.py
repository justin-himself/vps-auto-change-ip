import oci
import logging

class Oracle:

    def __init__(self, config):

        self.config = config["cloud_provider"]["oracle"]

        self.oci_config = {
            "user": self.config["auth"]["user_ocid"],
            "key_file": self.config["auth"]["key_file"],
            "fingerprint": self.config["auth"]["fingerprint"],
            "tenancy": self.config["auth"]["tenancy_ocid"],
            "region": self.config["auth"]["region"]
        }

        self.network_client = oci.core.VirtualNetworkClient(self.oci_config)   

    def _update_ip_list(self):
        
        self.ip_list = []
        
        for vnic_ocid in self.config["vnic_ocid"]:
            private_ip_data = self._list_private_ip(vnic_ocid)
            self.ip_list += [self._get_public_ip(x.id) for x in private_ip_data]
            
            if self.config["enable_ipv6"]:
                self.ip_list += self._list_ipv6(vnic_ocid)

        return self.ip_list

    """
    return a list of public ip
    """
    def list_ip(self):
        self.ip_list = self._update_ip_list()
        results = [item.ip_address for item in self.ip_list]
        return results

    """
    change specified ip and return new one
    """
    def change_ip(self, old_ip):

        # get ip information
        old_ip_data = None
        for item in self.ip_list:
            if item.ip_address == old_ip:
                old_ip_data = item
                break
        if old_ip_data is None:
            raise Exception("cannot get oldip information")

        # delete public ip
        if ":" in old_ip:
            self._delete_ipv6(old_ip_data.id)
        else:
            self._delete_public_ip(old_ip_data.id)

        # reassign public ip
        if ":" in old_ip:
            new_ip_data = self._create_ipv6(old_ip_data.vnic_id)
        else:
            new_ip_data = self._create_public_ip(old_ip_data.private_ip_id, old_ip_data.lifetime)
    
        # update the list
        self.ip_list.remove(old_ip_data)
        self.ip_list.append(new_ip_data)

        return new_ip_data.ip_address

    def _list_private_ip(self, vnic_ocid):
        list_private_ips_response = self.network_client.list_private_ips(
            vnic_id=vnic_ocid)

        return list_private_ips_response.data

    def _list_ipv6(self, vnic_ocid):
        list_ipv6s_response = self.network_client.list_ipv6s(
            vnic_id=vnic_ocid)
        return list_ipv6s_response.data

    def _get_private_ip(self, private_ip_ocid):
        get_private_ip_response = self.network_client.get_private_ip(
            private_ip_id=private_ip_ocid)
        return get_private_ip_response.data

    # https://docs.oracle.com/iaas/api/#/en/iaas/latest/PublicIp/GetPublicIpByPrivateIpId
    def _get_public_ip(self, private_ip_ocid):

            get_public_ip_by_private_ip_id_response = self.network_client.get_public_ip_by_private_ip_id(
                get_public_ip_by_private_ip_id_details=oci.core.models.GetPublicIpByPrivateIpIdDetails(
                    private_ip_id=private_ip_ocid
                )
            )

            return get_public_ip_by_private_ip_id_response.data

    def _get_public_ipv6(self, ipv6_ocid):

        get_ipv6_response = self.network_client.get_ipv6(
            ipv6_id=ipv6_ocid,
            )
    
        return get_ipv6_response.data


    def _delete_public_ip(self, public_ip_ocid):
        delete_public_ip_response = self.network_client.delete_public_ip(
            public_ip_id=public_ip_ocid)   
        return delete_public_ip_response.headers

    def _delete_ipv6(self, ipv6_ocid):
        delete_ipv6_response = self.network_client.delete_ipv6(
            ipv6_id=ipv6_ocid)
        return delete_ipv6_response.headers

    def _create_public_ip(self, private_ip_ocid, lifetime):
        create_public_ip_response = self.network_client.create_public_ip(
            create_public_ip_details=oci.core.models.CreatePublicIpDetails(
                compartment_id=self.config["auth"]["tenancy_ocid"],
                lifetime=lifetime,
                private_ip_id=private_ip_ocid)
        )
        return create_public_ip_response.data  
    
    def _create_ipv6(self, vnic_ocid):
        create_ipv6_response = self.network_client.create_ipv6(
            create_ipv6_details=oci.core.models.CreateIpv6Details(
                vnic_id=vnic_ocid))
        return create_ipv6_response.data  


