import utils
import yaml
import time
import schedule, threading
import logging, coloredlogs

def main():

    # load config
    with open("config/config.yaml", "r") as f:
        config = yaml.load(f, Loader=yaml.SafeLoader)

    # configure logging
    logger = logging.getLogger(__name__)
    coloredlogs.install(level='INFO',fmt="[%(asctime)s][%(module)s][%(levelname)s] %(message)s")

    # schedule tasks
    for task in config["Tasks"]:
        for s in task["Schedule"]:
            cmd = f"schedule.{s}.do(do_task, task)"
            logging.info(f"applied {cmd}")
            exec(cmd)

    # run each task immediately once
    for task in config["Tasks"]:
        do_task(task)

    # main loop 
    while True:
        time.sleep(1)
        schedule.run_pending()


def do_task(task_config):

    def _():
        
        # load modules
        pinger = utils.load_module("Pinger", task_config["Pinger"]["Type"])
        isp = utils.load_module("ISP", task_config["ISP"]["Type"])
        if task_config["Panel"]["Type"] is not "None":
            panel = utils.load_module("Panel", task_config["Panel"]["Type"])
        if task_config["CDN"]["Type"] is not "None":
            cdn = utils.load_module("ISP", task_config["CDN"]["Type"])
            
        # Fetch address list from panel / cdn
        addr_list = panel.get_node_address_list()
        arr = []

        # Ping each addresss
        ping_result_list = pinger.ping(addr_list)

        # Change IP at ISP 
        addr_change_list = []
        for oldip, result in ping_result_list:
            if result == False:
                newip = isp.update_ip_address(oldip)    
                addr_change_list.append((oldip, newip))

        # Feedback to Panel
        for addr_change_

        # Feedback to CDN

    
    task_thread = threading.Thread(target=_)
    task_thread.start()


if __name__ == "__main__":
    exit(main())