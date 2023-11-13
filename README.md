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

Setup Service Name to **Supabase Serverless** (using this name is important as the following commands will refernce the service with this name)


**Step 2: Spinup Supabase Proxy Instance**

You need to get bearer token first via **signup API**

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
Calling **signin API** returns a **jwtToken** value that needs to be used as bearer token in all subsequent calls. 

***Step 2a: List Services***
```
curl -X 'GET' \
  'https://api.omnistrate.cloud/2022-09-01-00/service' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer xxxxxx'
```
Calling **list services API** returns **serviceId** that needs to be used in subsequent calls.

Another way to get the **serviceId** is by navigating on the UI and get the id from the Url of the Service page. 

![image](https://github.com/omnistrate/pg-proxy/assets/1789738/99c318bb-fc1c-41d8-868b-e1b7d13d1db6)

***Step 2b: Describe Service***

Once you have the service Id you can call the following API to get the **serviceProviderId**  value

```
curl -X 'GET' \
  'https://api.omnistrate.cloud/2022-09-01-00/service/<serviceId>' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer xxxxxx'
```

Calling **describe service API** returns **serviceProviderId** that needs to be used in subsequent calls. 

***Step 2c: Provision Proxy Instance***
```
curl -X 'POST' \  'https://api.omnistrate.cloud/2022-09-01-00/resource-instance/<serviceProviderId>/supabase-serverless/v1/dev/supabase-serverless-omnistrate-hosted/supabase-serverless-omnistrate-hosted-model-omnistrate-dedicated-tenancy/proxy' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer xxxxxx' \
  -H 'Content-Type: application/json' \
  -d '{
  "cloud_provider": "aws",
  "region": "us-east-2",
  "requestParams": {
    "custom_availability_zone": "us-east-2a"
  }
}'

```

Note: ***Supbase Serverless*** is the service name that being used in this example, other service names will subject to different url formats.

**Step 3: Setup Supabase Instance**

In Omnistrate console access page, provision supabase instance in us-east-2 once proxy instance is up and running (you can find proxy instance status via operate page)

![image](https://github.com/omnistrate/pg-proxy/assets/1789738/03dafa77-2cd2-4abb-9159-b8a6bd5843de)


**Step 4: Access Supabase Instance**

Once supabase instance is up and running, check the connectivity from access page and get the endpoint/port for connection

```
psql -U postgres -W -h <endpoint> postgres -p <port>
```

**Notes**

1. This is a pretty simple example to show how proxy can integrate with Omnistrate platform, please limit the number of backend instances for one proxy instance less than 9.
2. Auto stop in this example relies on prometheus metrics pg_stat_database_num_backends. If you want to use other metrics, please update docker compose.



