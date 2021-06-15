import json
import configparser
import requests

f = open('../frontend/src/i18n/en.json',)
data = json.load(f)
f.close()

def flatten(data):
    flattened = {}

    for key, value in data.items():
        if isinstance(value, dict):
            temp = flatten(value)
            for k, v in temp.items():
                flattened[key + '.' + k] = v
        else:
            flattened[key] = value

    return flattened


flattened = flatten(data)

config = configparser.ConfigParser()
config.read('config')
main = config['main']
url = main['host'] + '/api/v2/messages?key=' + main['key']
headers = {'accept': 'application/json'}

for key, value in flattened.items():
    payload = {
        'brand': 'filebrowser',
        'body': value,
        'slug': key,
    }
    response = requests.post(
        url,
        data=payload,
        headers=headers
    )
    print(response.status_code, response.text)
