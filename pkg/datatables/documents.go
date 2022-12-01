package datatables

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RetrieveDocuments function, used to retrieve documents from the database
// converts the query to a bson.D object, and then calls the mongo.Collection.Find() function
func RetrieveDocuments(query *Query, ctx context.Context, db *mongo.Database, searchFields []string) (*Response, error) {
	collection := db.Collection(query.TableName)

	findOptions := options.Find()

	if query.Limit > 0 {
		findOptions.SetLimit(int64(query.Limit))
	}

	findOptions.SetSkip(int64(query.Offset))

	// Generate orderBy bson.D object (ordered)
	orderByBson := bson.D{}

	// foreach in findOptions.OrderBy
	for field, desc := range query.OrderBy {
		var orderByInt = 1
		if desc == true {
			orderByInt = -1
		}

		// append to bsonD
		orderByBson = append(orderByBson,
			bson.E{Key: field, Value: orderByInt})
	}

	// generate filter bson.M object (unordered)
	fieldsBson := bson.M{}
	for _, field := range query.Fields {
		fieldsBson[field] = 1
	}

	// remove _id field from fields
	fieldsBson["_id"] = 0

	// set find options
	findOptions.Sort = orderByBson
	findOptions.Projection = fieldsBson

	var filteredOrSearched = false
	var filtered = false

	// generate search bson.M object (unordered)
	findBson := bson.M{}

	// TODO: multiple filters on same field using $and or $or
	var fieldList []string

	for _, filter := range query.Filters {
		filtered = true
		filteredOrSearched = true
		fieldList = append(fieldList, filter.Field)
	}

	var andBson []bson.M

	if query.SearchBy != "" && searchFields != nil {
		searchBson := bson.M{}
		filteredOrSearched = true
		filtered = true

		searchBson["$or"] = []bson.M{}

		for _, field := range searchFields {
			searchBson["$or"] = append(searchBson["$or"].([]bson.M),
				bson.M{
					field: primitive.Regex{Pattern: query.SearchBy, Options: "i"},
				})
		}

		andBson = append(andBson, searchBson)
	}

	// foreach field, foreach value
	for _, field := range fieldList {
		// foreach filter with field
		var orBson []bson.M
		for _, filter := range query.Filters {
			if filter.Field == field {
				orBson = append(orBson,
					bson.M{
						field: filter,
					})
			}
		}

		andBson = append(andBson,
			bson.M{
				"$or": orBson,
			})
	}

	if filtered {
		findBson["$and"] = andBson
	}

	// execute query
	cursor, err := collection.Find(ctx, findBson, findOptions)
	if err != nil {
		return nil, err
	}

	var data []primitive.D
	err = cursor.All(ctx, &data)
	if err != nil {
		return nil, err
	}

	// set totalCount and filtered for pagination
	totalCount, err := collection.EstimatedDocumentCount(ctx, nil)
	var filteredCount = totalCount

	if filteredOrSearched {
		filteredCount, err = collection.CountDocuments(ctx, findBson)
	}

	if data == nil {
		data = []primitive.D{}
	}

	var response = &Response{
		Data:          data,
		Count:         totalCount,
		FilteredCount: filteredCount,
	}

	return response, nil
}