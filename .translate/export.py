import json
import configparser
import requests

config = configparser.ConfigParser()
config.read('config')
main = config['main']

source_locale = "en_GB"

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

key_query = '?key=' + main['key']
response = requests.get(main['host'] + '/api/v3/brands/' + main['brand'] + '/languages/' + source_locale + '/dictionary' + key_query)
if response.status_code != 200:
    print('could not fetch existing messages')
    exit()

messages = json.loads(response.text)

f = open('../frontend/src/i18n/' + source_locale + '.json')
data = json.load(f)
f.close()

flattened = flatten(data)

url = main['host'] + '/api/v2/messages' + key_query
headers = {'accept': 'application/json'}

for key, value in flattened.items():
    if key in messages:
        continue

    payload = {
        'brand': main['brand'],
        'body': value,
        'slug': key,
    }
    response = requests.post(
        url,
        data=payload,
        headers=headers
    )
    print(response.status_code, response.text)
