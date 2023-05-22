from Pinger.interface import PingerInterface
import os, icmplib

class ICMPPing(PingerInterface):

    """
    Pinger implemented using icmplib, requires
    root privilege to function.
    """

    TEST_ADDR_EXPECTATIONS = [
        ("101.6.6.6", True), # China IP
        ("baidu.com", True), # China domain
        ("1.1.1.1", True), # Unblocked Foreign IP
        ("142.250.66.238", False) # Blocked Google IP
    ]

    def __init__(self):

        # root detection
        if os.getuid() != 0:
            raise Exception("ICMP requires root privilege.")


    def ping(self, address_list):

        test_list = [x[0] for x in self.TEST_ADDR_EXPECTATIONS]
        merged_results = [x.is_alive for x in icmplib.multiping(address_list + test_list)]
        ping_result = {address_list[idx]:merged_results[idx] for idx in range(len(address_list))}
        test_result = {test_list[idx]:merged_results[len(address_list) + idx] for idx in range(len(test_list))}
        
        # test the expectations 
        for (key,val) in self.TEST_ADDR_EXPECTATIONS:
            if test_result[key] != val:
                raise Exception(f"Ping network error, {key} returns a result of {test_result[key]}")

        return ping_result
        