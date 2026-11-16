import pika
import structlog

logger = structlog.get_logger()

class RabbitMQService:
    def __init__(self, config):
        self.config = config
        self.connection = None
        self.channel = None
    
    def connect(self):
        credentials = pika.PlainCredentials(
            self.config.rabbitmq_user,
            self.config.rabbitmq_password
        )
        params = pika.ConnectionParameters(
            host=self.config.rabbitmq_host,
            port=self.config.rabbitmq_port,
            credentials=credentials
        )
        self.connection = pika.BlockingConnection(params)
        self.channel = self.connection.channel()
        logger.info("rabbitmq_connected")
    
    def consume(self, queue, callback):
        self.channel.queue_declare(queue=queue, durable=True)
        self.channel.basic_consume(
            queue=queue,
            on_message_callback=callback,
            auto_ack=False
        )
        logger.info("rabbitmq_consuming", queue=queue)
        self.channel.start_consuming()