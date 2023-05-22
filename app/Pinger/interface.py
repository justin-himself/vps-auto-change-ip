from abc import ABC, abstractmethod
from typing import List, Dict

class PingerInterface(ABC):

    @abstractmethod
    def ping(self, address_list: List[str]) -> Dict[str,bool]:
        """ Test the status of a list of addresses

        Args:
            address_list: A list of addresses to be tested

        Returns:
            Dict[str,bool]: The ping result mapped to the incoming address
        
        """
        pass