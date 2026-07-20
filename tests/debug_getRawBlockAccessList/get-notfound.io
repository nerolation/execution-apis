// requests the block access list of a non-existent block
>> {"jsonrpc":"2.0","id":1,"method":"debug_getRawBlockAccessList","params":["0x2710"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32001,"message":"Resource not found"}}
