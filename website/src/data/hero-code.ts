export const heroCode = `package main

import (
    "fmt"

    "github.com/larsartmann/go-branded-id"
)

type UserBrand struct{}

func (UserBrand) Name() string { return "User" }

type OrderBrand struct{}

type UserID = id.ID[UserBrand, string]
type OrderID = id.ID[OrderBrand, string]

func GetUser(id UserID) error   { return nil }
func GetOrder(id OrderID) error { return nil }

func main() {
    userID := id.NewID[UserBrand]("user-123")

    fmt.Println(userID) // User:user-123

    // GetOrder(userID) // COMPILE ERROR
    GetUser(userID)     // OK
}`;
