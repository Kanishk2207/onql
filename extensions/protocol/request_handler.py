import json
class RequestHandler:
    def __init__(self, schema, sdk):
        self.schema = schema
        self.sdk = sdk
      
    async def handle_request(self, msg):
        print(f"Received request: {msg}")
        res = await self.executeCommand(msg['payload'])
        await self.sdk.response(msg, res)

    async def executeCommand(self,command):
        command = json.loads(command)
        if command[0] == "desc":
            protocols = await self.schema.protocols()
            protocols = json.loads(protocols)
            try:
                for key in command[1:]:
                    protocols = protocols[key]
                    # print(f"Available protocols: {protocols}")
            except KeyError:
                return "error Protocol not found on keyword %s" % command
            return json.dumps(protocols)

        elif command[0] == "set":
            return await self.schema.set(command[1], command[2])
        elif command[0] == "drop":
            return await self.schema.delete(command[1])
       