import json

class Schema:
    def __init__(self,sdk):
        self.sdk = sdk

    async def databases(self):
        payload = {
            "function":"GetDatabases",
            "args": []
        }
        response = await self.sdk.request("database", json.dumps(payload))
        print(response)
        return response["payload"]

    async def tables(self,db):
        payload = {
            "function":"GetTables",
            "args": [db]
        }
        response = await self.sdk.request("database", json.dumps(payload))
        print(response)
        return response["payload"]

    async def schema(self,db,table):
        payload = {
            "function":"GetTableSchema",
            "args": [db, table]
        }
        response = await self.sdk.request("database", json.dumps(payload))
        print(response)
        return response["payload"]
    
    async def fullSchema(self):
        payload = {
            "function":"GetFullSchema",
            "args": []
        }
        response = await self.sdk.request("database", json.dumps(payload))
        print(response)
        return response["payload"]

    async def createDatabase(self, db_name):
        payload = {
            "function":"CreateDatabase",
            "args": [db_name]
        }
        response = await self.sdk.request("database", json.dumps(payload))
        print(response)
        return response["payload"]

    async def createTable(self, db_name, table_name, schema):
        payload = {
            "function":"CreateTable",
            "args": [db_name, table_name, schema]
        }
        response = await self.sdk.request("database", json.dumps(payload))
        print(response)
        return response["payload"]

    async def RenameDatabase(self,oldName,newName):
        payload = {
            "function":"RenameDatabase",
            "args": [oldName,newName]
        }
        response = await self.sdk.request("database", json.dumps(payload))
        print(response)
        return response["payload"]
    
    async def RenameTable(self,db,oldName,newName):
        payload = {
            "function":"RenameTable",
            "args": [db,oldName,newName]
        }
        response = await self.sdk.request("database", json.dumps(payload))
        print(response)
        return response["payload"]

    async def AlterTable(self, db, table, alters):
        payload = {
            "function":"AlterTable",
            "args": [db, table, alters]
        }
        response = await self.sdk.request("database", json.dumps(payload))
        print(response)
        return response["payload"]

    async def DropTable(self, db, table):
        payload = {
            "function":"DeleteTable",
            "args": [db, table]
        }
        response = await self.sdk.request("database", json.dumps(payload))
        print(response)
        return response["payload"]
    
    async def DropDatabase(self, db):
        payload = {
            "function":"DeleteDatabase",
            "args": [db]
        }
        response = await self.sdk.request("database", json.dumps(payload))
        print(response)
        return response["payload"]
    
    async def RefereshIndexes(self):
        payload = {
            "function":"RefereshIndexes",
            "args": []
        }
        response = await self.sdk.request("database", json.dumps(payload))
        print(response)
        return response["payload"]

class Table:
    def __init__(self):
        self.schema = {}

    def add(self, name, type = "string", storage = "disk", blank = "no", default=None):
        self.schema[name] = {
            "type": type,
            "storage": storage,
            "blank": blank,
            "default": default
        }

    def get(self):
        return self.schema
    
    def remove(self, name):
        if name in self.schema:
            del self.schema[name]

    def clear(self):
        self.schema = {}

