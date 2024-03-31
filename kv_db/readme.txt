Key-Value Store API Documentation

Overview:
This API allows clients to store, retrieve, delete, and bulk update key-value pairs. All requests except for GET must be authenticated using an API key.

Authentication:
- To authenticate, include your API key in the header of each request using the X-API-Key header.

Endpoints:

POST - Set a Key-Value Pair
- Endpoint: /
- Method: POST
- Header: X-API-Key: <your_api_key>
- Body: (Text) The value to store.
- URL Parameter: The key is included in the URL, for example, /path/to/key.
- Description: Stores or updates the value of the specified key. If the key already exists, its value is updated.
- Response: A success message if the key-value pair is successfully stored.
- Example Request: curl -X POST http://localhost:8080/path/to/key -H "X-API-Key: <your_api_key>" -d "value"

GET - Retrieve a Value by Key
- Endpoint: /{key}
- Method: GET
- Description: Retrieves the value associated with the specified key. 
- Response: The value of the key if it exists; otherwise, a "Key not found" error.
- Example Request: curl -X GET http://localhost:8080/path/to/key 

DELETE - Delete a Key-Value Pair
- Endpoint: /{key}
- Method: DELETE
- Header: X-API-Key: <your_api_key>
- Description: Deletes the specified key and its associated value from the store.
- Response: A success message if the key is successfully deleted.
- Example Request: curl -X DELETE http://localhost:8080/path/to/key -H "X-API-Key: <your_api_key>"

PUT - Bulk Update Key-Value Pairs
- Endpoint: /
- Method: PUT
- Header: X-API-Key: <your_api_key>
- Body: (Text) New contents for the entire store, formatted as key,value pairs separated by newlines.
- Description: Replaces the entire store with the new key-value pairs provided in the request body. This is a bulk update operation and will overwrite all existing data.
- Response: A success message if the store is successfully replaced.
- Example Request: curl -X PUT http://localhost:8080/ -H "X-API-Key: <your_api_key>" -d "key1,value1\nkey2,value2"

Error Handling:
- If an API key is missing or invalid, the server will respond with a 401 Unauthorized status code.
- If a requested key does not exist during a GET operation, the server will respond with a 404 Not Found status code.
- For unsupported methods, the server will respond with a 405 Method Not Allowed status code.

