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

    def _update_up_list(self):
        self.ip_list = []
        for ip_ocid in self.config["private_ip_ocid"]:
            self.ip_list.append(self._get_public_ip(ip_ocid))
        return self.ip_list

    """
    return a list of public ip
    """
    def list_ip(self):
        self.ip_list = self._update_up_list()
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
        self._delete_public_ip(old_ip_data.id)

        # reassign public ip
        self._create_public_ip(old_ip_data.private_ip_id, old_ip_data.lifetime)
    
        # update the list
        new_ip_data = self._get_public_ip(old_ip_data.private_ip_id)
        self.ip_list.remove(old_ip_data)
        self.ip_list.append(new_ip_data)

        return new_ip_data.ip_address

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


    def _delete_public_ip(self, public_ip_ocid):
        delete_public_ip_response = self.network_client.delete_public_ip(
            public_ip_id=public_ip_ocid)   
        return delete_public_ip_response.headers

    def _create_public_ip(self, private_ip_ocid, lifetime):
        create_public_ip_response = self.network_client.create_public_ip(
            create_public_ip_details=oci.core.models.CreatePublicIpDetails(
                compartment_id=self.config["auth"]["tenancy_ocid"],
                lifetime=lifetime,
                private_ip_id=private_ip_ocid)
        )

        return create_public_ip_response.data  

