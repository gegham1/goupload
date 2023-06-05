# goupload
Simple service to upload and read Promotion records

### Running the app

```bash
# development
$ go get ./...
$ go run .
```

### Processing the file
As we access promotions by a number (incremental id) we can consider using arrays as well.

We can also consider storing promotions in their serialized state and which will further improve upload latency

Using Map
- constant access by key

Using Array
- constant access by index
- more memory efficient compare to maps as no key need to be stored


### Additional improvements:
- dockerize the application
- setup a CI/CD pipeline
- use K8s to deploy and scale the app on prod
- add health check endpoint
- use observability tool (datadog) to manage logs and monitor the API
- handling larger files
  1. to prevent memory leak and allow the app to scale horizontally, use redis instead of in-memory structure
- maintaining immutability in redis
  1. use prefix for all keys and remove all keys with the prefix before starting the upload process
  2. consider using redis distributed locks to prevent cases when two or more instances of the app can process a file in the same time (although as file upload will happen in every 30 min. this should not be a problem)
- performance in peak periods
  1. enable Autoscaling
  2. for read: with redis and autoscaling the app can easily handle millions of requests per minute
  3. for upload: app can easily handle predicted workload (req per 30 min)
