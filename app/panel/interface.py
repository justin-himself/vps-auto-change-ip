from abc import ABC, abstractmethod
from typing import List, Tuple

class PanelInterface(ABC):
    
    @abstractmethod
    def get_node_address_list(self) -> List[Tuple[str,str]]:
        pass

    @abstractmethod
    def update_node_address(self, old_addr:str, new_addr:str):
        pass
