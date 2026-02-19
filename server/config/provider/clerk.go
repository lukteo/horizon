package provider

import (
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
)

func NewClerkProvider(env *EnvProvider) *user.Client {
	config := &clerk.ClientConfig{}
	config.Key = &env.clerkSecretKey

	client := user.NewClient(config)

	return client
}
