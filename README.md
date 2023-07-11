## shorty-go
A Go URL Shortener 

This is a simple URL shortener service built with Go and Redis. It provides a REST API that accepts a URL as an argument and returns a shortened URL as a result. It also implements a redirection API that redirects users to the original URL when they click on the shortened URL.

Garden did the magic with building, and deploying it.

### Requirements

- Docker
- Redis
- Garden

### Usage

1. Clone this repository: `git clone https://github.com/Saurabh-Thakre/shorty-go.git`
2. Change into the project directory: `cd shorty-go`
3. Create a `.env` file with the following environment variables:

```
DB_ADDR="db:6379"
DB_PASS=""
APP_PORT=":3000"
DOMAIN="localhost:3000"
API_QUOTA=10
```

### Here is how you can use the application
5. To access the URL shortener, send a POST request to `http://localhost:3000/api/v1` with a JSON body containing a `url` field with the URL you want to shorten. For example:

```
curl --location --request POST 'http://localhost:3000/api/v1' \
--header 'Content-Type: application/json' \
--data-raw '{
    "url": "https://www.google.com/search?q=golang&oq=golang&aqs=chrome.0.69i59j0i433j69i60j0i131i433j0i433j46i131i433j0i433j46j0i131i433j69i60.1829j0j7&sourceid=chrome&ie=UTF-8"
}'
```

The response will contain a shortened URL, for example:

```
{
  "url": "https://www.youtube.com/watch?v=h2RdcrMLsdfsfQ",
  "short": "localhost:3000/59dec7",
  "expiry": 24,
  "rate_limit": 7,
  "rate_limit_reset": 29
}
```
