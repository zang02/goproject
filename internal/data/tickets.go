package data

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Ticket struct {
	ID        string    `bson:"_id" json:"_id,omitempty"`
	UserLogin string    `json:"userlogin"`
	CreatedAt string    `json:"created,omitempty"`
	Total     int64     `json:"total,omitempty"`
	Products  []Product `json:"products"`
}

type Product struct {
	Name   string `json:"name"`
	Price  int    `json:"price"`
	Amount int    `json:"amount"`
}

type TicketModel struct {
	DB *mongo.Database
}

func (t *TicketModel) Insert(ticket Ticket) (Ticket, error) {
	res, err := t.DB.Collection("tickets").InsertOne(context.TODO(), ticket)
	ticket.ID = fmt.Sprint(res.InsertedID)
	if err != nil {
		return Ticket{}, err
	}
	return ticket, nil
}

func (t *TicketModel) GetById(id string) (Ticket, error) {
	var ticket Ticket
	err := t.DB.Collection("tickets").FindOne(context.TODO(), bson.M{"_id": id}).Decode(&ticket)
	return ticket, err
}

func (t *TicketModel) GetLatest() ([]Ticket, error) {
	var tickets []Ticket
	collection := t.DB.Collection("tickets")
	options := options.Find().SetSort(bson.D{{"created", 1}})
	cursor, err := collection.Find(context.Background(), bson.D{}, options)
	if err != nil {
		return []Ticket{}, err
	}

	if err = cursor.All(context.TODO(), &tickets); err != nil {
		return []Ticket{}, err
	}
	return tickets, nil
}
