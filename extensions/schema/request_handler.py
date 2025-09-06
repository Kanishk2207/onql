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
        if command[0] == "databases":
            return await self.schema.databases()
        elif command[0] == "tables":
            return await self.schema.tables(command[1])
        elif command[0] == "desc" and len(command) == 1:
            return await self.schema.fullSchema()
        elif command[0] == "desc":
            schema = await self.schema.fullSchema()
            schema = json.loads(schema)
            try:
                for key in command[1:]:
                    schema = schema[key]
            except Exception as e:
                return "error Schema not found on keyword %s" % key
            return json.dumps(schema)
           
        elif command[0] == "create":
            if command[1] == "db":
                return await self.schema.createDatabase(command[2])
            elif command[1] == "table":
                return await self.schema.createTable(command[2],command[3], command[4])
        elif command[0] == "rename":
            if command[1] == "db":
                return await self.schema.RenameDatabase(command[2],command[3])
            elif command[1] == "table":
                return await self.schema.RenameTable(command[2],command[3], command[4])
        elif command[0] == "alter":
            return await self.schema.AlterTable(command[1],command[2],command[3])
        elif command[0] == "drop":
            if command[1] == "db":
                return await self.schema.DropDatabase(command[2])
            elif command[1] == "table":
                return await self.schema.DropTable(command[2],command[3])
        elif command[0] == "set":
            return await self.set(command[1])


    async def set(self,data):
        result = []
        for key, value in data.items():
           res = await self.schema.createDatabase(key)
           result.append(res)
           for table_name, table_data in value.items():
            #    set or alter columns here
               for columnName in table_data:
                   table_data[columnName]["storage"] = "disk"
                # check if table exists then alter table
               if table_name in json.loads(await self.schema.tables(key)):
                    tableSchema = json.loads(await self.executeCommand(json.dumps(["desc",key,table_name])))
                #    check if column already exists
                    for column in table_data:
                       

                        alters = {}

                        if column not in tableSchema:
                            # set schema
                            alters["addColumn"] = {
                                "name" : column,
                                **table_data.get(column)
                            }
                        else:
                            alters["changeBlank"] = {
                                "name" : column,
                                "newBlank": table_data[column]["blank"]
                            }
                            alters["changeDefault"] = {
                                "name": column,
                                "newDefaut": table_data[column]["default"]
                            }
                        
                        res = await self.schema.AlterTable(key,table_name,alters)
                        result.append(res)


                    # check drop columns
                    for column in tableSchema:
                            alters = {}
                            if column not in table_data:
                                alters["dropColumn"] = {
                                    "name": column
                                }
                                res = await self.schema.AlterTable(key,table_name,alters)
                                result.append(res)
               else:
                    res = await self.schema.createTable(key, table_name, table_data)
                    result.append(res)
        return json.dumps(result)
               


# {"this":{"id":},{:this}}[1,2,3]

