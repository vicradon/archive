from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    db_host: str = "localhost"
    db_port: int = 5432
    db_name: str = "jobs"
    db_user: str = "postgres"
    db_password: str = "postgres"
    db_pool_size: int = 5
    
    rabbitmq_host: str = "localhost"
    rabbitmq_port: int = 5672
    rabbitmq_user: str = "guest"
    rabbitmq_password: str = "guest"

    keepalive_interval: int = 10
    
    gemini_api_key: str

    centrifuge_generic_user_id: str = "system"
    centrifuge_hmac_secret: str = ""
    centrifuge_url: str
    centrifuge_websocket_url: str
    centrifuge_api_key: str
    
    class Config:
        env_file = ".env"

settings = Settings()
