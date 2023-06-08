from Pinger.interface import PingerInterface
from typing import Tuple
import tcppinglib
import asyncio

class TcpPing(PingerInterface):

    """
    The address should be in format like addr:tcpport
    TODO:
        The if a lot of address failed to respond, the mulitping time will be very long
    """

    TEST_ADDR_EXPECTATIONS = [
        ("baidu.com:80", True), # China domain
        ("1.1.1.1:53", True), # Foreign IP unblocked
        ("142.250.66.238:80", False) # Google IP blocked
    ]

    def ping(self, address_list):

        test_list = [x[0] for x in self.TEST_ADDR_EXPECTATIONS]
        merged_results = [x.is_alive for x in TcpPing.multi_tcpping_with_port(address_list + test_list)]
        ping_result = {address_list[idx]:merged_results[idx] for idx in range(len(address_list))}
        test_result = {test_list[idx]:merged_results[len(address_list) + idx] for idx in range(len(test_list))}
        
        # test the expectations 
        for (key,val) in self.TEST_ADDR_EXPECTATIONS:
            if test_result[key] != val:
                raise Exception(f"Ping network error, {key} returns a result of {test_result[key]}")

        return ping_result
    
    @staticmethod
    def seperate_ip_port(tcp_address : str, default_port = 80) -> Tuple[str, str]:
        if ":" not in tcp_address or tcp_address.endswith("]"):
            return tcp_address, default_port
        
        ip = ":".join(tcp_address.split(":")[:-1])
        port = "".join(tcp_address.split(":")[-1])
        return ip, port
        
    @staticmethod
    def multi_tcpping_with_port(address_list):
        """
        A wrapper for multi_tcpping provided by tcppinglib
        that enables multiping with different ports
        """
        async def __async_multi_tcpping(addresses: list):
            TIMEOUT: float = 2
            COUNT: int = 2
            INTERVAL: float = 0.2
            CONCURRENT_TASKS=50
            loop = asyncio.get_running_loop()
            tasks = []
            tasks_pending = set()
            for address in addresses:

                ip, port = TcpPing.seperate_ip_port(address)
                
                if len(tasks_pending) >= CONCURRENT_TASKS:
                    _, tasks_pending = await asyncio.wait(
                        tasks_pending, return_when=asyncio.FIRST_COMPLETED
                    )
                task = loop.create_task(tcppinglib.async_tcpping(ip, port, TIMEOUT, COUNT, INTERVAL))
                tasks.append(task)
                tasks_pending.add(task)
            await asyncio.wait(tasks_pending)
            return [task.result() for task in tasks]

        return asyncio.run(
            __async_multi_tcpping(
                addresses=address_list
            )
        )