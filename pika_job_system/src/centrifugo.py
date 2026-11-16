## centrifugo.py
import jwt
from datetime import datetime, timedelta, timezone
from typing import Optional
import structlog
from centrifuge import Client, Subscription
from centrifuge.exceptions import DuplicateSubscriptionError

logger = structlog.get_logger()

class CentrifugoPool:
    def __init__(self, config):
        self.config = config
        self.user_id = config.centrifuge_generic_user_id
        self.client = None
        self.subscriptions: dict[str, Subscription] = {}
        logger.info("centrifugo_pool_created")
    
    def _make_token(self, payload: dict) -> str:
        return jwt.encode(
            {**payload, "exp": datetime.now(timezone.utc) + timedelta(hours=2)},
            self.config.centrifuge_hmac_secret,
            algorithm="HS256",
        )
    
    async def _get_client_token(self) -> str:
        return self._make_token({"sub": self.user_id})
    
    async def _get_subscription_token(self, channel: str) -> str:
        return self._make_token({"sub": self.user_id, "channel": channel})
    
    async def get_or_connect(self, channel: str) -> Subscription:
        """Get existing subscription or create new one on persistent connection."""
        if not self.client:
            self.client = Client(
                "wss://api.telex.im/centrifugo/connection/websocket",
                get_token=self._get_client_token,
                use_protobuf=False,
            )
            await self.client.connect()
            logger.info("centrifugo_connected")
        
        if channel in self.subscriptions:
            logger.debug("reusing_subscription", channel=channel)
            return self.subscriptions[channel]
        
        try:
            sub = self.client.new_subscription(
                channel,
                get_token=lambda: self._get_subscription_token(channel),
            )
            await sub.subscribe()
            self.subscriptions[channel] = sub
            logger.info("new_subscription_created", channel=channel)
            return sub
        except DuplicateSubscriptionError:
            sub = self.client.get_subscription(channel)
            self.subscriptions[channel] = sub
            return sub
    
    async def publish(self, channel: str, data: dict):
        """Publish to channel using persistent connection."""
        sub = await self.get_or_connect(channel)
        await sub.publish(data=data)
        logger.debug("published_to_channel", channel=channel)