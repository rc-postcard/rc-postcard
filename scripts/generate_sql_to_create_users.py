#!/usr/bin/python3

from ast import Add
import os
import psycopg2
import requests


def main():
    RC_ACCESS_TOKEN = os.getenv("RC_ACCESS_TOKEN")

    r = requests.get('http://recurse.com/api/v1/profiles?access_token=' +
                     RC_ACCESS_TOKEN + '&scope=current&limit=50')
    for user in r.json():
        print("INSERT INTO user_info (recurse_id, user_name, user_email) VALUES ({0}, \'{1}\', \'{2}\') ON CONFLICT (recurse_id) DO NOTHING;".format(
            user["id"], user["name"], user["email"]))

    print("INSERT INTO user_info (recurse_id, user_name, user_email) VALUES ({0}, \'{1}\', \'{2}\') ON CONFLICT (recurse_id) DO NOTHING;".format(
        0, "Recurse Center", "admissions@recurse.com"))


if __name__ == '__main__':
    main()
