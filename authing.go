// Package authing provides a Graphql client for `Authing`` which is an IDaaS provider
package authing

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"

	"github.com/kelvinji2009/graphql"
	"golang.org/x/oauth2"
)

const (
	// Development env
	userEndpointDevURL  = "http://users.authing.dodora.cn/graphql"
	oauthEndpointDevURL = "http://oauth.authing.dodora.cn/graphql"
	// Production env
	userEndpointProdURL  = "https://users.authing.cn/graphql"
	oauthEndpointProdURL = "https://oauth.authing.cn/graphql"
)

const pubPEM = `
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC4xKeUgQ+Aoz7TLfAfs9+paePb
5KIofVthEopwrXFkp8OCeocaTHt9ICjTT2QeJh6cZaDaArfZ873GPUn00eOIZ7Ae
+TiA2BKHbCvloW3w5Lnqm70iSsUi5Fmu9/2+68GZRH9L7Mlh8cFksCicW2Y2W2uM
GKl64GDcIq3au+aqJQIDAQAB
-----END PUBLIC KEY-----`

// Client is a client for interacting with the GraphQL API of `Authing`
type Client struct {
	client *graphql.Client

	clientID string

	// Log is called with various debug information.
	// To log to standard out, use:
	//  client.Log = func(s string) { log.Println(s) }
	Log func(s string)
}

// NewClient creates a new Authing user endpoint GraphQL API client
func NewClient(clientID string, appSecret string, isDev bool) *Client {
	c := &Client{
		clientID: clientID,
	}

	if c.client == nil {
		var endpointURL string
		if isDev {
			endpointURL = userEndpointDevURL
		} else {
			endpointURL = userEndpointProdURL
		}
		client := graphql.NewClient(endpointURL, nil)
		accessToken, err := getAccessTokenByAppSecret(client, clientID, appSecret)
		if err != nil {
			log.Println(err)
			return nil
		}

		src := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: accessToken},
		)

		httpClient := oauth2.NewClient(context.Background(), src)

		c.client = graphql.NewClient(endpointURL, httpClient)
	}

	return c
}

// NewOauthClient creates a new Authing oauth endpoint GraphQL API client
func NewOauthClient(clientID string, appSecret string, isDev bool) *Client {
	c := &Client{
		clientID: clientID,
	}

	if c.client == nil {
		var endpointURL string
		if isDev {
			endpointURL = userEndpointDevURL

		} else {
			endpointURL = userEndpointProdURL
		}
		client := graphql.NewClient(endpointURL, nil)
		accessToken, err := getAccessTokenByAppSecret(client, clientID, appSecret)
		if err != nil {
			log.Println(err)
			return nil
		}

		src := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: accessToken},
		)

		httpClient := oauth2.NewClient(context.Background(), src)

		if isDev {
			endpointURL = oauthEndpointDevURL

		} else {
			endpointURL = oauthEndpointProdURL
		}

		c.client = graphql.NewClient(endpointURL, httpClient)
	}

	return c
}

func (c *Client) logf(format string, args ...interface{}) {
	c.Log(fmt.Sprintf(format, args...))
}

func getAccessTokenByAppSecret(client *graphql.Client, clientID string, appSecret string) (string, error) {
	var q struct {
		GetAccessTokenByAppSecret graphql.String `graphql:"getAccessTokenByAppSecret(secret: $secret, clientId: $id)"`
	}

	variables := map[string]interface{}{
		"id":     graphql.String(clientID),
		"secret": graphql.String(appSecret),
	}

	err := client.Query(context.Background(), &q, variables)
	if err != nil {
		return "", err
	}

	accessToken := string(q.GetAccessTokenByAppSecret)
	return accessToken, err
}

func encryptPassword(password []byte) string {
	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		panic("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic("failed to parse DER encoded public key: " + err.Error())
	}

	pubKey := pub.(*rsa.PublicKey)

	cipherPassword, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, password)
	if err != nil {
		panic("failed to encrypt the password with PKCS1v15" + err.Error())
	}
	base64Password := base64.StdEncoding.EncodeToString(cipherPassword)

	return base64Password
}

