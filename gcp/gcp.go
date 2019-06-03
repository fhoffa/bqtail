package gcp

import (
	"context"
	"google.golang.org/api/option"

	htransport "google.golang.org/api/transport/http"
	"net/http"
)

const userAgent = "viant/bqtail"

var bigQueryScope = "https://www.googleapis.com/auth/bigquery"
var bigQueryInsertScope = "https://www.googleapis.com/auth/bigquery.insertdata"
var prodAddr = "https://www.googleapis.com/bigquery/v2/"

func getDefaultClient(ctx context.Context, scopes ...string) (*http.Client, error) {
	o := []option.ClientOption{
		option.WithScopes(scopes...),
		option.WithEndpoint(prodAddr),
		option.WithUserAgent(userAgent),
	}
	httpClient, _, err := htransport.NewClient(ctx, o...)
	return httpClient, err
}

//GetClient creates a new google cloud client.
func NewHttpClient(scopes ...string) (*http.Client, context.Context, error) {
	ctx := context.Background()
	scopes = append(scopes, bigQueryScope, bigQueryInsertScope)
	client, err := getDefaultClient(ctx, scopes...)
	return client, ctx, err

}
