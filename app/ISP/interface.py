from abc import ABC, abstractmethod
from typing import List

class ISPInterface(ABC):

    @abstractmethod
    def get_ip_address_list(self) -> List[str]:
        pass

    @abstractmethod
    def update_ip_address(self, old_ip:str) -> str:
        pass
    