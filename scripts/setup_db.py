#!/usr/bin/python3

from ast import Add
import os
import lob
import psycopg2


def main():
    lob.api_key = os.getenv("LOB_API_TEST_KEY")

    # create recurse center entry in your lob account
    address = lob.Address.create(
        description='Recurse Center',
        name='Recurse Center',
        email='admissions@recurse.com',
        address_line1='397 Bridge Street',
        address_city='Brooklyn',
        address_state='NY',
        address_zip='11201',
        address_country='US',
        metadata={
            'rc_id': '0'
        }
    )

    address_id = address["id"]
    print(address_id)

    connection_string = os.getenv("PG_DATABASE_URL")
    conn = psycopg2.connect(connection_string)

    cursor = conn.cursor()
    cursor.execute(
        "INSERT INTO user_info (recurse_id, lob_address_id, user_name, user_email) VALUES (%s, %s, %s, %s) ON CONFLICT (recurse_id) DO UPDATE SET lob_address_id = excluded.lob_address_id;",
        (0, address_id, 'Recurse Center', 'admissions@recurse.com'))
    conn.commit()
    cursor.close()
    conn.close()

if __name__ == '__main__':
    main()
