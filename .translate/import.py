import json
import configparser
import json
import requests

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

def deflatten(data):
    deflattened = {}

    for key, value in data.items():
        parts = key.split('.')
        temp = deflattened
        for idx, part in enumerate(parts):
            if part not in temp:
                if idx == (len(parts) - 1):
                    temp[part] = value
                else:
                    temp[part] = {}
            temp = temp[part]

    return deflattened

def log_missing_translations(current, new):
    for slug, value in current.items():
        if slug not in new:
            print("removed source translation -> %s: \"%s\"" % (slug, value))

config = configparser.ConfigParser()
config.read('config')
main = config['main']

key_query = '?key=' + main['key']
headers = {'accept': 'application/json'}

response = requests.get(main['host'] + '/api/v2/brands/' + main['brand'] + '/languages' + key_query, headers=headers)
if response.status_code != 200:
    raise Exception('Could not fetch brand languages')

languages = json.loads(response.text)

for language in languages:
    print(language['code'])
    response = requests.get(main['host'] + '/api/v3/brands/' + main['brand'] + '/languages/' + language['code'] + '/dictionary' + key_query + '&fallback_locale=en_GB')
    if response.status_code != 200:
        print('could not fetch translations for messages: %s' % language['code'])
        continue

    decoded = json.loads(response.text)
    parsed = deflatten(decoded)

    # log existing translations that are not present in the freshly parsed data
    if language['code'] == 'en_GB':
        f = open('../frontend/src/i18n/' + language['code'] + '.json')
        data = json.load(f)
        f.close()
        current_translations = flatten(data)
        log_missing_translations(current_translations, decoded)

    fd = open('../frontend/src/i18n/%s.json' % language['code'], 'w')
    fd.write(json.dumps(parsed, indent=2, ensure_ascii=False))
    fd.close()
