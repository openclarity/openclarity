#!/usr/bin/env python3

import asyncio

from plugin.models import Config, State, Status, Metadata, Stop, Result, VMClarityData, Vulnerability
from plugin.scanner import AbstractScanner
from plugin.server import Server


class ExampleScanner(AbstractScanner):
    def __init__(self):
        self.status = Status(state=State.READY, message="Scanner ready")

    def get_status(self) -> Status:
        return self.status

    def set_status(self, status: Status):
        self.status = status

    def get_metadata(self) -> Metadata:
        return Metadata(
            name="Example scanner",
            version="v0.1.2",
        )

    async def stop(self, stop: Stop):
        # cleanup logic
        return

    async def start(self, config: Config):
        logger = Server.logger()

        # Mark scan started
        logger.info(f"Scanner is running, config={config.to_str()}")
        self.set_status(Status(state=State.RUNNING, message="Scan running"))

        # Example scanning
        await asyncio.sleep(5)
        try:
            result = Result(
                vmclarity=VMClarityData(
                    vulnerabilities=[Vulnerability(
                        vulnerability_name="vulnerability #1",
                        description="some vulnerability",
                    )],
                ),
            )
            self.export_result(result=result, output_file=config.output_file)

        except Exception as e:
            logger.error(f"Scanner failed with error {e}")
            self.set_status(Status(state=State.FAILED, message="Scan failed"))
            return

        # Mark scan done
        logger.info("Scanner finished running")
        self.set_status(Status(state=State.DONE, message="Scan done"))


if __name__ == '__main__':
    Server.run(ExampleScanner())
