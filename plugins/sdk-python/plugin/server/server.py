import sys
import asyncio
import logging
from threading import Thread

from flask import Flask, jsonify, request, copy_current_request_context
from flask.json.provider import DefaultJSONProvider

from plugin.models import Config, State, ErrorResponse, Stop  # noqa: E501
from plugin.models.base_model import Model  # noqa: E501
from plugin.scanner import AbstractScanner  # noqa: E501
from plugin.server.config import _ServerConfig  # noqa: E501


# API_VERSION defines the current version of the Scanner Plugin SDK API.
API_VERSION = "1.0.0"

# Load config internally
_config = _ServerConfig()

# Init logger
_logger = logging.getLogger('plugin.scanner')
_logger.addHandler(logging.StreamHandler(sys.stdout))


class Server:

    @classmethod
    def run(cls, scanner: AbstractScanner):
        """
        Starts Plugin HTTP Server and uses provided scanner to respond to
        requests. Run logs data to standard output via logger. This operation
        blocks until exit. It handles graceful termination. Server listens on
        address loaded from ENV_LISTEN_ADDRESS. It exists with error code 1 on error.
        Can only be called once.

        :param scanner:
        :return:
        """
        try:
            cls.logger().info("Started HTTP server")
            server = _Server(scanner)
            server.start()
            cls.logger().info("Stopped HTTP server")
            exit(0)
        except Exception as e:
            cls.logger().error(f"HTTP server exited with error: {e}")
            exit(1)

    @classmethod
    def logger(cls):
        return _logger


class _Server:
    def __init__(self, scanner: AbstractScanner):
        self.scanner = scanner

        # Configure REST server
        self.app = Flask(__name__)
        self.app.json = _JSONProvider(self.app)
        self.register_routes()

    def start(self):
        self.app.run(host=_config.get_host(),
                     port=_config.get_port())

    def register_routes(self):
        self.app.add_url_rule('/healthz', 'get_healthz', self.get_healthz, methods=['GET'])
        self.app.add_url_rule('/metadata', 'get_metadata', self.get_metadata, methods=['GET'])
        self.app.add_url_rule('/config', 'post_config', self.post_config, methods=['POST'])
        self.app.add_url_rule('/status', 'get_status', self.get_status, methods=['GET'])
        self.app.add_url_rule('/stop', 'post_stop', self.post_stop, methods=['POST'])

    def get_healthz(self):
        status = self.scanner.get_status()
        if status.state != State.NOTREADY:
            return {}, 200

        return {}, 503

    def get_metadata(self):
        metadata = self.scanner.get_metadata()

        # Override API version so that we know on host which the actual API server being
        # used for compatibility purposes.
        metadata.api_version = API_VERSION

        return jsonify(metadata), 200

    def get_status(self):
        status = self.scanner.get_status()
        return jsonify(status), 200

    def post_config(self):
        req_data = request.get_json()
        config = Config().from_dict(req_data)

        if self.scanner.get_status().state != State.READY:
            resp = ErrorResponse(message="scanner is not in ready state")
            return jsonify(resp), 409

        @copy_current_request_context
        def start_scanner(config):
            loop = asyncio.new_event_loop()
            asyncio.set_event_loop(loop)
            try:
                loop.run_until_complete(self.scanner.start(config))
            finally:
                loop.close()

        Thread(target=start_scanner, args=(config,)).start()

        return {}, 201

    def post_stop(self):
        req_data = request.get_json()
        stop_data = Stop().from_dict(req_data)

        @copy_current_request_context
        def stop_scanner(stop_data):
            loop = asyncio.new_event_loop()
            asyncio.set_event_loop(loop)
            try:
                loop.run_until_complete(self.scanner.stop(stop_data))
            finally:
                loop.close()

        Thread(target=stop_scanner, args=(stop_data,)).start()

        return {}, 201


class _JSONProvider(DefaultJSONProvider):
    def __init__(self, app):
        super().__init__(app)

    def default(self, o):
        if isinstance(o, Model):
            return o.to_json()
        return super().default(o)
