package workspace

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const metaSK = "META"

type envelope[T any] struct {
	PK   string `dynamodbav:"PK"`
	SK   string `dynamodbav:"SK"`
	Item T      `dynamodbav:"Item"`
}

type Repo struct {
	db    *dynamodb.Client
	table string
}

func NewRepo(db *dynamodb.Client, table string) *Repo {
	return &Repo{db: db, table: table}
}

func pk(channel, threadTS string) string { return "WS#" + channel + "#" + threadTS }

func msgSK(slackTS string) string { return "MSG#" + slackTS }

func (r *Repo) CreateWorkspace(ctx context.Context, ws Workspace) (bool, error) {
	item, err := attributevalue.MarshalMap(envelope[Workspace]{
		PK:   pk(ws.Channel, ws.ThreadTS),
		SK:   metaSK,
		Item: ws,
	})
	if err != nil {
		return false, err
	}

	_, err = r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &r.table,
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})
	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repo) GetWorkspace(ctx context.Context, channel, threadTS string) (*Workspace, error) {
	out, err := r.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &r.table,
		Key:       key(pk(channel, threadTS), metaSK),
	})
	if err != nil {
		return nil, err
	}
	if out.Item == nil {
		return nil, nil
	}

	var env envelope[Workspace]
	if err := attributevalue.UnmarshalMap(out.Item, &env); err != nil {
		return nil, err
	}
	return &env.Item, nil
}

func (r *Repo) UpdateStatus(ctx context.Context, channel, threadTS string, s Status) error {
	now, err := attributevalue.Marshal(time.Now().UTC())
	if err != nil {
		return err
	}
	_, err = r.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                &r.table,
		Key:                      key(pk(channel, threadTS), metaSK),
		UpdateExpression:         aws.String("SET Item.#s = :s, Item.updatedAt = :u"),
		ConditionExpression:      aws.String("attribute_exists(PK)"),
		ExpressionAttributeNames: map[string]string{"#s": "status"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":s": &types.AttributeValueMemberS{Value: string(s)},
			":u": now,
		},
	})
	return err
}

func (r *Repo) AppendMessage(ctx context.Context, channel, threadTS string, m Message) (bool, error) {
	item, err := attributevalue.MarshalMap(envelope[Message]{
		PK:   pk(channel, threadTS),
		SK:   msgSK(m.SlackTS),
		Item: m,
	})
	if err != nil {
		return false, err
	}

	_, err = r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &r.table,
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})
	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repo) ListMessages(ctx context.Context, channel, threadTS string, limit int32) ([]Message, error) {
	in := &dynamodb.QueryInput{
		TableName:              &r.table,
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :m)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: pk(channel, threadTS)},
			":m":  &types.AttributeValueMemberS{Value: "MSG#"},
		},
		ScanIndexForward: aws.Bool(true),
	}
	if limit > 0 {
		in.Limit = aws.Int32(limit)
	}

	out, err := r.db.Query(ctx, in)
	if err != nil {
		return nil, err
	}

	envs := make([]envelope[Message], 0, len(out.Items))
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &envs); err != nil {
		return nil, err
	}
	msgs := make([]Message, len(envs))
	for i := range envs {
		msgs[i] = envs[i].Item
	}
	return msgs, nil
}

func key(pkVal, skVal string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: pkVal},
		"SK": &types.AttributeValueMemberS{Value: skVal},
	}
}
