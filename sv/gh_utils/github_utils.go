package gh_utils

import (
	"context"
	"errors"
	"github.com/Khan/genqlient/graphql"
)

const GCLCLIENT_KEY = "genqlclient"

func GetGraphQLClient(ctx context.Context) (graphql.Client, error) {
	if cl, ok := ctx.Value(GCLCLIENT_KEY).(graphql.Client); ok {
		return cl, nil
	}
	return nil, errors.New("can't find any client in current context")
}

func InitContext(ctx context.Context, client graphql.Client) context.Context {
	return context.WithValue(ctx, GCLCLIENT_KEY, client)
}
