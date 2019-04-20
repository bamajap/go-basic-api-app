# go-basic-api-app
Background
----------
Create a very basic API application using ​the Go language​, where we can retrieve product information.

Tasks
-----
1. Create a web application connected to your own DynamoDB instance.
    - Create a sample table that will store products. Products will minimally have a name and a price.
2. Your web application should utilize RESTful routing and be able to query out the relevant product information through its API, sorting by price descending by default.


Assumptions + Notes
-------------------
* App will be setup with a local DynamoDB instance.
* Return values will be presented in JSON format (or a short error message).
* Sorting DynamoDB query results in descending order is not intuitive, so, for simplification, query results will be manually sorted.
* Assuming that all update requests include values for the new Name and/or Price.


API
---
* Get All: GET http://localhost:8000
* Create: POST http://localhost:8000/product
* Read: GET http://localhost:8000/product/{id}
* Update: PUT http://localhost:8000/product/{id}
* Delete: DELETE http://localhost:8000/product/{id}

* DynamoDB Endpoint: http://localhost:8080
