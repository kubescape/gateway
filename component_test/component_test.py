import argparse
import json
import logging
import random
import time
from builtins import Exception

import docker
import requests
from websocket import create_connection

logger = logging.getLogger(__name__)

MASTER_TARGETS = ["customer"]
DEFAULT_ATTRIBUTES = {"customer": "ComponentTest", "cluster": "local"}


class Notification(object):
    def __init__(self, **kwargs):
        self.target = kwargs if kwargs else DEFAULT_ATTRIBUTES
        self.notification = "ComponentTest special message"

    def __cmp__(self, other):
        return self.notification == other.notification

    def json(self):
        return json.dumps({i: j for i, j in self.__dict__.items() if not callable(getattr(self, i))})


class ComponentTest(object):
    def __init__(self, image: str):
        self.docker_client = docker.DockerClient()
        self.image: str = image
        self.network = None
        self.master = None
        self.edge = None
        self.client = None
        self.notification = Notification()

    def run_network(self):
        self.network = self.docker_client.networks.create(name=self.random_name("network"), attachable=True)

    def run_master_container(self):
        self.master = self.run_container(name=self.random_name("master"))  ## ports={8001: 8001, 8002: 8002}

    def run_edge_container(self):
        master_host = "ws://{}:8002/waitfornotification".format(self.master.name)
        environment = ["MASTER_HOST={}".format(master_host), "MASTER_TARGETS={}".format(";".join(MASTER_TARGETS))]
        self.edge = self.run_container(name=self.random_name("edge"), environment=environment)  # ports={8002: 8002}

    def run_container(self, name: str, environment: list = [], ports: dict = {}):
        print("running container: {}".format(name))
        return self.docker_client.containers.run(image=self.image, detach=True, name=name, environment=environment,
                                                 ports=ports, network=self.network.name)

    def run_client(self):
        edge_ip = self.get_container_ip(container=self.edge, network=self.network)
        url = "ws://{}:8002/waitfornotification?{}".format(edge_ip, self.convert_dict_to_url(self.notification.target))
        self.client = self.connect_websocket(url)

    def receive_notification(self):
        print("receive_notification")
        data = self.client.recv()
        json.loads(data)

    def push_notification(self):
        print("push_notification")
        master_ip = self.get_container_ip(container=self.master, network=self.network)
        url = "http://{}:8001/sendnotification?{}".format(master_ip, self.convert_dict_to_url(self.notification.target))
        requests.post(url=url, data=self.notification.json())

    def __del__(self):
        if self.client:
            self.close_websocket(self.client)
        if self.edge:
            self.remove_container(self.edge)
        if self.master:
            self.remove_container(self.master)
        if self.network:
            self.remove_network(self.network)

    @staticmethod
    def connect_websocket(url):
        return create_connection(url=url)

    @staticmethod
    def random_name(name: str):
        return "{}_{}".format(name, random.randint(0, 1000))

    @staticmethod
    def remove_container(container):
        try:
            container.stop()
            container.remove(v=True)
        except Exception as e:
            logger.error(e)

    @staticmethod
    def remove_network(network):
        try:
            network.remove()
        except Exception as e:
            logger.error(e)

    @staticmethod
    def close_websocket(ws):
        try:
            ws.close()
        except Exception as e:
            logger.error(e)

    @staticmethod
    def inspect(container):
        return docker.APIClient().inspect_container(container.id)

    @staticmethod
    def convert_dict_to_url(d: dict):
        return "&".join(["{}={}".format(i, j) for i, j in d.items()])

    @staticmethod
    def get_container_ip(container, network):
        return ComponentTest.inspect(container=container)['NetworkSettings']['Networks'][network.name]['IPAddress']

    def run(self):
        """
        setup:
        1. run network
        2. run master
        3. run edge
        4. run client (websocket to edge)

        test:
        1. send notification
        2. receive notification from websocket

        :return:
        """
        # setup
        self.run_network()
        self.run_master_container()
        self.run_edge_container()
        self.run_client()

        # test
        self.push_notification()  # master
        self.receive_notification()  # client
        # self.test_received_notification
        time.sleep(10)


def input_parser():
    parser = argparse.ArgumentParser("Run notification server component test")

    parser.add_argument("--image", action="store", dest="image", required=True,
                        help="notification server image")

    return parser.parse_args()


if __name__ == "__main__":
    args = input_parser()
    # logger.setLevel(logging.DEBUG)

    ct = ComponentTest(image=args.image)
    try:
        ct.run()
    except Exception as e:
        print(e)
    finally:
        print("cleaning up")
        del ct
