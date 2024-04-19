import os
from urllib.parse import urlparse

ENV_LOG_LEVEL = "PLUGIN_SERVER_LOG_LEVEL"
DEFAULT_LOG_LEVEL = "info"

ENV_LISTEN_ADDRESS = "PLUGIN_SERVER_LISTEN_ADDRESS"
DEFAULT_LISTEN_ADDRESS = "http://0.0.0.0:8080"


class _ServerConfig:
    def __init__(self):
        self.log_level = os.environ.get(ENV_LOG_LEVEL, DEFAULT_LOG_LEVEL)
        self.listen_address = os.environ.get(ENV_LISTEN_ADDRESS, DEFAULT_LISTEN_ADDRESS)

    def get_host(self):
        return urlparse(self.listen_address).hostname

    def get_port(self):
        return urlparse(self.listen_address).port

    def get_log_level(self):
        return self.log_level.upper()
