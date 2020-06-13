// Copyright © 2019 Tobias Jeske
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/tjeske/containerflight/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type mongoDbProvider struct {
	client     *mongo.Client
	collection *mongo.Collection
}

var collectionName = "config"

func NewMongoDbProvider(host, user, password, database string) StorageProvider {
	clientOptions := options.Client().ApplyURI("mongodb://" + user + ":" + password + "@" + host)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	util.CheckErr(err)

	// test connectivity
	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		log.Fatal("Couldn't connect to the database", err)
	}
	log.Debug("Connected to MongoDB!")

	collection := client.Database(database).Collection(collectionName)

	return &mongoDbProvider{client: client, collection: collection}
}

func (p *mongoDbProvider) UpdateApp(app Template) {
	filter := bson.M{"Name": app.Name}
	options := options.Replace()
	options.SetUpsert(true)
	documentReturned, err := p.collection.ReplaceOne(context.TODO(), filter, app, options)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("Update configuration for app '%s': %+v -> %+v", app.Name, app, documentReturned)
}

func (p *mongoDbProvider) GetAllTemplates() *[]TemplateName {
	var apps []*Template
	filter := bson.M{}
	cur, err := p.collection.Find(context.TODO(), filter)
	util.CheckErrMsg(err, "Error on Finding all the documents")

	for cur.Next(context.TODO()) {
		var app Template
		err = cur.Decode(&app)
		util.CheckErrMsg(err, "Error on Decoding the document")

		apps = append(apps, &app)
	}

	res := make([]TemplateName, len(apps))
	for i, app := range apps {
		res[i] = TemplateName{Uuid: "hgjg", Name: app.Name}
	}

	return &res
}

func (p *mongoDbProvider) GetTemplateByName(name string) *Template {
	var app Template
	filter := bson.M{"Name": app.Name}
	documentReturned := p.collection.FindOne(context.TODO(), filter)
	documentReturned.Decode(&app)
	return &app
}
