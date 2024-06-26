from pymongo import MongoClient
import json


for i in [1, 10, 100, 1000]:
    client = MongoClient('mongodb://localhost:27017/')
    filter = {
        'message_id': i
    }
    sort = list({
        'message_id': 1,
        'offset': 1
    }.items())

    result = client['reassembleudp']['fragments'].find(
        filter=filter,
        sort=sort
    )

    with open(f'{i:04}.json', 'w') as f:
        f.write(json.dumps(list(result), default=str))