//------------------------------------------------------------------------------------

// UserRegisterInput ...
type UserRegisterInput struct {
	Email            graphql.String `json:"email"`
	Phone            graphql.String `json:"phone"`
	Password         graphql.String `json:"password"`
	RegisterInClient graphql.String `json:"registerInClient"` // FIXME: Mandotory
}

// UserRegisterMutation ...
type UserRegisterMutation struct {
	Register struct {
		Unionid          graphql.String
		Email            graphql.String
		Phone            graphql.String
		EmailVerified    graphql.Boolean
		PhoneVerified    graphql.Boolean
		RegisterInClient graphql.String
		RegisterMethod   graphql.String
		Token            graphql.String
		TokenExpiredAt   graphql.String
		// TODO: more needed fields from `ExtendUser`
	} `graphql:"register(userInfo: $userInfo)"`
}

// Register ...
func (c *Client) Register(input *UserRegisterInput) (UserRegisterMutation, error) {
	var m UserRegisterMutation

	password := graphql.String(encryptPassword([]byte(string(input.Password))))

	variables := map[string]interface{}{
		"userInfo": UserRegisterInput{
			Email:            input.Email,
			Phone:            input.Phone,
			Password:         password,
			RegisterInClient: input.RegisterInClient,
		},
	}

	err := c.client.Mutate(context.Background(), &m, variables)

	if err != nil {
		// log.Println("Register failed: " + err.Error())
		return m, err
	}

	return m, nil
}

//------------------------------------------------------------------------------------

// UserLoginInput ...
type UserLoginInput struct {
	Unionid          graphql.String `json:"unionid"`
	Email            graphql.String `json:"email"`
	Phone            graphql.String `json:"phone"`
	Password         graphql.String `json:"password"`
	LastIP           graphql.String `json:"lastIP,omitempty"` // FIXME: Mandotory in struct
	RegisterInClient graphql.String `json:"registerInClient"`
	VerifyCode       graphql.String `json:"verifyCode,omitempty"`
}

// UserLoginMutation ...
type UserLoginMutation struct {
	Login struct {
		ID             graphql.String `graphql:"_id"`
		Email          graphql.String
		EmailVerified  graphql.Boolean
		Username       graphql.String
		Nickname       graphql.String
		Company        graphql.String
		Photo          graphql.String
		Browser        graphql.String
		Token          graphql.String
		TokenExpiredAt graphql.String
		LoginsCount    graphql.Int
		LastLogin      graphql.String
		// LastIP         graphql.String //FIXME: it may cause `mutation` failed
		SignedUp  graphql.String
		Blocked   graphql.Boolean
		IsDeleted graphql.Boolean
		// TODO: more needed fields from `ExtendUser`
	} `graphql:"login(unionid: $unionid, email: $email, phone: $phone, password: $password, lastIP: $lastIP, registerInClient: $registerInClient, verifyCode: $verifyCode)"`
}

// Login ...
func (c *Client) Login(input *UserLoginInput) (UserLoginMutation, error) {
	var m UserLoginMutation

	password := graphql.String(encryptPassword([]byte(string(input.Password))))

	variables := map[string]interface{}{
		"unionid":          input.Unionid,
		"email":            input.Email,
		"phone":            input.Phone,
		"password":         password,
		"lastIP":           input.LastIP,
		"registerInClient": input.RegisterInClient,
		"verifyCode":       input.VerifyCode,
	}

	err := c.client.Mutate(context.Background(), &m, variables)

	if err != nil {
		// log.Println("Login failed: " + err.Error())
		return m, err
	}

	return m, nil
}

//------------------------------------------------------------------------------------

// CheckLoginStatusQuery ...
type CheckLoginStatusQuery struct {
	CheckLoginStatus struct {
		Message graphql.String
		Code    graphql.Int
		Status  graphql.Boolean
	} `graphql:"checkLoginStatus"`
}

