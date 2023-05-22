import importlib
import os
import inspect
from typing import Dict, Any, Type

def load_module(
        package_name:str, 
        module_name:str, 
        class_name : str = None,
        params : Dict[str, Any] = {},
    ) -> Type:
    """
    fuzzy search and load module, returns an instance
    if class_name is not provided, it is assumed to be equal to module_name
    """

    if class_name is None:
        class_name = module_name
    package_name = package_name.lower()
    module_name = module_name.lower()
    class_name = class_name.lower()
    

    # Find package
    real_package_name = None
    for root, folders, files in os.walk("."):
        for folder in folders:
            if folder.lower() == package_name:
                real_package_name = folder
    if real_package_name is None:
        raise Exception(f"Can't find specified package {package_name}")

    # Find module
    real_module_name = None
    for root, folders, files in os.walk(f"./{real_package_name}"):
        for file in files:
            if file.strip(".py").lower() == module_name:
                real_module_name = file.strip(".py")
    if real_module_name is None:
        raise Exception(f"Can't find module {module_name} in {real_package_name}")
    
    # Find class object
    class_obj = None
    module = importlib.import_module(real_package_name + "." + real_module_name)
    for name, obj in inspect.getmembers(module):
        if inspect.isclass(obj) and name.lower() == class_name:
            class_obj = obj
    if class_obj is None:
        raise Exception(f"Can't find {class_name} in {real_module_name}.")
    
    # Instanitiate the class object
    return class_obj(**params)