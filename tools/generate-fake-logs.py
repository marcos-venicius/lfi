#!/usr/bin/env python3
import random
import datetime
from faker import Faker

fake = Faker()


def fake_ip():
    return fake.ipv4_private()


def fake_timestamp():
    now = datetime.datetime(2025, 3, 28, 14, 56, 53)
    delta = datetime.timedelta(seconds=random.randint(-3600*24*7, 3600*24*7))
    future_time = now + delta
    return future_time.strftime("[%d/%b/%Y:%H:%M:%S +0000]")


def fake_http_method():
    methods = ["GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"]
    return random.choice(methods)

def fake_path():
    paths = [
        "/test/fake/{}/view",
        "/cart/find/here",
        "/foo/bar",
        "/shop/buy/products",
        "/ship/product/{}",
        "/dashboard/user/{}/info",
        "/user/data",
        "/cart/list",
        "/invoices",
        "/users/profile"
    ]
    base_path = random.choice(paths)
    if "{}" in base_path:
        return base_path.format(fake.uuid4().replace('-', ''))
    else:
        return base_path


def fake_http_version():
    versions = ["HTTP/1.1", "HTTP/2.0"]
    return random.choice(versions)


def fake_status_code():
    status_codes = [200, 204, 304, 400, 401, 403, 404, 500, 502, 503]
    return random.choice(status_codes)


def fake_response_size(status_code):
    if status_code in [204, 304]:
        return 0
    elif status_code >= 400:
        return random.randint(0, 100)
    else:
        return random.randint(0, 10000)


def fake_referer():
    referers = [
        "https://www.youtube.com/",
        "https://www.google.com/",
        "https://app.example.com/",
        "-",
        ""
    ]
    return random.choice(referers)


def fake_user_agent():
    return fake.user_agent()


def generate_log_line():
    ip = fake_ip()
    timestamp = fake_timestamp()
    method = fake_http_method()
    path = fake_path()
    http_version = fake_http_version()
    status_code = fake_status_code()
    response_size = fake_response_size(status_code)
    referer = fake_referer()
    user_agent = fake_user_agent()

    log_line = f"{ip} - - {timestamp} \"{method} {path} {http_version}\" {status_code} {response_size} \"{referer}\" \"{user_agent}\""
    return log_line


if __name__ == "__main__":
    for _ in range(1000):
        print(generate_log_line())
