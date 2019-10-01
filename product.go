/*
Copyright 2017 Ontario Systems

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package isclib

// Product represents a particular ISC product
type Product uint

const (
	// Cache is the ISC product Cache
	Cache Product = iota
	// Ensemble is the ISC product Ensemble
	Ensemble
	// Iris is the ISC product IRIS Data Platform
	Iris
	// None indicates that there are no ISC products
	None Product = 0
)

// ParseProduct parses a string representing a ISC product into a Product.
// The default for unknown strings is Cache.
func ParseProduct(product string) Product {
	switch product {
	default:
		return Cache
	case "Cache":
		return Cache
	case "Ensemble":
		return Ensemble
	case "IDP", "IRIS":
		return Iris
	}
}
