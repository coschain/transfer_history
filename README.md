# transfer_history

## Http interface description

### 1.Get all the transfer records of an account starting from a block
--------
 URL| test env: http://qa.exchangeservice.contentos.io/api/getTransferHistory  online env: https://exchangeservice.contentos.io/api/getTransferHistory
--------- | --------|
HTTP method | POST(x-www-form-urlencoded)  
Return Format  | JSON  
Authorization |  Verification code

##### Request parameter：
--------
| parameter     | required    | Defaults  | Description |
| ------------- |-------------| -----|----
| start      |    Y     |   NO   | From which block height to get
| account    |    Y     |   NO   | Which account's transfer record needs to be obtained
| direction  |    Y     |   NO   | Transfer in or out 1:transfer out  2:transfer in
| code       |    Y     |   NO   | Authorization verification code, the request will only be processed if the verification code is correct


##### Return example:
```
 {
  "Status": 200,
  "Msg": "",
  "HeadBlockHeight": "16790",  //Latest irreversible block on the chain
  "MaxBlockHeight": "756",  // The maximum block height of this transfer record query, which can be increased by 1 when the next time you get it.
  "List": [
    {
      "OperationId": "3f516268e3f83cf7102a673ce379c6928265ccea11cf2ccc54a624bf35fbb6eb_0",  //Transaction id, remove "_0" is trx hash
      "From": "initmier", // transferred account
      "To": "account1",  //receipt account
      "Memo": "",     //memo
      "Amount": "1000000",  //transfer amount(the actual amount*1000000)
      "BlockHeight": "516"  //block height of transaction
    },
    {
      "OperationId": "a5c644c13bcc5ee578c1344a0563b34180169b3851d805b4ddea5b8b935ea723_0",
      "From": "initminer",
      "To": "account1",
      "Memo": "kjkljkj",
      "Amount": "1000000",
      "BlockHeight": "756"
    }
  ]
}

{
  "Status": 503,
  "Msg": "lack parameter direction",
  "HeadBlockHeight": "",
  "MaxBlockHeight": "",
  "List": []
}
```  
#### Error Code
| Error Code      |      Description     |
| ------------- |-------------|
| 500      |    system error     |  
| 501、502 |   query failed |  
| 503      |    lack param     |   
| 504      |   wrong parameter      |   
| 505      |    wrong verification code    |   
| 506      |    wrong transfer direction   |   

### 2.Get all the transfer records of an account in a block
--------
URL| test env: http://qa.exchangeservice.contentos.io/api/getTransferHistoryByBlock  online env: https://exchangeservice.contentos.io/api/getTransferHistoryByBlock
--------- | --------|
HTTP method | POST(x-www-form-urlencoded)  
Return Format  | JSON  
Authorization |  Verification code


##### Request parameter：
--------
| parameter     | required    | Defaults  | Description |
| ------------- |-------------| -----|----
| block      |    Y     |   NO   | Get the transfer transaction in which block
| account    |    Y     |   NO   | Which account's transfer record needs to be obtained
| direction  |    Y     |   NO   | Transfer in or out 1:transfer out  2:transfer in
| code       |    Y     |   NO   | Authorization verification code, the request will only be processed if the verification code is correct

##### Return example:

```
{
  "Status": 200,
  "Msg": "",
  "List": [
    {
      "OperationId": "a5c644c13bcc5ee578c1344a0563b34180169b3851d805b4ddea5b8b935ea723_0", //交易id,去掉"_0"即是trx hash
      "From": "initminer",  // transferred account
      "To": "account1",     //receipt account
      "Memo": "kjkljkj",   //memo
      "Amount": "1000000",  //transfer amount(the actual amount*1000000)
      "BlockHeight": "756"  //block height of the transaction
    }
  ]
}

{
  "Status": 505,
  "Msg": "verification code test1 is invalid",
  "List": []
}
```


#### Error Code
--------
| Error Code      |      Description     |
--------- | --------|
| 500      |    system error     |  
| 501、502 |   query failed |  
| 503      |    lack param     |   
| 504      |   wrong parameter      |   
| 505      |    wrong verification code    |   
| 506      |    wrong transfer direction   |  



 
