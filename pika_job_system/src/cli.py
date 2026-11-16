import readline
import pika
import json
import structlog
from config import settings

structlog.configure(processors=[structlog.processors.JSONRenderer()])

def send_message(channel, queue, message):
    channel.queue_declare(queue=queue, durable=True)
    channel.basic_publish(
        exchange='',
        routing_key=queue,
        body=json.dumps(message),
        properties=pika.BasicProperties(delivery_mode=2)
    )
    print(f"✓ Sent to '{queue}': {message}")

def main():
    credentials = pika.PlainCredentials(settings.rabbitmq_user, settings.rabbitmq_password)
    params = pika.ConnectionParameters(
        host=settings.rabbitmq_host,
        port=settings.rabbitmq_port,
        credentials=credentials,
        heartbeat=settings.keepalive_interval,
        blocked_connection_timeout=300
    )
    connection = pika.BlockingConnection(params)
    channel = connection.channel()
    
    current_queue = "default"
    print(f"📬 RabbitMQ CLI (queue: {current_queue})")
    print("Commands: /queue <name> | /type <text|summarize> | /quit")
    
    job_type = "text"
    
    while True:
        try:
            user_input = input(f"\n[{current_queue}] > ").strip()
            
            if not user_input:
                continue
            
            if user_input == "/quit":
                break
            
            if user_input.startswith("/queue "):
                current_queue = user_input[7:].strip()
                print(f"→ Switched to queue: {current_queue}")
                continue
            
            if user_input.startswith("/type "):
                job_type = user_input[6:].strip()
                print(f"→ Job type set to: {job_type}")
                continue
            
            message = {"type": job_type, "text": user_input}
            send_message(channel, current_queue, message)
            
        except KeyboardInterrupt:
            break
        except Exception as e:
            print(f"✗ Error: {e}")
    
    connection.close()
    print("\n👋 Goodbye!")

if __name__ == "__main__":
    main()