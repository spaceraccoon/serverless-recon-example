import json

import psycopg2
from psycopg2 import sql
import requests


def lambda_handler(event, context):
    """Sample pure Lambda function

    Parameters
    ----------
    event: dict, required
        API Gateway Lambda Proxy Input Format

        Event doc: https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-lambda-proxy-integrations.html#api-gateway-simple-proxy-for-lambda-input-format

    context: object, required
        Lambda Context runtime methods and attributes

        Context doc: https://docs.aws.amazon.com/lambda/latest/dg/python-context-object.html

    Returns
    ------
    API Gateway Lambda Proxy Output Format: dict

        Return doc: https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-lambda-proxy-integrations.html
    """

    try:
        ip = requests.get("http://checkip.amazonaws.com/")
    except requests.RequestException as e:
        # Send some context about this error to Lambda Logs
        print(e)

        raise e

    if "queryStringParameters" not in event or "domain" not in event["queryStringParameters"]:
        return {
            "statusCode": 400,
            "body": json.dumps({
                "error": "Requires domain parameter."
            }),
        }

    try:
        connection = psycopg2.connect(user="guest", host="crt.sh", port="5432", database="certwatch")
        connection.set_session(readonly=True, autocommit=True)

        cursor = connection.cursor()

        cursor.execute(sql.SQL("SELECT DISTINCT NAME_VALUE FROM certificate_identity ci WHERE ci.NAME_TYPE = 'dNSName' AND reverse(lower(ci.NAME_VALUE)) LIKE reverse(lower('%.{}'));".format(event["queryStringParameters"]["domain"])))
        records = [r[0] for r in cursor.fetchall()]
    except (Exception, psycopg2.Error) as e :
        print ("Error while connecting to PostgreSQL", e)
    finally:
        #closing database connection.
        if(connection):
            cursor.close()
            connection.close()

    return {
        "statusCode": 200,
        "body": json.dumps({
            "domain": event["queryStringParameters"],
            "subdomains": records,
        }),
    }
