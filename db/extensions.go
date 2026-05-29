package db

import (
	"context"
)

type Attribute struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type AttributeValue struct {
	ID          string `json:"id"`
	AttributeID string `json:"attribute_id"`
	Value       string `json:"value"`
}

type CreateAttributeValueParams struct {
	AttributeID string
	Value       string
}

func (q *Queries) ListAttributes(ctx context.Context) ([]Attribute, error) {
	rows, err := q.db.QueryContext(ctx, "SELECT id, name, created_at FROM attributes")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Attribute
	for rows.Next() {
		var i Attribute
		if err := rows.Scan(&i.ID, &i.Name, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}

func (q *Queries) CreateAttribute(ctx context.Context, name string) (Attribute, error) {
	var i Attribute
	err := q.db.QueryRowContext(ctx, "INSERT INTO attributes (name) VALUES ($1) RETURNING id, name, created_at", name).Scan(&i.ID, &i.Name, &i.CreatedAt)
	return i, err
}

func (q *Queries) ListAttributeValues(ctx context.Context, attributeID string) ([]AttributeValue, error) {
	rows, err := q.db.QueryContext(ctx, "SELECT id, attribute_id, value FROM attribute_values WHERE attribute_id = $1", attributeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AttributeValue
	for rows.Next() {
		var i AttributeValue
		if err := rows.Scan(&i.ID, &i.AttributeID, &i.Value); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}

func (q *Queries) CreateAttributeValue(ctx context.Context, arg CreateAttributeValueParams) (AttributeValue, error) {
	var i AttributeValue
	err := q.db.QueryRowContext(ctx, "INSERT INTO attribute_values (attribute_id, value) VALUES ($1, $2) RETURNING id, attribute_id, value", arg.AttributeID, arg.Value).Scan(&i.ID, &i.AttributeID, &i.Value)
	return i, err
}
