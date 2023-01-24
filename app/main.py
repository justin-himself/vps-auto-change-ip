import yaml
import importlib
import schedule
import logging, coloredlogs
import os, sys
from time import sleep
from ping_client import local_ping, remote_ping

# configure logging
logger = logging.getLogger(__name__)
coloredlogs.install(level='INFO',fmt="[%(asctime)s][%(module)s][%(levelname)s] %(message)s")

# load_config 
with open("config/config.yaml", "r") as f:
    config = yaml.load(f, Loader=yaml.SafeLoader)

logging.info("Start init...")


# init ping
if config["ping"]["client_config"]["use_local"]:
    # quit if non-root
    user = os.getuid()
    if user != 0:                                                                                      
        logging.error("Re-Execute as root please.")
        sys.exit()
    ping = local_ping
else:
    ping = remote_ping

# import modules
cloud_providers = []    
for provider in config["cloud_provider"]:
    try:
        module = importlib.import_module("cloud_provider." + provider)
        instance = getattr(module, provider.capitalize())(config)
        cloud_providers.append(instance)
    except Exception as e:
        logging.exception(str(e))

logging.info("Initiated " + str(len(cloud_providers)) + 
             "/" + str(len(config['cloud_provider'])) + 
             " cloud providers.")

panel = None
try:
    module_name = next(iter(config["panel"]))
    module = importlib.import_module("panel." + module_name)
    panel = getattr(module, module_name.capitalize())(config)
    if len(config["panel"]) > 1:
        logging.warning(f"Configured mulitple panels, using {module.__name__} for now.")
except Exception as e:
    logging.exception(e)

logging.info(f"panel {module_name} was successfully initialized.")

if len(cloud_providers) == 0 or panel is None:
    logging.error("cloud_provider or panel init failed")
    sys.exit()


# main loop
def mainloop():

    logging.info("Start execution...")

    # test ping
    try:
        test_ping_result = ping(["1.1.1.1"], config)
    except Exception as e:
        logging.error("test ping throw exception:")
        logging.exception(e)
        return 
    if not test_ping_result:
        logging.error("test ping failed")
        return 
    logging.info("test ping success.")

    # fetch ip
    ip_provider_map = []
    for provider in cloud_providers:
        ips = provider.list_ip()
        try:
            ip_provider_map += [ (ip, provider) for ip in ips]
        except Exception as e:
            logging.error(f"{provider.__name__} failed with follwing exception:")
            logging.exception(e)

    if ip_provider_map == []:
        logging.warning("0 ip fetch, exiting this run.")
        return

    logging.info(f"fetched {len(ip_provider_map)} ip")

    ping_results = ping([x[0] for x in ip_provider_map], config)

    total_cnt = failed_cnt = 0
    for idx, result in enumerate(ping_results):
        if not result:
            total_cnt += 1
            try:
                oldip = ip_provider_map[idx][0]
                newip = ip_provider_map[idx][1].change_ip(oldip)
                panel.update_ip(oldip, newip)
                logging.info(f"{oldip} -> {newip}")
            except Exception as e:
                failed_cnt += 1
                logging.error(f"fail to change {oldip} into {newip}")
                logging.exception(e)
                return

    logging.info(f"Task done. {total_cnt - failed_cnt}/{total_cnt} IP changed.")
            
# schedule mainloop
for s in config["schedule"]:
    cmd = f"schedule.{s}.do(mainloop)"
    logging.info(f"applied {cmd}")
    exec(cmd)

mainloop()

while True:
    sleep(1)
    schedule.run_pending()
