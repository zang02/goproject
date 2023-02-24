package data

import (
	"go.mongodb.org/mongo-driver/mongo"
)

// dependency injection pattern
type Models struct {
	Tokens  TokenModel
	Users   UserModel
	Tickets TicketModel
}

func NewModels(db *mongo.Database) Models {
	return Models{
		Tickets: TicketModel{DB: db},
		Tokens:  TokenModel{DB: db},
		Users:   UserModel{DB: db},
	}
}
