import argparse
import json
import logging
import operator
import random
from builtins import Exception

import docker
import requests
from websocket import create_connection

logger = logging.getLogger(__name__)

MASTER_TARGETS = ["customer"]
DEFAULT_ATTRIBUTES = {"customer": "ComponentTest", "cluster": "local"}
DEFAULT_MESSAGE = "ComponentTest special message"


class Notification(object):
    def __init__(self, **kwargs):
        self.target = kwargs["target"] if "target" in kwargs else DEFAULT_ATTRIBUTES
        self.notification = kwargs["notification"] if "notification" in kwargs else DEFAULT_MESSAGE

    def __repr__(self):
        return json.dumps({i: j for i, j in self.__dict__.items() if not callable(getattr(self, i))})

    def __eq__(self, other):
        return self.__repr__() == other.__repr__()


class ComponentTest(object):
    def __init__(self, image: str):
        self.docker_client = docker.DockerClient()
        self.image: str = image
        self.network = None
        self.master = None
        self.edge: list = []
        self.client: list = []
        self.notification: list = [Notification()]

    def run_network(self):
        self.network = self.docker_client.networks.create(name=self.random_name("network"), attachable=True)

    def run_master_container(self):
        self.master = self.run_container(name=self.random_name("master"))

    def run_edge_container(self):
        master_host = "ws://{}:8001/waitfornotification".format(self.master.name)
        environment = ["MASTER_HOST={}".format(master_host), "MASTER_ATTRIBUTES={}".format(";".join(MASTER_TARGETS))]
        self.edge.append(self.run_container(name=self.random_name("edge"), environment=environment))

    def run_container(self, name: str, environment: list = [], ports: dict = {}):
        print("running container: {}".format(name))
        return self.docker_client.containers.run(image=self.image, detach=True, name=name, environment=environment,
                                                 ports=ports, network=self.network.name)

    def run_client(self, edge, notification: Notification):
        edge_ip = self.get_container_ip(container=edge, network=self.network)
        url = "ws://{}:8001/waitfornotification?{}".format(edge_ip, self.convert_dict_to_url(notification.target))
        self.client.append(self.connect_websocket(url))

    def push_notification(self, notification: Notification):
        print("push_notification")
        master_ip = self.get_container_ip(container=self.master, network=self.network)
        url = "http://{}:8002/sendnotification?{}".format(master_ip, self.convert_dict_to_url(notification.target))
        print("post. url: {}, data: {}".format(url, repr(notification)))
        r = requests.post(url=url, data=repr(notification))
        assert r.status_code == 200, "error in posting notification, status code: {}, message: {}".format(r.status_code,
                                                                                                          r.text)
        print("post successfully")

    def __del__(self):
        for i in self.client:
            self.close_websocket(i)
        for i in self.edge:
            self.remove_container(i)
        if self.master:
            self.remove_container(self.master)
        if self.network:
            self.remove_network(self.network)

    @staticmethod
    def test_received_notification(notf1, notf2, op=operator.eq):
        print("testing received notification")
        assert op(notf1, notf2), "the notifications are not the same"

    @staticmethod
    def receive_notification(client):
        print("receive_notification")
        data = client.recv()
        return Notification(**json.loads(data))

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
        self.run_client(self.edge[0], self.notification[0])

        # test
        self.push_notification(self.notification[0])  # master
        received = self.receive_notification(self.client[0])  # client
        self.test_received_notification(received, self.notification[0])


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
