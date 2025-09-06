import json

class Protocol:
    def __init__(self,sdk):
        self.sdk = sdk

    async def protocols(self):
        payload = {
            "function":"GetAllProtocols",
            "args": []
        }
        response = await self.sdk.request("database", json.dumps(payload))
        print(response)
        return response["payload"]

    # async def protocolByPassword

    async def delete(self,password):
        payload = {
            "function":"DeleteProtocolByPassword",
            "args": [password]
        }
        response = await self.sdk.request("database", json.dumps(payload))
        print(response)
        return response["payload"]

    async def set(self, password, data):
        payload = {
            "function": "SetProtocol",
            "args": [password, data]
        }
        response = await self.sdk.request("database", json.dumps(payload))
        print(response)
        return response["payload"]
    
    