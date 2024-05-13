import json
from abc import ABC, abstractmethod

from plugin.models import Config, Status, Metadata, Stop, Result  # noqa: E501


class AbstractScanner(ABC):

    @abstractmethod
    def get_metadata(self) -> Metadata:
        pass

    @abstractmethod
    def get_status(self) -> Status:
        pass

    @abstractmethod
    def set_status(self, status: Status):
        pass

    @abstractmethod
    async def start(self, config: Config):
        pass

    @abstractmethod
    async def stop(self, stop: Stop):
        pass

    @staticmethod
    def export_result(result: Result, output_file: str):
        with open(output_file, 'w') as f:
            json.dump(result.to_json(), f)
