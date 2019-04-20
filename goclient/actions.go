package main

import (
	"encoding/base64"
	"fmt"
)

func (c *CookiejarClient) count() (string, error) {
	res, err := c.sendRequest(fmt.Sprintf("state/%s", c.getAddress()), "", nil)
	if err != nil {
		return "", err
	}

	responseMap, err := c.getResponseMap(res)
	if err != nil {
		return "", fmt.Errorf("Error reading response: %v", err)
	}

	data, ok := responseMap["data"].(string)
	if !ok {
		return "", fmt.Errorf("assertion to string failed")
	}

	resData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("Decoding error: %v", err)
	}

	return string(resData), nil
}

func (c *CookiejarClient) clear() error {
	if _, err := c.wrapAndSend("clear", 0, 10); err != nil {
		return err
	}

	return nil
}

func (c *CookiejarClient) bake(amount int) (string, error) {
	return c.wrapAndSend("bake", amount, 10)
}

func (c *CookiejarClient) eat(amount int) (string, error) {
	return c.wrapAndSend("eat", amount, 10)
}
