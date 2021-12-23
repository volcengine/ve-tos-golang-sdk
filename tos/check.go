package tos

import "errors"

func IsValidBucketName(name string) error {
	if length := len(name); length < 3 || length > 63 {
		return errors.New("tos: bucket name length must between [3, 64)")
	}

	for i := range name {
		if char := name[i]; !(('a' <= char && char <= 'z') || ('0' <= char && char <= '9') || char == '-') {
			return errors.New("tos: bucket name can consist only of lowercase letters, numbers, and '-' ")
		}
	}

	if name[0] == '-' || name[len(name)-1] == '-' {
		return errors.New("tos: bucket name must begin and end with a letter or number")
	}

	return nil
}

func isValidNames(bucket string, key string, keys ...string) error {
	if err := IsValidBucketName(bucket); err != nil {
		return err
	}

	if err := isValidKey(key, keys...); err != nil {
		return err
	}

	return nil
}

func isValidKey(key string, keys ...string) error {
	if len(key) == 0 {
		return errors.New("tos: object name is empty")
	}

	for _, k := range keys {
		if len(k) == 0 {
			return errors.New("tos: object name is empty")
		}
	}

	return nil
}