// CheckLoginStatus ...
func (c *Client) CheckLoginStatus() (CheckLoginStatusQuery, error) {
	var q CheckLoginStatusQuery

	err := c.client.Query(context.Background(), &q, nil)
	if err != nil {
		// log.Println("Check login status failed: " + err.Error())
		return q, err
	}

	return q, nil
}

//------------------------------------------------------------------------------------

// UserQueryParameter ...
type UserQueryParameter struct {
	ID               graphql.String `graphql:"_id"`
	RegisterInClient graphql.String `json:"registerInClient"`
	// Token                 graphql.String  `json:"token,omitempty"` // TODO:
	// Auth                  graphql.Boolean `json:"auth,omitempty"`
	// UserLoginHistoryPage  graphql.Int     `json:"userLoginHistoryPage,omitempty"`
	// UserLoginHistoryCount graphql.Int     `json:"userLoginHistoryCount,omitempty"`
}

// UserQuery ...
type UserQuery struct {
	User struct {
		ID             graphql.String `graphql:"_id"`
		Email          graphql.String
		Unionid        graphql.String
		EmailVerified  graphql.Boolean
		Phone          graphql.String
		PhoneVerified  graphql.Boolean
		Username       graphql.String
		Nickname       graphql.String
		Photo          graphql.String
		Company        graphql.String
		Token          graphql.String
		TokenExpiredAt graphql.String
		LoginsCount    graphql.Int
		LastLogin      graphql.String
		SignedUp       graphql.String
		Blocked        graphql.Boolean
		IsDeleted      graphql.Boolean
		// TODO: more fields from `ExtendUser`
	} `graphql:"user(id: $id, registerInClient: $registerInClient)"` // TODO: more parameters according to schema
}

// User ...
func (c *Client) User(parameter *UserQueryParameter) (UserQuery, error) {
	var q UserQuery

	variables := map[string]interface{}{
		"id":               parameter.ID,
		"registerInClient": parameter.RegisterInClient,
	}

	err := c.client.Query(context.Background(), &q, variables)

	if err != nil {
		return q, err
	}

	return q, nil
}

//------------------------------------------------------------------------------------

// UsersQueryParameter ...
type UsersQueryParameter struct {
	RegisterInClient graphql.String `json:"registerInClient,omitempty"`
	Page             graphql.Int    `json:"page,omitempty"`
	Count            graphql.Int    `json:"count,omitempty"`
}

// UsersQuery ...
type UsersQuery struct {
	Users struct {
		List []struct {
			ID            graphql.String `graphql:"_id"`
			Email         graphql.String
			Unionid       graphql.String
			EmailVerified graphql.Boolean
			Phone         graphql.String
			PhoneVerified graphql.Boolean
			Username      graphql.String
			Nickname      graphql.String
			// TODO: more fields from `ExtendUser`
		}
		TotalCount graphql.Int
	} `graphql:"users(registerInClient: $registerInClient, page: $page, count: $count)"`
}

// Users ...
func (c *Client) Users(parameter *UsersQueryParameter) (UsersQuery, error) {
	var q UsersQuery

	variables := map[string]interface{}{
		"registerInClient": parameter.RegisterInClient,
		"page":             parameter.Page,
		"count":            parameter.Count,
	}

	err := c.client.Query(context.Background(), &q, variables)

	if err != nil {
		return q, err
	}

	return q, nil
}

//------------------------------------------------------------------------------------

// RemoveUsersInput ...
type RemoveUsersInput struct {
	IDs              []graphql.String `json:"ids"`
	RegisterInClient graphql.String   `json:"registerInClient"`
	Operator         graphql.String   `json:"operator"` // FIXME: Mandatory
}

// RemoveUsersMutation ...
type RemoveUsersMutation struct {
	RemoveUsers []struct {
		ID      graphql.String `graphql:"_id"`
		Email   graphql.String
		Unionid graphql.String
	} `graphql:"removeUsers(ids: $ids, registerInClient: $registerInClient, operator: $operator)"`
}

