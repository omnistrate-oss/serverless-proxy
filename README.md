# Omnistrate Postgres Proxy Example
This is a simple example to show how to leverage postgres proxy with Omnistrate platform to build serverless product.

## Build Example Proxy Docker Image

- Generic proxy example

    ```
    make docker-build-generic
    ```

- MySQL proxy example  

    ```
    make docker-build-generic
    ```
    
- Postgres proxy example

    ```
    make docker-build-generic
    ```

## Setup Simple Postgres Serverless With Omnistrate

Pre-requisite:
Please follow https://github.com/omnistrate-oss/account-setup to setup AWS bootstrap account/role

**Step 1: Import Service Definition From Docker Compose**

Sample docker compose file exist in /dockercompose/postgres_demo.yaml folder

You can import the docker image using the UI using the **Import docker compose** option

![image](https://github.com/omnistrate-oss/serverless-proxy/assets/1789738/08a6257c-5877-41cb-a827-7ab23dbe537b)


The docker compose example uses two images: 
- docker.io/njnjyyh/pg-proxy-demo:1 <- postgres image with prometheus metrics exporter (https://github.com/prometheus-community/postgres_exporter)
- docker.io/njnjyyh/pg-proxy-demo:2.0 <- proxy image built from this repo

Setup Service Name to **Postgres Serverless**. Using this name is important as the following commands will reference the service with this name. 

**Step 2: Setup Postgres Instance**

In Omnistrate console access page, provision a Postgres instance 

**Step 3: Access Postgres instance**

Once Postgres instance is up and running, check the connectivity from access page and get the endpoint/port for connection. Note that the endpoint shown in the page is pointing to the proxy and not directly to the provisioned instance. 

Note that if the instance is not used for some time it will be stopped automatically. 

```
psql -U postgres -W -h <endpoint> postgres -p <port>
```

The default password for the demo is **postgres**

Once we attempt to start the connection the instance will be automatically started, it will take a few minutes for the server to start and then you can operate on the open connection.

While the connection is open the instance will be Running and after the connection is close the instance will be Stopped automatically. Auto stop in this example relies on prometheus metrics pg_stat_database_num_backends. If you want to use other metrics, please update docker compose.
