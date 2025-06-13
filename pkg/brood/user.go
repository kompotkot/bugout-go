package brood

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/bugout-dev/bugout-go/pkg/utils"
)

func getUserID(decoder *json.Decoder) (string, error) {
	userIDWrapper := struct {
		UserID string `json:"user_id"`
	}{}
	decodeErr := decoder.Decode(&userIDWrapper)
	return userIDWrapper.UserID, decodeErr
}

func (client BroodClient) Auth(token string) (AuthUser, error) {
	authRoute := client.Routes.Auth
	request, requestErr := http.NewRequest("GET", authRoute, nil)
	if requestErr != nil {
		return AuthUser{}, requestErr
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	request.Header.Add("Accept", "application/json")

	response, err := client.HTTPClient.Do(request)
	if err != nil {
		return AuthUser{}, err
	}
	defer response.Body.Close()

	var buf bytes.Buffer
	bodyReader := io.TeeReader(response.Body, &buf)

	statusErr := utils.HTTPStatusCheck(response)
	if statusErr != nil {
		return AuthUser{}, statusErr
	}

	var authUser AuthUser
	decodeErr := json.NewDecoder(bodyReader).Decode(&authUser)
	if decodeErr != nil {
		return authUser, decodeErr
	}

	return authUser, nil
}

func (client BroodClient) CreateUser(username, email, password string) (User, error) {
	userRoute := client.Routes.User
	data := url.Values{}
	data.Add("username", username)
	data.Add("email", email)
	data.Add("password", password)
	response, err := client.HTTPClient.PostForm(userRoute, data)
	if err != nil {
		return User{}, err
	}
	defer response.Body.Close()

	var buf bytes.Buffer
	bodyReader := io.TeeReader(response.Body, &buf)

	statusErr := utils.HTTPStatusCheck(response)
	if statusErr != nil {
		return User{}, statusErr
	}

	var user User
	decodeErr := json.NewDecoder(bodyReader).Decode(&user)
	if decodeErr != nil {
		return user, decodeErr
	}
	if user.Id == "" {
		userID, decodeErr := getUserID(json.NewDecoder(&buf))
		if decodeErr != nil {
			return user, decodeErr
		}
		user.Id = userID
	}

	return user, nil
}

func (client BroodClient) GenerateToken(username, password string) (string, error) {
	tokenRoute := client.Routes.Token
	data := url.Values{}
	data.Add("username", username)
	data.Add("password", password)

	response, err := client.HTTPClient.PostForm(tokenRoute, data)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	statusErr := utils.HTTPStatusCheck(response)
	if statusErr != nil {
		return "", statusErr
	}

	var token UserGeneratedToken
	decodeErr := json.NewDecoder(response.Body).Decode(&token)
	if decodeErr != nil {
		return token.Id, decodeErr
	}

	return token.Id, nil
}

func (client BroodClient) AnnotateToken(token, tokenType, note string) (string, error) {
	tokenRoute := client.Routes.Token
	data := url.Values{}
	data.Add("access_token", token)
	data.Add("token_type", tokenType)
	data.Add("token_note", note)
	encodedData := data.Encode()

	request, err := http.NewRequest("PUT", tokenRoute, strings.NewReader(encodedData))
	if err != nil {
		return "", err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(encodedData)))

	response, err := client.HTTPClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	statusErr := utils.HTTPStatusCheck(response)
	if statusErr != nil {
		return "", statusErr
	}

	tokenBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(tokenBytes), nil
}

