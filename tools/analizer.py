import re
import json

types = set()
events = set()

with open('logs/logs.log', 'r') as f:

    for row in f.readlines():
        # print(row)
        try:
            find_string = re.findall(r'\[.*\]', row)[0]
            parsed_string = json.loads(find_string)
            if parsed_string[0] == 'event':
                events.add(parsed_string[1]['event_type'])
            else:
                types.add(parsed_string[0])

        except IndexError:
            pass
    for t in types:
        print(f'├── {t}')
    # print(types)
    print('├── events')
    for e in events:
        print(f'   └── {e}')

