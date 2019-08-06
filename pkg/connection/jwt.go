// Copyright ¬© 2018 The Knative Authors
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

package connection

import (
	"context"
	"io/ioutil"
	"time"

	"golang.org/x/oauth2/jwt"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/yaml"
)

type key struct {
	PrivateKeyID string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`

	Email       string `json:"email"`
	ClientEmail string `json:"client_email"` // Alternate for Email

	TokenURL string `json:"token_url"`
	TokenURI string `json:"token_uri"` // Alternate for TokenURL

	Subject string `json:"subject"`
}

func (k *key) ToJWTConfig(scopes []string) (*jwt.Config, error) {
	ret := &jwt.Config{}
	ret.PrivateKeyID = k.PrivateKeyID
	ret.PrivateKey = []byte(k.PrivateKey)
	if k.ClientEmail != "" {
		ret.Email = k.ClientEmail
	} else {
		ret.Email = k.Email
	}
	if k.TokenURI != "" {
		ret.TokenURL = k.TokenURI
	} else {
		ret.TokenURL = k.TokenURL
	}

	ret.Subject = k.Subject

	ret.Scopes = scopes

	return ret, nil
}

type JWTTokenClientConfig struct {
	config clientcmd.ClientConfig
	token  string
}

func (c *JWTTokenClientConfig) RawConfig() (clientcmdapi.Config, error) {
	return c.config.RawConfig()
}

func (c *JWTTokenClientConfig) ClientConfig() (*rest.Config, error) {
	ret, err := c.config.ClientConfig()
	if err != nil {
		return nil, err
	}
	ret.BearerToken = c.token
	return ret, nil
}

func (c *JWTTokenClientConfig) Namespace() (string, bool, error) {
	return c.config.Namespace()
}

func (c *JWTTokenClientConfig) ConfigAccess() clientcmd.ConfigAccess {
	return c.config.ConfigAccess()
}

func WrapConfigWithJWT(config clientcmd.ClientConfig, keyFile string, scopes []string, tokenURL string) (clientcmd.ClientConfig, error) {

	fileContents, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	k := &key{}
	err = yaml.Unmarshal(fileContents, k)

	jwtConfig, err := k.ToJWTConfig(scopes)
	if err != nil {
		return nil, err
	}
	if tokenURL != "" {
		jwtConfig.TokenURL = tokenURL
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ts := jwtConfig.TokenSource(ctx)
	tok, err := ts.Token()
	if err != nil {
		return nil, err
	}
	return &JWTTokenClientConfig{config, tok.AccessToken}, nil
}