func (client BroodClient) ListTokens(token string) (UserTokensList, error) {
	listTokensRoute := client.Routes.ListTokens
	request, requestErr := http.NewRequest("GET", listTokensRoute, nil)
	if requestErr != nil {
		return UserTokensList{}, requestErr
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	request.Header.Add("Accept", "application/json")

	response, err := client.HTTPClient.Do(request)
	if err != nil {
		return UserTokensList{}, err
	}
	defer response.Body.Close()

	statusErr := utils.HTTPStatusCheck(response)
	if statusErr != nil {
		return UserTokensList{}, statusErr
	}

	var result UserTokensList
	decodeErr := json.NewDecoder(response.Body).Decode(&result)
	return result, decodeErr
}

/*
	Find Brood user if exists

query parameters:
- **user_id** (UUID): Brood user ID
- **username** (string): User name
- **email** (string): User email
- **application_id** (UUID): Application user belongs to
*/
func (client BroodClient) FindUser(token string, queryParameters map[string]string) (User, error) {
	findUserRoute := client.Routes.FindUser
	request, requestErr := http.NewRequest("GET", findUserRoute, nil)
	if requestErr != nil {
		return User{}, requestErr
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	request.Header.Add("Accept", "application/json")

	query := request.URL.Query()
	for k, v := range queryParameters {
		query.Add(k, v)
	}
	request.URL.RawQuery = query.Encode()

	response, err := client.HTTPClient.Do(request)
	if err != nil {
		return User{}, err
	}
	defer response.Body.Close()

	var buf bytes.Buffer
	bodyReader := io.TeeReader(response.Body, &buf)

	statusErr := utils.HTTPStatusCheck(response)
	if statusErr != nil {
		return User{}, statusErr
	}

	var user User
	decodeErr := json.NewDecoder(bodyReader).Decode(&user)
	if decodeErr != nil {
		return user, decodeErr
	}
	if user.Id == "" {
		userID, decodeErr := getUserID(json.NewDecoder(&buf))
		if decodeErr != nil {
			return user, decodeErr
		}
		user.Id = userID
	}

	return user, nil
}

func (client BroodClient) GetUser(token string) (User, error) {
	userRoute := client.Routes.User
	request, requestErr := http.NewRequest("GET", userRoute, nil)
	if requestErr != nil {
		return User{}, requestErr
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	request.Header.Add("Accept", "application/json")

	response, err := client.HTTPClient.Do(request)
	if err != nil {
		return User{}, err
	}
	defer response.Body.Close()

	var buf bytes.Buffer
	bodyReader := io.TeeReader(response.Body, &buf)

	statusErr := utils.HTTPStatusCheck(response)
	if statusErr != nil {
		return User{}, statusErr
	}

	var user User
	decodeErr := json.NewDecoder(bodyReader).Decode(&user)
	if decodeErr != nil {
		return user, decodeErr
	}
	if user.Id == "" {
		userID, decodeErr := getUserID(json.NewDecoder(&buf))
		if decodeErr != nil {
			return user, decodeErr
		}
		user.Id = userID
	}

	return user, nil
}

func (client BroodClient) VerifyUser(token, code string) (User, error) {
	confirmRoute := client.Routes.ConfirmRegistration
	data := url.Values{}
	data.Add("verification_code", code)
	encodedData := data.Encode()

	request, requestErr := http.NewRequest("POST", confirmRoute, strings.NewReader(encodedData))
	if requestErr != nil {
		return User{}, requestErr
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(encodedData)))
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	request.Header.Add("Accept", "application/json")

	response, err := client.HTTPClient.Do(request)
	if err != nil {
		return User{}, err
	}
	defer response.Body.Close()

	var buf bytes.Buffer
	bodyReader := io.TeeReader(response.Body, &buf)

	statusErr := utils.HTTPStatusCheck(response)
	if statusErr != nil {
		return User{}, statusErr
	}

	var user User
	decodeErr := json.NewDecoder(bodyReader).Decode(&user)
	if decodeErr != nil {
		return user, decodeErr
	}
	if user.Id == "" {
		userID, decodeErr := getUserID(json.NewDecoder(&buf))
		if decodeErr != nil {
			return user, decodeErr
		}
		user.Id = userID
	}

	return user, nil
}

func (client BroodClient) ChangePassword(token, currentPassword, newPassword string) (User, error) {
	changePasswordRoute := client.Routes.ChangePassword
	data := url.Values{}
	data.Add("current_password", currentPassword)
	data.Add("new_password", newPassword)
	encodedData := data.Encode()

	request, requestErr := http.NewRequest("POST", changePasswordRoute, strings.NewReader(encodedData))
	if requestErr != nil {
		return User{}, requestErr
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(encodedData)))
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	request.Header.Add("Accept", "application/json")

	response, err := client.HTTPClient.Do(request)
	if err != nil {
		return User{}, err
	}
	defer response.Body.Close()

	var buf bytes.Buffer
	bodyReader := io.TeeReader(response.Body, &buf)

	statusErr := utils.HTTPStatusCheck(response)
	if statusErr != nil {
		return User{}, statusErr
	}

	var user User
	decodeErr := json.NewDecoder(bodyReader).Decode(&user)
	if decodeErr != nil {
		return user, decodeErr
	}
	if user.Id == "" {
		userID, decodeErr := getUserID(json.NewDecoder(&buf))
		if decodeErr != nil {
			return user, decodeErr
		}
		user.Id = userID
	}

	return user, nil
}
