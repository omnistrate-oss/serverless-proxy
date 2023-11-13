# Omnistrate Postgres Proxy Example
This is a simple example to show how to leverage postgres proxy with Omnistrate platform to build serverless product.
This code is not intended for production use case.

## Build Example Proxy Docker Image

```
make docker-build
```


## Setup Simple Supabase Serverless With Omnistrate

**Step 1: Import Service Definition From Docker Compose**

Sample docker compose file exist in /dockercompose/supabase_demo.yaml folder

The docker compose example uses two images: 
- njnjyyh/supabase-demo:latest <- supabase standard image built from **https://github.com/supabase/postgres/tree/develop/docker/all-in-one**
- docker.io/njnjyyh/pg-proxy-demo:latest <- proxy image built fromt this repo


**Step 2: Spinup Supabase Proxy Instance**

You need to get bearer token first via signup API

API docs can be found here: https://api.omnistrate.cloud/docs/external/
```
curl -X 'POST' \
  'https://api.omnistrate.cloud/2022-09-01-00/signin' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "email": "youruser@company.com",
  "hashedPassword": "hashedPassword"
}'
```
The hashedPasword can be generated using the command line
```
echo -n "yourPassword" | openssl dgst -sha256
```
Calling **signin** api returns a jwtToken value that needs to be used as bearer token in all subsequent calls. 

```
curl -X 'POST' \  'https://api.omnistrate.cloud/2022-09-01-00/resource-instance/<serviceProviderId>/<serviceKey>/<serviceAPIVersion>/<serviceEnvironmentKey>/<serviceModelKey>/<productTierKey>/proxy' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer xxxxxx' \
  -H 'Content-Type: application/json' \
  -d '{
  "cloud_provider": "aws",
  "region": "us-east-1",
  "requestParams": {
    "custom_availability_zone": "us-east-1a"
  }
}'

```

**Step 3: Setup Supabase Instance**

In Omnistrate console access page, provision supabase instance once proxy instance is up and running (you can find proxy instance status via operate page)

**Step 4: Access Supabase Instance**

Once supabase instance is up and running, check the connectivity from access page and get the endpoint/port for connection

```
psql -U postgres -W -h <endpoint> postgres -p <port>
```

**Note**

1. This is a pretty simple example to show how proxy can integrate with Omnistrate platform, please limit the number of backend instances for one proxy instance less than 9.
2. Auto stop in this example relies on prometheus metrics pg_stat_database_num_backends. If you want to use other metrics, please update docker compose.



