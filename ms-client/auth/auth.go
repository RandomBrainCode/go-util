package auth

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	msgraph "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go-core/authentication"
	"time"
)

type AuthParams struct {
	clientID     string
	clientSecret string
	tenantID     string
	authorityURL string
	redirectURI  string
	scopes       []string
}

type TokenCredential struct {
	accessToken string
}

func (c *TokenCredential) GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: c.accessToken, ExpiresOn: time.Now().Add(1 * time.Hour)}, nil
}

func GetCredential(clientSecret string) (confidential.Credential, error) {
	return confidential.NewCredFromSecret(clientSecret)
}

func GetClient(clientID string, tenantID string, authorityURL string, credential confidential.Credential) (confidential.Client, error) {
	if len(authorityURL) > 0 && authorityURL[len(authorityURL)-1] != '/' {
		authorityURL = authorityURL + "/"
	}
	authority := authorityURL + tenantID

	return confidential.New(authority, clientID, credential)
}

func GetAuthURL(appClient confidential.Client, clientID string, redirectURI string, scopes []string) (string, error) {
	return appClient.AuthCodeURL(context.TODO(), clientID, redirectURI, scopes)
}

func GetAuthProvider(userToken *TokenCredential) (*authentication.AzureIdentityAuthenticationProvider, error) {
	return authentication.NewAzureIdentityAuthenticationProvider(userToken)
}

func GetGraphAdapter(provider *authentication.AzureIdentityAuthenticationProvider) (*msgraph.GraphRequestAdapter, error) {
	return msgraph.NewGraphRequestAdapter(provider)
}

func GetGraphClient(adapter *msgraph.GraphRequestAdapter) *msgraph.GraphServiceClient {
	return msgraph.NewGraphServiceClient(adapter)
}
