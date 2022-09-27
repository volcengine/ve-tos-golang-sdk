package tos

import (
	"unicode/utf8"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

// IsValidBucketName validate bucket name, return TosClientError if failed
func IsValidBucketName(name string) error {
	if length := len(name); length < 3 || length > 63 {
		return InvalidBucketNameLength
	}
	for i := range name {
		if char := name[i]; !(('a' <= char && char <= 'z') || ('0' <= char && char <= '9') || char == '-') {
			return InvalidBucketNameCharacter
		}
	}
	if name[0] == '-' || name[len(name)-1] == '-' {
		return InvalidBucketNameStartingOrEnding
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
		return InvalidObjectNameLength
	}
	if key[0] == '/' || key[0] == '\\' {
		return InvalidObjectNameStartingOrEnding
	}
	bytes := []byte(key)
	ok := utf8.Valid(bytes)
	if !ok {
		return InvalidObjectNameCharacterSet
	}
	for _, r := range []rune(key) {
		if (r >= 0 && r < 32) || (r > 127 && r < 256) {
			return InvalidObjectNameCharacterSet
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
		aclType == enum.ACLAuthRead || aclType == enum.ACLBucketOwnerRead ||
		aclType == enum.ACLBucketOwnerFullControl || aclType == enum.ACLLogDeliveryWrite {
		return nil
	}

	return InvalidACL
}

// isValidStorageClass validate Storage Class, return TosClientError if failed
func isValidStorageClass(storageClass enum.StorageClassType) error {

	if storageClass == enum.StorageClassIa || storageClass == enum.StorageClassStandard {
		return nil
	}

	return InvalidStorageClass
}

func isValidGrantee(granteeType enum.GranteeType) error {
	if granteeType == enum.GranteeUser || granteeType == enum.GranteeGroup {
		return nil
	}
	return InvalidGrantee
}

func isValidCannedType(cannedType enum.CannedType) error {
	if cannedType == enum.CannedAllUsers || cannedType == enum.CannedAuthenticatedUsers {
		return nil
	}
	return InvalidCanned
}

func isValidAzRedundancy(redundancyType enum.AzRedundancyType) error {
	if redundancyType == enum.AzRedundancySingleAz || redundancyType == enum.AzRedundancyMultiAz {
		return nil
	}
	return InvalidAzRedundancy
}

func isValidMetadataDirective(directiveType enum.MetadataDirectiveType) error {
	if directiveType == enum.MetadataDirectiveCopy || directiveType == enum.MetadataDirectiveReplace {
		return nil
	}
	return InvalidMetadataDirective
}

func isValidPermission(permissionType enum.PermissionType) error {
	if permissionType == enum.PermissionRead || permissionType == enum.PermissionReadAcp ||
		permissionType == enum.PermissionWriteAcp || permissionType == enum.PermissionWrite ||
		permissionType == enum.PermissionFullControl {
		return nil
	}
	return InvalidPermission
}

func isValidSSECAlgorithm(algorithm string) error {
	if algorithm == enum.SSETosAlg || algorithm == enum.SSEKMS {
		return nil
	}
	return InvalidSSECAlgorithm
}
