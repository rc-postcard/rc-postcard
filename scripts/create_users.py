#!/usr/bin/python3

from ast import Add
import os
import lob
import psycopg2
import requests


def main():
    lob.api_key = os.getenv("LOB_API_TEST_KEY")
    RC_ACCESS_TOKEN = os.getenv("RC_ACCESS_TOKEN")

    connection_string = os.getenv("PG_DATABASE_URL")
    conn = psycopg2.connect(connection_string)

    r = requests.get('http://recurse.com/api/v1/profiles?access_token=' + RC_ACCESS_TOKEN + '&scope=current&limit=50')
    for user in r.json():
        print(user["id"], user["email"], user["name"])

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
                'test': "test_1"
            }
        )

        address_id = address["id"]

        cursor = conn.cursor()
        cursor.execute(
            "INSERT INTO user_info (recurse_id, lob_address_id, user_name, user_email) VALUES (%s, %s, %s, %s) ON CONFLICT (recurse_id) DO UPDATE SET lob_address_id = excluded.lob_address_id;",
            (user["id"], address_id, user["name"], user["email"]))
        conn.commit()
        cursor.close()
    conn.close()

if __name__ == '__main__':
    main()