import json
import structlog
from config import settings
from db import DatabasePool
from rabbitmq import RabbitMQService
from processor import JobProcessor
from centrifugo import CentrifugoPool

logger = structlog.get_logger()

class Worker:
    def __init__(self, queue, worker_id=0):
        self.queue = queue
        self.worker_id = worker_id
        self.db = DatabasePool(settings)
        self.rabbitmq = RabbitMQService(settings)
        self.processor = JobProcessor(settings.gemini_api_key)
        self.centrifugo = CentrifugoPool(settings) if settings.centrifuge_hmac_secret else None

        logger.info("worker_created", queue=queue, worker_id=worker_id)
    
    def start(self):
        self.rabbitmq.connect()
        self.rabbitmq.consume(self.queue, self._callback)
    
    def _callback(self, ch, method, properties, body):
        data = json.loads(body)
        logger.info("job_received", queue=self.queue, worker_id=self.worker_id, data=data)
        
        job_id = self.db.register_job(self.queue, data)
        
        try:
            job_type = data.get('type', 'text')
            text = data.get('text', '')
            
            if job_type == 'summarize':
                result = self.processor.process_summarize(text)
            else:
                result = self.processor.process_text(text)
            
            self.db.mark_done(job_id)
            logger.info("job_completed", job_id=job_id, result_len=len(result))
            ch.basic_ack(delivery_tag=method.delivery_tag)
        except Exception as e:
            logger.error("job_failed", job_id=job_id, error=str(e))
            ch.basic_nack(delivery_tag=method.delivery_tag, requeue=True)
