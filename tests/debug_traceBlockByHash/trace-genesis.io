// requests a trace of the genesis block by hash; must return an error since there is no parent state to replay from
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceBlockByHash","params":["0x2a6133e19662c6c61cebc63760292250368bdbc28e4c5366b5ca7525ef6cdf48"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"genesis is not traceable"}}
