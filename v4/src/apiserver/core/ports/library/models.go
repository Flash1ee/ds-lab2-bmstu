package library

import "ds-lab2-bmstu/pkg/collections"

type Info struct {
	ID      string
	Name    string
	Address string
	City    string
}

type Infos collections.Countable[Info]

type Book struct {
	ID        string
	Name      string
	Author    string
	Genre     string
	Condition string
	Available uint64
}

type Books collections.Countable[Book]

type ReservedBook struct {
	Book    Book
	Library Info
}
