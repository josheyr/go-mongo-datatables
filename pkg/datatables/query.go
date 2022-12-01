package datatables

import "go.mongodb.org/mongo-driver/bson/primitive"

// Query structure, used to query the database
type Query struct {
	TableName string            `json:"table_name"`
	Fields    []string          `json:"fields"`
	Filters   []Filter          `json:"filters"`
	OrderBy   map[string]bool   `json:"order_by"`
	Limit     int               `json:"limit"`
	Offset    int               `json:"offset"`
	SearchBy  string            `json:"search_by"`
	Searches  map[string]string `json:"search_fields"`
	Output    string            `json:"output"`
	Download  bool              `json:"download"`
}

type Filter struct {
	Field string
	Value interface{}
}

type FilterValue struct {
	Type     string `json:"type"`
	Int      int64
	Float    float64
	Str      string
	Bool     bool
	IntArr   []int64 // used for min and max
	FloatArr []float64
}

type Response struct {
	Data          []primitive.D
	Count         int64
	FilteredCount int64
}