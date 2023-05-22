from abc import ABC, abstractmethod
from typing import Dict

class CDNInterface(ABC):

    @abstractmethod
    def get_record_list(self) -> Dict[str, str]:
        pass

    @abstractmethod
    def update_record(self, domain:str, new_ip_addr:str) -> None:
        pass

