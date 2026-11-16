import sys
import multiprocessing
import structlog
import yaml
from typing import Dict, Optional
from pydantic import BaseModel
from pydantic_settings import BaseSettings
from worker import Worker

# --- Logging setup ---
structlog.configure(
    processors=[
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.add_log_level,
        structlog.processors.JSONRenderer()
    ]
)
logger = structlog.get_logger()


class WorkerConfig(BaseModel):
    queue_name: str
    worker_count: Optional[int] = 1


class WorkerQueue(BaseSettings):
    workers: Dict[str, WorkerConfig]

    @classmethod
    def from_yaml(cls, path: str) -> "WorkerQueue":
        with open(path, "r") as f:
            raw = yaml.safe_load(f)
        return cls(workers=raw)


def start_worker(queue: str, worker_id: int):
    Worker(queue, worker_id).start()


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python main.py <config.yaml>")
        sys.exit(1)

    config_path = sys.argv[1]
    worker_queue = WorkerQueue.from_yaml(config_path)

    processes = []

    for name, settings in worker_queue.workers.items():
        queue = settings.queue_name
        worker_count = settings.worker_count

        logger.info("starting_workers", name=name, queue=queue, worker_count=worker_count)

        for i in range(worker_count):
            p = multiprocessing.Process(target=start_worker, args=(queue, i))
            p.start()
            processes.append(p)

    for p in processes:
        p.join()
