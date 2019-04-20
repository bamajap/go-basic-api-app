/*
Author: Jason Payne
*/
package dummydb

import (
	"fmt"
	"sort"
)

/*
Product -
*/
type Product struct {
	Id    int `json:"id,string"`
	Name  string
	Price float64 `json:",string"`
}

func (p Product) String() string {
	return fmt.Sprintf("<(Id: %v) {%v} @ %v>", p.Id, p.Name, p.Price)
}

type Products []Product

var Items Products

func (pArr Products) GetAll() (Products, error) {
	// Price-descending sort
	sort.Slice(pArr, func(i, j int) bool { return pArr[i].Price > pArr[j].Price })
	return pArr, nil
}

func (pArr *Products) AddProduct(newProduct Product) error {
	*pArr = append(*pArr, newProduct)
	return nil
}

func (pArr Products) GetProduct(product *Product) error {
	for _, p := range pArr {
		if product.Id == p.Id {
			*product = p
			return nil
		}
	}
	return fmt.Errorf("Product <%v> does not exist", product.Id)
}

func (pArr *Products) UpdateProduct(newProduct Product) error {
	for i, op := range *pArr {
		if op.Id == newProduct.Id {
			(*pArr)[i] = newProduct
			return nil
		}
	}
	return fmt.Errorf("Product <%v> does not exist", newProduct.Id)
}

func (pArr *Products) DeleteProduct(p Product) error {
	for i, op := range *pArr {
		if op.Id == p.Id {
			*pArr = append((*pArr)[:i], (*pArr)[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("Product <%v> does not exist", p.Id)
}

func Initialize() error {
	Items = Products{
		{1, "Apple", 0.98},
		{2, "Orange", 0.98},
		{3, "Bananas", 2.25},
		{4, "Frozen Pizza", 4.99},
	}
	return nil
}

func Cleanup() error {
	fmt.Println("Cleaning up...")
	return nil
}
