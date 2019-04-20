/*
Author: Jason Payne
*/
package dynamodb

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

/*
Product - Go object representation of items that will be managed by the app.
*/
type Product struct {
	Id    int `json:"id"`
	Name  string
	Price float64
}

func (p Product) String() string {
	return fmt.Sprintf("<(Id: %v) {%v} @ %v>", p.Id, p.Name, p.Price)
}

// Products - wrapper for the DynamoDB Go type that will allow local methods to be called from DynamoDB instances.
type Products struct {
	*dynamodb.DynamoDB
}

// Items - global DynamoDB instance.
var Items Products

// TableName - name for the table that will serve as the DynamoDB instance.
const TableName = "Products"

// IdAttribute - attribute name for the partition key.
const IdAttribute = "id"

// GetAll - responds with all of the Products in price-descending order.
func (db Products) GetAll() ([]Product, error) {
	// Price-descending sort
	temp := []Product{}

	result, err := Items.Scan(&dynamodb.ScanInput{TableName: aws.String(TableName)})
	if err != nil {
		return nil, fmt.Errorf("Query GetAll failed:\n%v", err)
	}

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &temp)
	if err != nil {
		return nil, fmt.Errorf("Unmarshalling GetAll failed:\n%v", err)
	}

	// Manually sort the results to get a Price-descending sort
	sort.Slice(temp, func(i, j int) bool { return temp[i].Price > temp[j].Price })

	return temp, nil
}

// AddProduct - adds a new Product to the database.
func (db *Products) AddProduct(newProduct Product) error {
	data, err := dynamodbattribute.MarshalMap(newProduct)
	if err != nil {
		return fmt.Errorf("AddProduct -> Error marshalling product: %v", err)
	}

	// Setup the insert criteria.
	item := &dynamodb.PutItemInput{
		Item:      data,
		TableName: aws.String(TableName),
	}

	// Insert the new Product into the database.
	_, err = Items.PutItem(item)
	if err != nil {
		return fmt.Errorf("AddProduct -> New product could not be added: %v", err)
	}

	return nil
}

// GetProduct - if it exists, retrieves the requested Product from the database;
func (db Products) GetProduct(product *Product) error {
	// Setup query criteria.
	result, err := Items.Query(&dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		ScanIndexForward:       aws.Bool(false),
		KeyConditionExpression: aws.String("id = :id"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":id": {N: aws.String(strconv.Itoa(product.Id))},
		},
	})

	// If the product was found, then there should only be one item.
	for _, i := range result.Items {
		var p Product
		err = dynamodbattribute.UnmarshalMap(i, &p)
		if err != nil {
			return fmt.Errorf("Unmarshalling GetProduct failed:\n%v", err)
		}

		*product = p

		return nil
	}

	// If the product was not found, then return the appropriate status message.
	return fmt.Errorf("Product <%v> does not exist", product.Id)
}

// UpdateProduct - if found, this updates an existing Product; otherwise adds the new Product.
func (db *Products) UpdateProduct(newProduct Product) error {
	// Setup the update criteria.
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(TableName),
		Key: map[string]*dynamodb.AttributeValue{
			IdAttribute: {N: aws.String(strconv.Itoa(newProduct.Id))},
		},
		UpdateExpression:         aws.String("SET #n = :name, Price = :price"),
		ExpressionAttributeNames: map[string]*string{"#n": aws.String("Name")},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":name":  {S: aws.String(newProduct.Name)},
			":price": {N: aws.String(fmt.Sprintf("%f", newProduct.Price))},
		},
		ReturnValues: aws.String("ALL_NEW"),
	}

	// Execute the update.
	_, err := Items.UpdateItem(input)
	if err != nil {
		return fmt.Errorf("New product <%v> could not be updated/added: %v", newProduct, err)
	}

	return nil
}

// DeleteProduct - if it exists, deletes the specified Product.
func (db *Products) DeleteProduct(p Product) error {
	// Setup the delete criteria.
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(TableName),
		Key: map[string]*dynamodb.AttributeValue{
			IdAttribute: {N: aws.String(strconv.Itoa(p.Id))},
		},
		ReturnValues: aws.String("ALL_OLD"),
	}

	// Process the deletion.
	results, err := Items.DeleteItem(input)
	if err != nil {
		return fmt.Errorf("Product <%v> could not be deleted: %v", p, err)
	}

	// If there was nothing to delete, then return an appropriate message.
	if len(results.Attributes) == 0 {
		return fmt.Errorf("Product <%v> does not exist", p)
	}

	return nil
}

// Initialize - a helper function that sets up the database when the app is run for the first time.
func Initialize() error {
	// Initialize the AWS session.
	const Region = "us-west-2"
	const Endpoint = "http://localhost:8080"
	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String(Region),
		Endpoint: aws.String(Endpoint),
	})
	if err != nil {
		return fmt.Errorf("INITIALIZATION ERROR: %v", err)
	}

	// Initialize the DynamoDB instance.
	Items = Products{dynamodb.New(sess, aws.NewConfig().WithLogLevel(aws.LogDebugWithHTTPBody))}
	Items.listTables()

	tableExists, err := Items.tableExists(TableName)
	if err != nil {
		return fmt.Errorf("INITIALIZATION ERROR: %v", err)
	}

	if !tableExists {
		createTable()
	} else {
		fmt.Println("Table already exists!")
	}

	return nil
}

// Cleanup - a helper function that performs any cleanup processing.
func Cleanup() error {
	fmt.Println("Cleaning up...")
	// TODO - Delete table and dynamic resources
	return nil
}

// createTable - local helper function that creates the Products DynamoDB table.
func createTable() error {
	fmt.Println("Creating table...")

	// Setup table create criteria.
	input := &dynamodb.CreateTableInput{
		TableName: aws.String(TableName),
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(IdAttribute), KeyType: aws.String("HASH"),
			},
		},
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(IdAttribute), AttributeType: aws.String("N"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits: aws.Int64(10), WriteCapacityUnits: aws.Int64(10),
		},
	}

	// Create the table.
	if _, err := Items.CreateTable(input); err != nil {
		fmt.Println("Error during CreateTable:")
		return fmt.Errorf("%v", err)
	}

	fmt.Printf("Table '%v' successfully created!\n", TableName)

	// Initialize the database with some data for testing purposes.
	enterTestData()

	return nil
}

// enterTestData - local helper function that populates the database with some dummy data for testing purposes.
func enterTestData() error {
	products := []Product{
		{1, "Apple", 0.98},
		{2, "Orange", 0.98},
		{3, "Bananas", 2.25},
		{4, "Frozen Pizza", 4.99},
	}

	for _, p := range products {
		err := Items.AddProduct(p)
		if err != nil {
			return fmt.Errorf("Error entering test data: %v", err)
		}
	}

	return nil
}

// tableExists - local helper function that determines if a table with a specific name exists or not.
func (db *Products) tableExists(name string) (bool, error) {
	result, err := db.ListTables(&dynamodb.ListTablesInput{})

	if err != nil {
		fmt.Println("Error during ListTables:")
		return false, fmt.Errorf("%v", err)
	}

	for _, n := range result.TableNames {
		if *n == name {
			return true, nil
		}
	}

	return false, nil
}

// listTables - local helper function that lists all DynamoDB tables.
func (db *Products) listTables() error {
	result, err := db.ListTables(&dynamodb.ListTablesInput{})

	if err != nil {
		fmt.Println("Error during ListTables:")
		return fmt.Errorf("%v", err)
	}

	fmt.Println("Tables:")
	fmt.Println("")

	for _, n := range result.TableNames {
		fmt.Println(*n)
	}
	return nil
}
