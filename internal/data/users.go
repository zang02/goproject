package data

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserModel struct {
	DB *mongo.Database
}

type User struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Login      string             `json:"login"`
	Email      string             `json:"email"`
	Name       string             `json:"name"`
	Password   string             `json:"password"`
	CreateDate string             `json:"create_date"`
}

func (u *UserModel) Insert(user User) error {
	user.CreateDate = humanDate(time.Now().Add(time.Hour * 6))

	_, err := u.DB.Collection("users").InsertOne(context.TODO(), user)

	return err
}

func (u *UserModel) GetByLogin(login string) (User, error) {
	var user User
	err := u.DB.Collection("users").FindOne(context.TODO(), bson.M{"login": login}).Decode(&user)
	return user, err
}

func (u *UserModel) GetAllUsers() ([]User, error) {
	cursor, err := u.DB.Collection("users").Find(context.Background(), bson.D{})
	if err != nil {
		return []User{}, err
	}
	var users []User
	if err = cursor.All(context.TODO(), &users); err != nil {
		return []User{}, err
	}
	return users, nil
}

func (u *UserModel) DeleteUserByLogin(login string) error {
	collection := u.DB.Collection("users")
	_, err := collection.DeleteOne(context.TODO(), bson.M{"login": login})
	return err
}

func (u *UserModel) UpdateUserByLogin(login string, newUser User) error {
	collection := u.DB.Collection("users")
	res := collection.FindOneAndUpdate(context.TODO(), bson.M{"login": login}, bson.M{"$set": newUser})
	return res.Err()
}
