package session

type Content struct {
	SelectFields []string
	TableName    string
}

func Generate(selectFields []string, tableName string) Content {
	return Content{
		SelectFields: selectFields,
		TableName:    tableName,
	}
}
