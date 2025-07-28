package tests

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

func sign(key []byte, value string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(value))
	return h.Sum(nil)
}

func getSigningKey(key []byte, dateStamp string, regionName, serviceName string) []byte {
	kDate := sign(key, dateStamp)
	kRegion := sign(kDate, regionName)
	kService := sign(kRegion, serviceName)
	kSigning := sign(kService, "request")
	return kSigning
}

type SignHeaderInput struct {
	method    string
	service   string
	host      string
	region    string
	params    url.Values
	accessKey string
	secretKey string
}

const (
	iso8601Layout = "20060102T150405Z"
	yyMMdd        = "20060102"
	emptySHA256   = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

func getSigningHeader(input SignHeaderInput) map[string]string {
	contentType := "application/x-www-form-urlencoded"
	accept := "application/json"
	t := time.Now().UTC()
	xdate := t.Format(iso8601Layout)
	datestamp := t.Format(yyMMdd)
	// *************  1: 拼接规范请求串*************
	canonicalUri := "/"
	canonicalQueryString := input.params.Encode()
	canonicalHeaders := "content-type:" + contentType + "\n" + "host:" + input.host + "\n" + "x-date:" + xdate + "\n"
	signedHeaders := "content-type;host;x-date"
	canonicalRequest := input.method + "\n" + canonicalUri + "\n" + canonicalQueryString + "\n" + canonicalHeaders + "\n" + signedHeaders + "\n" + emptySHA256
	// *************  2：拼接待签名字符串*************
	algorithm := "HMAC-SHA256"
	credentialScope := datestamp + "/" + input.region + "/" + input.service + "/" + "request"
	cr256 := sha256.Sum256([]byte(canonicalRequest))
	stringToSign := algorithm + "\n" + xdate + "\n" + credentialScope + "\n" + hex.EncodeToString(cr256[:])

	// *************  3：计算签名 *************
	signingKey := getSigningKey([]byte(input.secretKey), datestamp, input.region, input.service)

	signature := sign(signingKey, stringToSign)
	// *************4：添加签名到请求header中 * ************
	authorizationHeader := algorithm + " " + "Credential=" + input.accessKey + "/" + credentialScope + ", " + "SignedHeaders=" + signedHeaders + ", " + "Signature=" + hex.EncodeToString(signature)

	headers := map[string]string{"Accept": accept, "Content-Type": contentType, "X-Date": xdate, "Authorization": authorizationHeader}
	return headers
}

func assumeRole(host string, region, role string, accountId, accessKey, secretKey string) (*stsTokenResp, error) {
	service := "sts"
	queryParams := url.Values{"Action": []string{"AssumeRole"},
		"RoleSessionName": []string{"go_sdk"},
		"RoleTrn":         []string{fmt.Sprintf("trn:iam::%s:role/%s", accountId, role)},
		"Version":         []string{"2018-01-01"},
	}
	header := getSigningHeader(SignHeaderInput{
		method:    http.MethodGet,
		service:   service,
		host:      host,
		region:    region,
		params:    queryParams,
		accessKey: accessKey,
		secretKey: secretKey,
	})
	req, err := http.NewRequest(http.MethodGet, "https://"+host+"?"+queryParams.Encode(), nil)
	if err != nil {
		return nil, err
	}
	for key, value := range header {
		req.Header.Set(key, value)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err

	}
	fmt.Println(fmt.Sprintf("status code:%d, body: %s", resp.StatusCode, string(body)))
	roleResp := &stsTokenResp{}
	err = json.Unmarshal(body, roleResp)
	if err != nil {
		return nil, err
	}
	return roleResp, nil

}

type stsTokenResp struct {
	ResponseMetadata struct {
		RequestId string `json:"RequestId"`
		Action    string `json:"Action"`
		Version   string `json:"Version"`
		Service   string `json:"Service"`
		Region    string `json:"Region"`
	} `json:"ResponseMetadata"`
	Result struct {
		Credentials struct {
			ExpiredTime     time.Time `json:"ExpiredTime"`
			CurrentTime     time.Time `json:"CurrentTime"`
			AccessKeyId     string    `json:"AccessKeyId"`
			SecretAccessKey string    `json:"SecretAccessKey"`
			SessionToken    string    `json:"SessionToken"`
		} `json:"Credentials"`
		AssumedRoleUser struct {
			Trn           string `json:"Trn"`
			AssumedRoleId string `json:"AssumedRoleId"`
		} `json:"AssumedRoleUser"`
	} `json:"Result"`
}
