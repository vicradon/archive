import asyncio
import jwt
from datetime import datetime, timedelta, timezone
from centrifuge import Client, SubscriptionEventHandler, PublicationContext
from config import settings

class StreamHandler(SubscriptionEventHandler):
    def __init__(self, channel):
        self.channel = channel
        self.buffer = ""
    
    async def on_publication(self, ctx: PublicationContext):
        data = ctx.pub.data
        if data.get("type") == "stream":
            content = data.get("content", "")
            if data.get("done"):
                print(f"\n\n✅ [{self.channel}] Complete")
                self.buffer = ""
            else:
                self.buffer += content
                print(content, end="", flush=True)

async def main():
    import sys
    
    if len(sys.argv) < 2:
        print("Usage: python stream_viewer.py <channel1> [channel2] ...")
        sys.exit(1)
    
    channels = sys.argv[1:]
    
    def make_token(payload):
        return jwt.encode(
            {**payload, "exp": datetime.now(timezone.utc) + timedelta(hours=2)},
            settings.centrifuge_hmac_secret,
            algorithm="HS256",
        )
    
    async def get_client_token():
        return make_token({"sub": settings.centrifuge_generic_user_id})
    
    def get_sub_token(channel):
        async def _token():
            return make_token({"sub": settings.centrifuge_generic_user_id, "channel": channel})
        return _token
    
    client = Client(
        settings.centrifuge_url,
        get_token=get_client_token,
        use_protobuf=False,
    )
    
    await client.connect()
    print(f"🔌 Connected to Centrifugo")
    
    for channel in channels:
        handler = StreamHandler(channel)
        sub = client.new_subscription(channel, get_token=get_sub_token(channel), events=handler)
        await sub.subscribe()
        print(f"👂 Listening on channel: {channel}")
    
    print("\n📡 Streaming... (Ctrl+C to stop)\n")
    
    try:
        while True:
            await asyncio.sleep(1)
    except KeyboardInterrupt:
        await client.disconnect()
        print("\n\n👋 Disconnected")

if __name__ == "__main__":
    asyncio.run(main())