// RemoveUsers ... TODO: need to tune the graphql error response json unmarshal
func (c *Client) RemoveUsers(input *RemoveUsersInput) (RemoveUsersMutation, error) {
	var m RemoveUsersMutation

	variables := map[string]interface{}{
		"ids":              input.IDs,
		"registerInClient": input.RegisterInClient,
		"operator":         input.Operator,
	}

	err := c.client.Mutate(context.Background(), &m, variables)

	if err != nil {
		return m, err
	}

	return m, nil
}

//------------------------------------------------------------------------------------

// UserUpdateInput ...
// TODO: no need all fields
type UserUpdateInput struct {
	Email         graphql.String  `json:"email,omitempty"`
	Unionid       graphql.String  `json:"unionid,omitempty"`
	EmailVerified graphql.Boolean `json:"emailVerified,omitempty"`
	Phone         graphql.String  `json:"phone,omitempty"`
	PhoneVerified graphql.Boolean `json:"phoneVerified,omitempty"`
	Username      graphql.String  `json:"username,omitempty"`
	Nickname      graphql.String  `json:"nickname,omitempty"`
	Company       graphql.String  `json:"company,omitempty"`
	Photo         graphql.String  `json:"photo,omitempty"`
	Browser       graphql.String  `json:"browser,omitempty"`
	// Password         graphql.String  `json:"password,omitempty"`
	RegisterInClient graphql.String `json:"registerInClient"`
	RegisterMethod   graphql.String `json:"registerMethod,omitempty"`
	// Oauth            graphql.String  `json:"oauth,omitempty"`
}

// UpdateUserMutation ...
// TODO: no need all fields
type UpdateUserMutation struct {
	UpdateUser struct {
		Email            graphql.String
		Unionid          graphql.String
		EmailVerified    graphql.Boolean
		Phone            graphql.String
		PhoneVerified    graphql.Boolean
		Username         graphql.String
		Nickname         graphql.String
		Company          graphql.String
		Photo            graphql.String
		Browser          graphql.String
		Password         graphql.String
		RegisterInClient graphql.String
		RegisterMethod   graphql.String
		Oauth            graphql.String
		Token            graphql.String
		TokenExpiredAt   graphql.String
		LoginsCount      graphql.Int
		LastLogin        graphql.String
		// LastIP           graphql.String //FIXME: it may cause `mutation` failed
		SignedUp  graphql.String
		Blocked   graphql.Boolean
		IsDeleted graphql.Boolean
		// OldPassword graphql.String
	} `graphql:"updateUser(options: $options)"`
}

// UpdateUser ...
func (c *Client) UpdateUser(input *UserUpdateInput) (UpdateUserMutation, error) {
	var m UpdateUserMutation

	variables := map[string]interface{}{
		"options": UserUpdateInput{
			Username:         input.Username,
			Nickname:         input.Nickname,
			Phone:            input.Phone,
			RegisterInClient: input.RegisterInClient,
		},
	}

	err := c.client.Mutate(context.Background(), &m, variables)

	if err != nil {
		return m, err
	}

	return m, nil
}

//------------------------------------------------------------------------------------

// SendVerifyEmailInput ...
type SendVerifyEmailInput struct {
	Email  graphql.String `json:"email"`
	Client graphql.String `json:"client"`
	Token  graphql.String `json:"token,omitempty"`
}

// SendVerifyEmailMutation ...
type SendVerifyEmailMutation struct {
	SendVerifyEmail struct {
		Message graphql.String
		Code    graphql.Int
		Status  graphql.Boolean
	} `graphql:"sendVerifyEmail(email: $email, client: $client, token: $token)"`
}

