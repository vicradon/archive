import psycopg2
from psycopg2 import pool
import psycopg2.extras
import structlog

logger = structlog.get_logger()

class DatabasePool:
    def __init__(self, config):
        self.pool = psycopg2.pool.SimpleConnectionPool(
            1,
            config.db_pool_size,
            host=config.db_host,
            port=config.db_port,
            database=config.db_name,
            user=config.db_user,
            password=config.db_password
        )
        logger.info("db_pool_created", pool_size=config.db_pool_size)
        self._init_schema()
     
    def _init_schema(self):
        conn = self.get_conn()
        try:
            with conn.cursor() as cur:
                cur.execute("""
                    CREATE TABLE IF NOT EXISTS jobs (
                        id SERIAL PRIMARY KEY,
                        queue TEXT NOT NULL,
                        data JSONB NOT NULL,
                        status TEXT DEFAULT 'pending',
                        created_at TIMESTAMP DEFAULT NOW(),
                        completed_at TIMESTAMP
                    )
                """)
                conn.commit()
                logger.info("db_schema_initialized")
        finally:
            self.put_conn(conn)
    
    def get_conn(self):
        conn = self.pool.getconn()
        logger.debug("db_connection_acquired", conn_id=id(conn))
        return conn
    
    def put_conn(self, conn):
        self.pool.putconn(conn)
        logger.debug("db_connection_released", conn_id=id(conn))
    
    def register_job(self, queue, data):
        conn = self.get_conn()
        try:
            with conn.cursor() as cur:
                cur.execute(
                    "INSERT INTO jobs (queue, data) VALUES (%s, %s) RETURNING id",
                    (queue, psycopg2.extras.Json(data))
                )
                job_id = cur.fetchone()[0]
                conn.commit()
                logger.info("job_registered", job_id=job_id, queue=queue)
                return job_id
        finally:
            self.put_conn(conn)
    
    def mark_done(self, job_id):
        conn = self.get_conn()
        try:
            with conn.cursor() as cur:
                cur.execute(
                    "UPDATE jobs SET status = 'done', completed_at = NOW() WHERE id = %s",
                    (job_id,)
                )
                conn.commit()
                logger.info("job_marked_done", job_id=job_id)
        finally:
            self.put_conn(conn)