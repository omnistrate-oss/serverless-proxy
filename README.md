# Omnistrate Postgres Proxy Example
This is a simple example to show how to leverage postgres proxy with Omnistrate platform to build serverless product.
This code is not intended for production use but can be used as reference.

## Build Example Proxy Docker Image

```
make docker-build
```


## Setup Simple Postgres Serverless With Omnistrate

**Step 1: Import Service Definition From Docker Compose**

Sample docker compose file exist in /dockercompose/postgres_demo.yaml folder

You can import the docker image using the UI using the **Import docker compose** option

![image](https://github.com/omnistrate/pg-proxy/assets/1789738/08a6257c-5877-41cb-a827-7ab23dbe537b)


The docker compose example uses two images: 
- njnjyyh/postgres-demo:latest 
- docker.io/njnjyyh/pg-proxy-demo:latest <- proxy image built fromt this repo

Setup Service Name to **Postgres Serverless**. Using this name is important as the following commands will refernce the service with this name. 


**Step 2: Spinup Postgres Proxy Instance**

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

<img width="803" alt="Screenshot 2023-11-16 at 2 53 58 PM" src="https://github.com/omnistrate/pg-proxy/assets/19898780/1cd3cd44-cf28-4fda-bac1-a924b04609bf">


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
curl -X 'POST' \  'https://api.omnistrate.cloud/2022-09-01-00/resource-instance/<serviceProviderId>/postgres-serverless/v1/dev/postgres-serverless-omnistrate-hosted/postgres-serverless-omnistrate-hosted-model-omnistrate-dedicated-tenancy/proxy' \
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

Note: ***Postgres Serverless*** is the service name that being used in this example, other service names will subject to different url formats.

**Step 3: Setup Postgres Instance**

In Omnistrate console access page, provision postgres instance in us-east-2 once proxy instance is up and running
You can find proxy instance status via operate page
<img width="956" alt="Screenshot 2023-11-16 at 3 06 54 PM" src="https://github.com/omnistrate/pg-proxy/assets/19898780/2f01e08e-29a8-4f38-be4e-3bf8393226da">

And the provision the new instance

<img width="399" alt="Screenshot 2023-11-16 at 2 54 52 PM" src="https://github.com/omnistrate/pg-proxy/assets/19898780/4a986be0-6ba9-4091-bbfa-67c159b818e9">

**Step 4: Access postgres Instance**

Once postgres instance is up and running, check the connectivity from access page and get the endpoint/port for connection. Note that the endpoint shown in the page is pointing to the proxy and not directly to the provisioned instance. 

<img width="1626" alt="Screenshot 2023-11-16 at 3 07 26 PM" src="https://github.com/omnistrate/pg-proxy/assets/19898780/1b51d1d3-44c1-45fe-a1f2-be156fac904c">

Note that if the instance is not used for some time it will be Stopped automatically. 

```
psql -U postgres -W -h <endpoint> postgres -p <port>
```

The default password for the demo is **postgres**

Once we attempt to start the connection the instance will be automatically started, it will take a few minutes for the server to start and then you can operate on the open connection. 
While the connection is open the instance will be Running and after the connection is close the instance will be Stopped automatically. Auto stop in this example relies on prometheus metrics pg_stat_database_num_backends. If you want to use other metrics, please update docker compose.

## Update 

Step 2 can be skipped now, you can directly go to Step 3 and omnistrate platform will auto provision proxy instance for you.
