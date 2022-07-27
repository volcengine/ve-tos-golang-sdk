package tos

import (
	"unicode/utf8"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

// IsValidBucketName validate bucket name, return TosClientError if failed
func IsValidBucketName(name string) error {
	if length := len(name); length < 3 || length > 63 {
		return newTosClientError("tos: invalid bucket name, the length must be [3, 63]", nil)
	}
	for i := range name {
		if char := name[i]; !(('a' <= char && char <= 'z') || ('0' <= char && char <= '9') || char == '-') {
			return newTosClientError("tos: bucket name can consist only of lowercase letters, numbers, and '-' ", nil)
		}
	}
	if name[0] == '-' || name[len(name)-1] == '-' {
		return newTosClientError("tos: invalid bucket name, the bucket name can be neither starting with '-' nor ending with '-'", nil)
	}
	return nil
}

// isValidNames validate bucket name and keys, return TosClientError if failed
func isValidNames(bucket string, key string, keys ...string) error {
	if err := IsValidBucketName(bucket); err != nil {
		return err
	}
	if err := isValidKey(key, keys...); err != nil {
		return err
	}
	return nil
}

// validKey validate single key, return TosClientError if failed
func validKey(key string) error {
	if len(key) < 1 || len(key) > 696 {
		return newTosClientError("tos: invalid object name, the length must be [1, 696]", nil)
	}
	if key[0] == '/' || key[0] == '\\' {
		return newTosClientError("tos: invalid object name, the object name can not start with '/' or '\\' ", nil)
	}
	bytes := []byte(key)
	ok := utf8.Valid(bytes)
	if !ok {
		return newTosClientError("tos: invalid object name, the character set is illegal", nil)
	}
	for _, r := range []rune(key) {
		if (r >= 0 && r < 32) || (r > 127 && r < 256) {
			return newTosClientError("tos: object key is not allowed to contain invisible characters except space", nil)
		}
	}
	return nil
}

// isValidKey validate keys, return TosClientError if failed
func isValidKey(key string, keys ...string) error {
	if err := validKey(key); err != nil {
		return err
	}
	for _, k := range keys {
		if err := validKey(k); err != nil {
			return err
		}
	}
	return nil
}

// isValidACL validate aclType, return TosClientError if failed
func isValidACL(aclType enum.ACLType) error {
	if aclType == enum.ACLPrivate || aclType == enum.ACLPublicRead || aclType == enum.ACLPublicReadWrite ||
		aclType == enum.ACLAuthRead || aclType == enum.ACLBucketOwnerRead || aclType == enum.ACLBucketOwnerFullControl {
		return nil
	}

	return newTosClientError("tos: invalid ACL", nil)
}
