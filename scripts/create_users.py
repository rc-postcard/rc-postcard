#!/usr/bin/python3

from ast import Add
import os
import lob
import psycopg2
import requests


def main():
    lob.api_key = os.getenv("LOB_API_TEST_KEY")
    RC_ACCESS_TOKEN = os.getenv("RC_ACCESS_TOKEN")

    r = requests.get('http://recurse.com/api/v1/profiles?access_token=' + RC_ACCESS_TOKEN + '&scope=current&limit=50')
    for user in r.json():
        
        address = lob.Address.create(
            description=user["name"],
            name=user["name"],
            email=user["email"],
            address_line1='397 Bridge Street',
            address_city='Brooklyn',
            address_state='NY',
            address_zip='11201',
            address_country='US',
            metadata={
                'rc_id': user["id"],
                'test': "test_2"
            }
        )

        address_id = address["id"]

        print("INSERT INTO user_info (recurse_id, lob_address_id, user_name, user_email) VALUES ({0}, \'{1}\', \'{2}\', \'{3}\') ON CONFLICT (recurse_id) DO UPDATE SET lob_address_id = excluded.lob_address_id;".format(
            user["id"], address_id, user["name"], user["email"]
        ))

if __name__ == '__main__':
    main()