// SendVerifyEmail ...
func (c *Client) SendVerifyEmail(input *SendVerifyEmailInput) error {
	var m SendVerifyEmailMutation

	variables := map[string]interface{}{
		"email":  input.Email,
		"client": input.Client,
		"token":  input.Token,
	}

	err := c.client.Mutate(context.Background(), &m, variables)

	if err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------------------------------

// SendResetPasswordEmailInput ...
type SendResetPasswordEmailInput struct {
	Client graphql.String
	Email  graphql.String
}

// SendResetPasswordEmailMutation ...
type SendResetPasswordEmailMutation struct {
	SendResetPasswordEmail struct {
		Message graphql.String
		Code    graphql.Int
		Status  graphql.Boolean
	} `graphql:"sendResetPasswordEmail(client: $client, email: $email)"`
}

// SendResetPasswordEmail ...
func (c *Client) SendResetPasswordEmail(input *SendResetPasswordEmailInput) error {
	var m SendResetPasswordEmailMutation

	variables := map[string]interface{}{
		"email":  input.Email,
		"client": input.Client,
	}

	err := c.client.Mutate(context.Background(), &m, variables)

	if err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------------------------------

// VerifyResetPasswordVerifyCodeInput ...
type VerifyResetPasswordVerifyCodeInput struct {
	Client     graphql.String
	Email      graphql.String
	VerifyCode graphql.String
}

// VerifyResetPasswordVerifyCodeMutation ...
type VerifyResetPasswordVerifyCodeMutation struct {
	VerifyResetPasswordVerifyCode struct {
		Message graphql.String
		Code    graphql.Int
		Status  graphql.Boolean
	} `graphql:"verifyResetPasswordVerifyCode(client: $client, email: $email, verifyCode: $verifyCode)"`
}

// VerifyResetPasswordVerifyCode ...
func (c *Client) VerifyResetPasswordVerifyCode(input *VerifyResetPasswordVerifyCodeInput) error {
	var m VerifyResetPasswordVerifyCodeMutation

	variables := map[string]interface{}{
		"email":      input.Email,
		"client":     input.Client,
		"verifyCode": input.VerifyCode,
	}

	err := c.client.Mutate(context.Background(), &m, variables)

	if err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------------------------------

// ChangePasswordInput ...
type ChangePasswordInput struct {
	Email      graphql.String
	Client     graphql.String
	Password   graphql.String
	VerifyCode graphql.String
}

// ChangePasswordInputMutation ...
type ChangePasswordInputMutation struct {
	ChangePassword struct {
		Password graphql.String
	} `graphql:"changePassword(email: $email,client: $client,password: $password,verifyCode: $verifyCode)"`
}

// ChangePassword ...
func (c *Client) ChangePassword(input *ChangePasswordInput) error {
	var m ChangePasswordInputMutation

	password := graphql.String(encryptPassword([]byte(string(input.Password))))

	variables := map[string]interface{}{
		"email":      input.Email,
		"client":     input.Client,
		"verifyCode": input.VerifyCode,
		"password":   password,
	}

	err := c.client.Mutate(context.Background(), &m, variables)

	if err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------------------------------

// ReadOauthListQueryParameter ...
type ReadOauthListQueryParameter struct {
	ClientID   graphql.String
	DontGetURL graphql.Boolean
}

// ReadOauthListQuery ...
type ReadOauthListQuery struct {
	ReadOauthList []struct {
		Name        graphql.String
		Alias       graphql.String
		Image       graphql.String
		Description graphql.String
		Enabled     graphql.Boolean
		// Url         graphql.String
		Client graphql.String
		User   graphql.String
		// TODO more fields from `OAuthList`
	} `graphql:"ReadOauthList(clientId: $clientId, dontGetURL: $dontGetURL) "`
}

// ReadOauthList ...
func (c *Client) ReadOauthList(parameter *ReadOauthListQueryParameter) (ReadOauthListQuery, error) {
	var q ReadOauthListQuery

	variables := map[string]interface{}{
		"clientId":   parameter.ClientID,
		"dontGetURL": parameter.DontGetURL,
	}

	err := c.client.Query(context.Background(), &q, variables)

	if err != nil {
		return q, err
	}

	return q, nil
}

//------------------------------------------------------------------------------------
